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

	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/redact"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// subscribePoll is how often session.subscribe reads the journal and the
// session's state. The journal is append-only, so polling is enough.
const subscribePoll = 200 * time.Millisecond

// sessionSend delivers text to a session as a message ticket, so queueing
// and interrupts are the #16 delivery. By pool and key it starts the
// session when a chat pool has none open for the key.
func (c *Controller) sessionSend(ctx context.Context, sc scope, p rpc.SessionSendParams) (any, error) {
	if err := requireAdmin(sc, "session send"); err != nil {
		return nil, err
	}
	// An attach send with no text only checks that rw attach is allowed.
	probe := p.Attach && p.Text == "" && p.Context == ""
	if !probe && strings.TrimSpace(p.Text) == "" && strings.TrimSpace(p.Context) == "" {
		return nil, fmt.Errorf("%w: text is required", store.ErrInvalid)
	}
	var (
		sess    store.Ticket
		started bool
		err     error
	)
	switch {
	case p.ID != "" && (p.Pool != "" || p.Key != ""):
		return nil, fmt.Errorf("%w: give id or pool and key, not both", store.ErrInvalid)
	case p.ID != "":
		if sess, err = c.openSessionTicket(ctx, p.ID); err != nil {
			return nil, err
		}
		if p.Attach {
			if pl, ok := c.config().Pools[sess.Pool]; !ok || pl.Attach != config.AttachRW {
				return nil, fmt.Errorf("%w: pool %q of session %s does not allow rw attach (set attach: rw)", store.ErrInvalid, sess.Pool, sess.ID)
			}
		}
	case p.Attach:
		return nil, fmt.Errorf("%w: rw attach needs a session id", store.ErrInvalid)
	case p.Pool != "" && p.Key != "":
		if sess, started, err = c.keyedSession(ctx, sc, p.Pool, p.Key); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("%w: give a session id, or pool and key", store.ErrInvalid)
	}
	if probe {
		return rpc.SessionSendResult{Session: sess.ID}, nil
	}
	// A human writing lifts the read-only first turn (#606).
	c.humanWrote(ctx, sess.ID)
	c.noteOperator(ctx, sess, sc.actor)
	c.resumeSession(ctx, sess.ID)
	fields, err := json.Marshal(store.MessageFields{To: sess.ID, From: sc.actor, Interrupt: p.Interrupt, Context: strings.TrimSpace(p.Context)})
	if err != nil {
		return nil, err
	}
	title, body, _ := strings.Cut(p.Text, "\n")
	m, err := store.Create(ctx, c.store, sc.actor, store.NewTicket{
		Project: sess.Project, Kind: store.KindMessage, Title: title, Body: body, Fields: fields,
	})
	if err != nil {
		return nil, err
	}
	// Wake an idle runner now rather than at its next poll (#603).
	c.wakeRunner(sess.ID)
	return rpc.SessionSendResult{Session: sess.ID, Message: m, Started: started}, nil
}

func (c *Controller) openSessionTicket(ctx context.Context, id string) (store.Ticket, error) {
	t, err := c.store.Get(ctx, id)
	if err != nil {
		return t, err
	}
	if t.Kind != store.KindSession {
		return t, fmt.Errorf("%w: %s is a %s, not a session", store.ErrInvalid, id, t.Kind)
	}
	if t.Status == store.StatusClosed {
		return t, fmt.Errorf("%w: session %s", store.ErrClosed, id)
	}
	return t, nil
}

// keyedSession finds the open session for key in pool, or starts one when
// pool is a chat pool. Ticket pools get sessions from the supervisor only.
func (c *Controller) keyedSession(ctx context.Context, sc scope, pool, key string) (store.Ticket, bool, error) {
	find := func() (store.Ticket, bool, error) {
		open, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Status: store.StatusOpen})
		if err != nil {
			return store.Ticket{}, false, err
		}
		for _, t := range open {
			f, err := store.DecodeFields[store.SessionFields](t)
			if err != nil || f.Key != key {
				continue
			}
			if t.Pool != pool {
				return t, false, fmt.Errorf("%w: key %q belongs to session %s in pool %q", store.ErrConflict, key, t.ID, t.Pool)
			}
			return t, true, nil
		}
		return store.Ticket{}, false, nil
	}
	if t, ok, err := find(); err != nil || ok {
		return t, false, err
	}
	p, ok := c.config().Pools[pool]
	if !ok {
		return store.Ticket{}, false, fmt.Errorf("%w: unknown pool %q", store.ErrInvalid, pool)
	}
	if !p.Chat() {
		return store.Ticket{}, false, fmt.Errorf("%w: pool %s is not a chat pool; send to a session id", store.ErrInvalid, pool)
	}
	if err := c.humanSlot(ctx, pool); err != nil {
		return store.Ticket{}, false, err
	}
	t, err := c.newSession(ctx, sc.actor, p.Project, pool, p.Agent, key, pool+" "+key, "")
	if errors.Is(err, store.ErrAlreadyExist) {
		// A concurrent send created it; the key index allows one open.
		if t, ok, ferr := find(); ferr == nil && ok {
			return t, false, nil
		}
	}
	if err != nil {
		return store.Ticket{}, false, err
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
		return nil, fmt.Errorf("%w: session id is required", store.ErrInvalid)
	}
	t, err := c.store.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if t.Kind != store.KindSession {
		return nil, fmt.Errorf("%w: %s is a %s, not a session", store.ErrInvalid, p.ID, t.Kind)
	}
	if sc.peer == nil || sc.peer.conn == nil {
		return nil, fmt.Errorf("%w: subscribe needs a socket connection", store.ErrInvalid)
	}
	defer c.watch(p.ID)()
	// Events reach Desk, chat gateways and A2A peers (#574): redact with
	// the session's env and the configured rules, built once per subscription.
	red := c.redactor(journalEnv(c.root.SessionDir(p.ID)))
	send := func(ev rpc.SessionEvent) error {
		ev.Session = p.ID
		redactEvent(red, &ev)
		return sc.peer.conn.SendNotification(ctx, rpc.NotifySessionEvent, ev)
	}
	state := func(t store.Ticket) error {
		return send(rpc.SessionEvent{Type: rpc.EventState, Status: t.Status, Resolution: t.Resolution})
	}
	if err := state(t); err != nil {
		return nil, err
	}
	if p.SinceSeq == 0 && p.Tail > 0 {
		// Seq is 1-based and gapless, so the last Tail frames follow
		// last-Tail (#614).
		if last, _ := host.JournalTail(c.root.SessionDir(p.ID)); last > 0 {
			p.SinceSeq = max(last-p.Tail, 0)
		}
	}
	path := filepath.Join(c.root.SessionDir(p.ID), host.JournalFile)
	env := c.redactor(journalEnv(c.root.SessionDir(p.ID)))
	ev := rpc.EventDecoder{Redact: env.Text}
	var off int64
	res := rpc.SessionSubscribeResult{Seq: p.SinceSeq}
	last := t
	for {
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
		if last.Status == store.StatusClosed {
			// Frames read after the close was seen are the last ones.
			res.Closed, res.Resolution = true, last.Resolution
			return res, nil
		}
		select {
		case <-ctx.Done():
			return res, nil
		case <-c.after(subscribePoll):
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
	}
}

// humanSlot checks that a human-initiated session (a chat send by key, a
// manual session.new in a pool) may start in pool. The supervisor fills a
// pool up to max_sessions; humans get one slot over that, so queued work
// never blocks a human from reaching a session (#190). Pools with
// max_sessions 0 are manual-only and stay uncapped.
func (c *Controller) humanSlot(ctx context.Context, pool string) error {
	p, ok := c.config().Pools[pool]
	if !ok || p.MaxSessions == 0 {
		return nil
	}
	open, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Status: store.StatusOpen, Pool: pool})
	if err != nil {
		return err
	}
	if len(open) >= p.MaxSessions+1 {
		return fmt.Errorf("%w: pool %s has %d open sessions, max_sessions %d plus one reserved human slot", store.ErrConflict, pool, len(open), p.MaxSessions)
	}
	return nil
}
