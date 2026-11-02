// Package gateway connects chat networks to keyed chat-pool sessions. It
// is an admin-socket client: a message in an allowed thread becomes a
// session.send by pool and key, and the session's turn output, read with
// session.subscribe, is posted back to the thread.
package gateway

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Thread is where a message came from and where replies go.
type Thread struct {
	Room string // room or channel id
	Root string // thread root id
}

// Message is one received chat message.
type Message struct {
	Thread Thread
	ID     string // the message's own id, for reactions
	Sender string
	Text   string
}

// Adapter is one chat network. Keep it small so another network (Discord)
// fits behind it.
type Adapter interface {
	// Network names the network in session keys, e.g. "matrix".
	Network() string
	// Run receives messages until ctx ends, calling fn for each in order.
	Run(ctx context.Context, fn func(context.Context, Message)) error
	// Post sends text to a thread.
	Post(ctx context.Context, t Thread, text string) error
	// React marks a received message, to show progress.
	React(ctx context.Context, t Thread, msgID, key string) error
	// Open posts text as a new thread root in room and returns the thread.
	Open(ctx context.Context, room, text string) (Thread, error)
}

// Controller is the part of the admin API the gateway uses.
type Controller interface {
	Send(ctx context.Context, p rpc.SessionSendParams) (rpc.SessionSendResult, error)
	Subscribe(ctx context.Context, p rpc.SessionSubscribeParams, fn func(rpc.SessionEvent)) (rpc.SessionSubscribeResult, error)
	// List, Show and Update read and mark escalation tickets.
	List(ctx context.Context, p rpc.ListParams) ([]store.Ticket, error)
	Show(ctx context.Context, id string) (rpc.ShowResult, error)
	Update(ctx context.Context, p rpc.UpdateParams) (store.Ticket, error)
	// Events follows events.subscribe; escalation threads are driven by it.
	Events(ctx context.Context, p rpc.EventsSubscribeParams, fn func(rpc.Event)) (rpc.EventsSubscribeResult, error)
}

// Config is one gateway's routing.
type Config struct {
	Name    string // gateway name, recorded on escalation threads it opens
	Pool    string
	Rooms   []string // allowed rooms; empty allows none
	Senders []string // allowed sender ids; empty allows none
	// Marker prefixes a message to send it with interrupt. Default "!".
	Marker string
	// EscalationRoom gets one new thread per escalation ticket. Empty:
	// escalations are not posted.
	EscalationRoom string
}

// Reaction keys for progress.
const (
	ReactQueued = "👀"
)

// Gateway routes messages between one adapter and the controller.
type Gateway struct {
	cfg   Config
	ad    Adapter
	ctl   Controller
	log   *slog.Logger
	retry time.Duration

	mu         sync.Mutex
	watched    map[string]*watch // by session id
	escalating map[string]Thread // open escalation tickets with a thread, by ticket id
	wg         sync.WaitGroup
}

type watch struct {
	mu     sync.Mutex
	thread Thread
}

// New returns a gateway. log may be nil.
func New(cfg Config, ad Adapter, ctl Controller, log *slog.Logger) *Gateway {
	if cfg.Marker == "" {
		cfg.Marker = "!"
	}
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Gateway{
		cfg: cfg, ad: ad, ctl: ctl, log: log, retry: 2 * time.Second,
		watched: map[string]*watch{}, escalating: map[string]Thread{},
	}
}

// Key is the session key for a thread.
func (g *Gateway) Key(t Thread) string {
	return g.ad.Network() + ":" + t.Room + ":" + t.Root
}

// Run receives until ctx ends, then waits for the session watchers.
func (g *Gateway) Run(ctx context.Context) error {
	if g.cfg.EscalationRoom != "" {
		g.wg.Go(func() { g.runEscalations(ctx) })
	}
	err := g.ad.Run(ctx, g.Handle)
	g.wg.Wait()
	if errors.Is(err, context.Canceled) && ctx.Err() != nil {
		return nil
	}
	return err
}

// Allowed reports whether a message passes the room and sender allowlists.
func (g *Gateway) Allowed(m Message) bool {
	return slices.Contains(g.cfg.Rooms, m.Thread.Room) && slices.Contains(g.cfg.Senders, m.Sender)
}

// Handle routes one message. Messages from others are ignored.
func (g *Gateway) Handle(ctx context.Context, m Message) {
	if !g.Allowed(m) {
		g.log.Debug("gateway: ignored", "room", m.Thread.Room, "sender", m.Sender)
		return
	}
	text, interrupt := strings.CutPrefix(strings.TrimSpace(m.Text), g.cfg.Marker)
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	key := g.Key(m.Thread)
	sent := time.Now()
	p := rpc.SessionSendParams{Pool: g.cfg.Pool, Key: key, Text: text, Interrupt: interrupt}
	res, err := g.ctl.Send(ctx, p)
	if err != nil {
		g.log.Warn("gateway: send", "key", key, "err", err)
		g.post(ctx, m.Thread, "jobshop: "+err.Error())
		return
	}
	g.watch(ctx, res.Session, sent, m.Thread)
	if err := g.ad.React(ctx, m.Thread, m.ID, ReactQueued); err != nil {
		g.log.Debug("gateway: react", "err", err)
	}
}

// watch makes sure one goroutine posts the session's turns to t.
func (g *Gateway) watch(ctx context.Context, session string, sent time.Time, t Thread) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if w, ok := g.watched[session]; ok {
		w.mu.Lock()
		w.thread = t
		w.mu.Unlock()
		return
	}
	w := &watch{thread: t}
	g.watched[session] = w
	// Turns that started before this send (a session the gateway sees for
	// the first time, e.g. after a gateway restart) are not reposted.
	after := sent.Add(-time.Second)
	g.wg.Go(func() { g.follow(ctx, session, after, w) })
}

// follow posts each turn's text to the watch's thread until the session
// closes or ctx ends, resubscribing from the last seq on errors.
func (g *Gateway) follow(ctx context.Context, session string, after time.Time, w *watch) {
	defer func() {
		g.mu.Lock()
		delete(g.watched, session)
		g.mu.Unlock()
	}()
	var (
		buf   strings.Builder
		since int64
		skip  bool // the current turn predates the watch
		// broken is set by a tool call: text chunks stream one message, so
		// only a tool call between them starts a new one (#164).
		broken bool
	)
	thread := func() Thread { w.mu.Lock(); defer w.mu.Unlock(); return w.thread }
	for ctx.Err() == nil {
		res, err := g.ctl.Subscribe(ctx, rpc.SessionSubscribeParams{ID: session, SinceSeq: since}, func(ev rpc.SessionEvent) {
			switch ev.Type {
			case rpc.EventTurnStart:
				buf.Reset()
				broken = false
				skip = ev.Time != nil && ev.Time.Before(after)
			case rpc.EventTool:
				broken = true
			case rpc.EventText:
				if broken && strings.TrimSpace(buf.String()) != "" {
					buf.WriteString("\n\n")
					ev.Text = strings.TrimLeft(ev.Text, " \t\n")
				}
				broken = false
				buf.WriteString(ev.Text)
			case rpc.EventTurnEnd:
				out := strings.TrimSpace(buf.String())
				buf.Reset()
				if skip {
					skip = false
					return
				}
				if ev.Error != "" {
					out = strings.TrimSpace(out + "\n\n(turn failed: " + ev.Error + ")")
				}
				if out != "" {
					g.post(ctx, thread(), out)
				}
			}
		})
		since = max(since, res.Seq)
		if err == nil && res.Closed {
			return
		}
		if err != nil && ctx.Err() == nil {
			g.log.Warn("gateway: subscribe", "session", session, "err", err)
		}
		select {
		case <-ctx.Done():
		case <-time.After(g.retry):
		}
	}
}

func (g *Gateway) post(ctx context.Context, t Thread, text string) {
	if err := g.ad.Post(ctx, t, text); err != nil {
		g.log.Warn("gateway: post", "room", t.Room, "err", fmt.Sprint(err))
	}
}
