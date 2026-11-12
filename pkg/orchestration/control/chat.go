package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	host "github.com/gfx-labs/acphost/pkg/acphost"
	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/conversation"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/support/redact"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"

	"github.com/gfx-labs/jobshop/pkg/domain"
)

// subscribePoll is the fallback when a journal directory cannot be watched yet.
const subscribePoll = 200 * time.Millisecond

// sessionSend delivers text to a session as a message ticket, so queueing
// and interrupts are the #16 delivery. By pool and key it starts the
// session when a chat pool has none open for the key.
func (c *Controller) sessionSend(ctx context.Context, sc scope, p rpc.SessionSendParams) (any, error) {
	if err := requireAdmin(sc, "session send"); err != nil {
		return nil, err
	}
	if err := conversation.ValidateAttachments(p.Attachments); err != nil {
		return nil, err
	}
	// An attach send with no text only checks that rw attach is allowed.
	probe := p.Attach && p.Text == "" && p.Context == "" && len(p.Attachments) == 0
	if !probe && strings.TrimSpace(p.Text) == "" && strings.TrimSpace(p.Context) == "" && len(p.Attachments) == 0 {
		return nil, fmt.Errorf("%w: text is required", domain.ErrInvalid)
	}
	var (
		sess    domain.Ticket
		started bool
		err     error
	)
	switch {
	case p.ID != "" && (p.Pool != "" || p.Key != ""):
		return nil, fmt.Errorf("%w: give id or pool and key, not both", domain.ErrInvalid)
	case p.ID != "":
		if sess, err = c.openSessionTicket(ctx, p.ID); err != nil {
			return nil, err
		}
		if p.Interrupt {
			if err := c.checkInterrupt(ctx, "pool:"+sess.Pool); err != nil {
				return nil, err
			}
		}
		if p.Attach {
			if pl, ok := c.config().Pools[sess.Pool]; !ok || pl.Attach != config.AttachRW {
				return nil, fmt.Errorf("%w: pool %q of session %s does not allow rw attach (set attach: rw)", domain.ErrInvalid, sess.Pool, sess.ID)
			}
		}
	case p.Attach:
		return nil, fmt.Errorf("%w: rw attach needs a session id", domain.ErrInvalid)
	case p.Pool != "" && p.Key != "":
		// Checked before keyedSession, so a refused send starts no session.
		if p.Interrupt {
			if err := c.checkInterrupt(ctx, "pool:"+p.Pool); err != nil {
				return nil, err
			}
		}
		if sess, started, err = c.keyedSession(ctx, c.config(), sc, p.Pool, p.Key); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("%w: give a session id, or pool and key", domain.ErrInvalid)
	}
	if probe {
		return rpc.SessionSendResult{Session: sess.ID}, nil
	}
	// A human writing lifts the read-only first turn (#606).
	c.humanWrote(ctx, sess.ID)
	c.noteOperator(ctx, sess, sc.actor)
	c.resumeSession(ctx, sess.ID)
	id, err := c.store.NewID(sess.Project)
	if err != nil {
		return nil, err
	}
	attachments, err := c.saveAttachments(sc, sess.ID, id, p.Attachments)
	if err != nil {
		return nil, err
	}
	stored := false
	defer func() {
		if !stored && len(attachments) > 0 {
			conversation.RemoveAttachments(c.attachmentDir(sess.ID), id)
		}
	}()
	fields, err := json.Marshal(domain.MessageFields{To: sess.ID, From: sc.actor, Interrupt: p.Interrupt, Context: strings.TrimSpace(p.Context), Attachments: attachments})
	if err != nil {
		return nil, err
	}
	title, body, _ := strings.Cut(p.Text, "\n")
	m, err := store.Create(ctx, c.store, sc.actor, store.NewTicket{
		ID: id, Project: sess.Project, Kind: domain.KindMessage, Title: title, Body: body, Fields: fields,
	})
	if err != nil {
		return nil, err
	}
	stored = true
	// Wake an idle runner now rather than at its next poll (#603).
	c.wakeRunner(sess.ID)
	return rpc.SessionSendResult{Session: sess.ID, Message: m, Started: started}, nil
}

func (c *Controller) openSessionTicket(ctx context.Context, id string) (domain.Ticket, error) {
	t, err := c.store.Get(ctx, id)
	if err != nil {
		return t, err
	}
	if t.Kind != domain.KindSession {
		return t, fmt.Errorf("%w: %s is a %s, not a session", domain.ErrInvalid, id, t.Kind)
	}
	if t.Status == domain.StatusClosed {
		return t, fmt.Errorf("%w: session %s", domain.ErrClosed, id)
	}
	return t, nil
}

// keyedSession finds the open session for key in pool, or starts one when
// pool is a chat pool. Ticket pools get sessions from the supervisor only.
func (c *Controller) keyedSession(ctx context.Context, cfg *config.Config, sc scope, pool, key string) (domain.Ticket, bool, error) {
	find := func() (domain.Ticket, bool, error) {
		open, err := c.store.List(ctx, store.Filter{Kind: domain.KindSession, Status: domain.StatusOpen})
		if err != nil {
			return domain.Ticket{}, false, err
		}
		for _, t := range open {
			f, err := domain.DecodeFields[domain.SessionFields](t)
			if err != nil || f.Key != key {
				continue
			}
			if t.Pool != pool {
				return t, false, fmt.Errorf("%w: key %q belongs to session %s in pool %q", domain.ErrConflict, key, t.ID, t.Pool)
			}
			return t, true, nil
		}
		return domain.Ticket{}, false, nil
	}
	if t, ok, err := find(); err != nil || ok {
		return t, false, err
	}
	p, ok := cfg.Pools[pool]
	if !ok {
		return domain.Ticket{}, false, fmt.Errorf("%w: unknown pool %q", domain.ErrInvalid, pool)
	}
	if !p.Chat() {
		return domain.Ticket{}, false, fmt.Errorf("%w: pool %s is not a chat pool; send to a session id", domain.ErrInvalid, pool)
	}
	if err := c.humanSlot(ctx, cfg, pool); err != nil {
		return domain.Ticket{}, false, err
	}
	t, err := c.newSession(ctx, cfg, sc.actor, p.Project, pool, p.Agent, key, pool+" "+key, "")
	if errors.Is(err, domain.ErrAlreadyExist) {
		// A concurrent send created it; the key index allows one open.
		if t, ok, ferr := find(); ferr == nil && ok {
			return t, false, nil
		}
	}
	if err != nil {
		return domain.Ticket{}, false, err
	}
	return t, true, nil
}

// sessionSubscribe streams a session's journal as session.event
// notifications on the caller's connection, from SinceSeq until the
// session closes or the caller goes away. State events carry no seq, so
// a client resumes from the last seq it saw without gaps or repeats.
func (c *Controller) sessionSubscribe(ctx context.Context, sc scope, p rpc.SessionSubscribeParams) (any, error) {
	if err := requireAdmin(sc, "session subscribe"); err != nil {
		return nil, err
	}
	if p.ID == "" {
		return nil, fmt.Errorf("%w: session id is required", domain.ErrInvalid)
	}
	var changes <-chan rpc.Event
	var lagged <-chan struct{}
	if c.events != nil {
		sub, _ := c.events.add()
		defer c.events.remove(sub)
		changes, lagged = sub.ch, sub.lagged
	}
	t, err := c.store.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if t.Kind != domain.KindSession {
		return nil, fmt.Errorf("%w: %s is a %s, not a session", domain.ErrInvalid, p.ID, t.Kind)
	}
	if sc.peer == nil || sc.peer.conn == nil {
		return nil, fmt.Errorf("%w: subscribe needs a socket connection", domain.ErrInvalid)
	}
	defer c.watch(p.ID)()
	// Events reach Desk, chat gateways and A2A peers (#574): redact with
	// the session's env and the configured rules, built once per subscription.
	red := c.redactor(sessionEnv(c.root.SessionDir(p.ID)))
	send := func(ev rpc.SessionEvent) error {
		ev.Session = p.ID
		redactEvent(red, &ev)
		return sc.peer.conn.SendNotification(ctx, rpc.NotifySessionEvent, ev)
	}
	if p.SinceSeq == 0 && p.Tail > 0 {
		// Seq is 1-based and gapless, so the last Tail frames follow
		// last-Tail (#614).
		if last, _ := host.JournalTail(c.root.SessionDir(p.ID)); last > 0 {
			p.SinceSeq = max(last-p.Tail, 0)
		}
	}
	state := func(t domain.Ticket) error {
		return send(rpc.SessionEvent{Type: rpc.EventState, Status: t.Status, Resolution: t.Resolution, First: p.SinceSeq + 1})
	}
	if err := state(t); err != nil {
		return nil, err
	}
	path := filepath.Join(c.root.SessionDir(p.ID), host.JournalFile)
	var watcher *fsnotify.Watcher
	var writes <-chan fsnotify.Event
	var watchErrors <-chan error
	defer func() {
		if watcher != nil {
			_ = watcher.Close()
		}
	}()
	env := c.redactor(sessionEnv(c.root.SessionDir(p.ID)))
	ev := rpc.EventDecoder{Redact: env.Text}
	var off int64
	res := rpc.SessionSubscribeResult{Seq: p.SinceSeq}
	last := t
	for {
		if watcher == nil {
			// Register before reading, so an append cannot fall in a watch gap.
			w, err := fsnotify.NewWatcher()
			if err == nil {
				if err = w.Add(filepath.Dir(path)); err == nil {
					watcher, writes, watchErrors = w, w.Events, w.Errors
				} else {
					_ = w.Close()
				}
			}
		}
		frames, next, err := host.ReadFrom(path, off)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		off = next
		for _, fr := range frames {
			// Tool titles come from earlier frames, so all are decoded;
			// only those after SinceSeq are sent.
			out, ok := ev.Decode(fr.Seq, fr.T, fr.From, fr.Msg)
			if !ok || fr.Seq <= p.SinceSeq {
				continue
			}
			if err := send(out); err != nil {
				return nil, err
			}
			res.Seq = fr.Seq
		}
		if last.Status == domain.StatusClosed {
			// Frames read after the close was seen are the last ones.
			res.Closed, res.Resolution = true, last.Resolution
			return res, nil
		}
		var fallback <-chan time.Time
		if watcher == nil || changes == nil {
			fallback = c.after(subscribePoll)
		}
	awaitChange:
		for {
			select {
			case <-ctx.Done():
				return res, nil
			case <-fallback:
				break awaitChange
			case ev, ok := <-writes:
				if !ok || (ev.Name == filepath.Dir(path) && ev.Has(fsnotify.Remove|fsnotify.Rename)) {
					return nil, fmt.Errorf("session journal watch ended; reconnect to resume")
				}
				if ev.Name == path {
					break awaitChange
				}
			case err, ok := <-watchErrors:
				if !ok {
					return nil, fmt.Errorf("session journal watch closed; reconnect to resume")
				}
				return nil, fmt.Errorf("watch session journal: %w", err)
			case ev := <-changes:
				if ev.ID == p.ID {
					break awaitChange
				}
			case <-lagged:
				return nil, fmt.Errorf("session state subscription lagged; reconnect to resume")
			}
		}
		t, err := c.store.Get(ctx, p.ID)
		if err != nil {
			if ctx.Err() != nil {
				return res, nil
			}
			return nil, err
		}
		if t.Status != last.Status || t.Resolution != last.Resolution {
			if err := state(t); err != nil {
				return nil, err
			}
		}
		last = t
	}
}

// redactEvent masks secrets in the text fields of a session event.
func redactEvent(red *redact.Redactor, ev *rpc.SessionEvent) {
	ev.Text = red.Text(ev.Text)
	ev.Context = red.Text(ev.Context)
	ev.Error = red.Text(ev.Error)
	if ev.Tool != nil {
		ev.Tool.Title = red.Text(ev.Tool.Title)
		ev.Tool.Target = red.Text(ev.Tool.Target)
		ev.Tool.Intent = red.Text(ev.Tool.Intent)
	}
}

// humanSlot checks that a human-initiated session (a chat send by key, a
// manual session.new in a pool) may start in pool. The supervisor fills a
// pool up to max_sessions; humans get one slot over that, so queued work
// never blocks a human from reaching a session (#190). Pools with
// max_sessions 0 are manual-only and stay uncapped.
func (c *Controller) humanSlot(ctx context.Context, cfg *config.Config, pool string) error {
	p, ok := cfg.Pools[pool]
	if !ok || p.MaxSessions == 0 {
		return nil
	}
	open, err := c.store.List(ctx, store.Filter{Kind: domain.KindSession, Status: domain.StatusOpen, Pool: pool})
	if err != nil {
		return err
	}
	// A suspended session has no host and takes no slot (JOB-189).
	if n := countedSessions(open); n >= p.MaxSessions+1 {
		return fmt.Errorf("%w: pool %s has %d open sessions, max_sessions %d plus one reserved human slot", domain.ErrConflict, pool, n, p.MaxSessions)
	}
	return nil
}
