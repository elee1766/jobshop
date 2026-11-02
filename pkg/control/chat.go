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

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/host"
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
	probe := p.Attach && p.Text == ""
	if !probe && strings.TrimSpace(p.Text) == "" {
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
	fields, err := json.Marshal(store.MessageFields{To: sess.ID, From: sc.actor, Interrupt: p.Interrupt})
	if err != nil {
		return nil, err
	}
	title, body, _ := strings.Cut(p.Text, "\n")
	m, err := c.store.Create(ctx, sc.actor, store.NewTicket{
		Project: sess.Project, Kind: store.KindMessage, Title: title, Body: body, Fields: fields,
	})
	if err != nil {
		return nil, err
	}
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
	t, err := c.newSession(ctx, sc.actor, p.Project, pool, p.Agent, key, pool+" "+key)
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
	send := func(ev rpc.SessionEvent) error {
		ev.Session = p.ID
		return sc.peer.conn.SendNotification(ctx, rpc.NotifySessionEvent, ev)
	}
	state := func(t store.Ticket) error {
		return send(rpc.SessionEvent{Type: rpc.EventState, Status: t.Status, Resolution: t.Resolution})
	}
	if err := state(t); err != nil {
		return nil, err
	}
	path := filepath.Join(c.root.SessionDir(p.ID), host.JournalFile)
	ev := eventer{tools: map[string]string{}}
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
			out, ok := ev.event(fr)
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
		case <-time.After(subscribePoll):
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

// eventer turns journal frames into session events, keeping tool titles
// across frames (updates often omit them).
type eventer struct{ tools map[string]string }

type journalMsg struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (e *eventer) event(fr host.Frame) (rpc.SessionEvent, bool) {
	out := rpc.SessionEvent{Seq: fr.Seq}
	if fr.T > 0 {
		at := time.UnixMilli(fr.T).UTC()
		out.Time = &at
	}
	var m journalMsg
	if json.Unmarshal(fr.Msg, &m) != nil {
		return out, false
	}
	switch {
	case m.Method == "session/prompt":
		var p struct {
			Prompt []struct {
				Text string `json:"text"`
			} `json:"prompt"`
		}
		_ = json.Unmarshal(m.Params, &p)
		var parts []string
		for _, b := range p.Prompt {
			if b.Text != "" {
				parts = append(parts, b.Text)
			}
		}
		out.Type, out.Text = rpc.EventTurnStart, strings.Join(parts, "\n")
		return out, true
	case m.Method == "session/update":
		var u struct {
			Update struct {
				Kind    string `json:"sessionUpdate"`
				Content struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
				ToolCallID string `json:"toolCallId"`
				Title      string `json:"title"`
				Status     string `json:"status"`
			} `json:"update"`
		}
		if json.Unmarshal(m.Params, &u) != nil {
			return out, false
		}
		up := u.Update
		switch up.Kind {
		case "agent_message_chunk":
			if (up.Content.Type == "text" || up.Content.Type == "") && up.Content.Text != "" {
				out.Type, out.Text = rpc.EventText, up.Content.Text
				return out, true
			}
		case "tool_call", "tool_call_update":
			title := up.Title
			if title == "" {
				title = e.tools[up.ToolCallID]
			}
			if title != "" {
				e.tools[up.ToolCallID] = title
			}
			// Updates without a status are output deltas.
			if up.Status == "" && up.Kind == "tool_call_update" {
				return out, false
			}
			status := up.Status
			if status == "" {
				status = "pending"
			}
			out.Type = rpc.EventTool
			out.Tool = &rpc.ToolSummary{ID: up.ToolCallID, Title: title, Status: status}
			return out, true
		}
	case m.Method == "" && len(m.Result) > 0:
		// Only session/prompt results carry a stopReason.
		var res struct {
			StopReason string `json:"stopReason"`
		}
		if json.Unmarshal(m.Result, &res) == nil && res.StopReason != "" {
			out.Type, out.StopReason = rpc.EventTurnEnd, res.StopReason
			return out, true
		}
	case m.Method == "" && m.Error != nil && fr.From == "agent":
		out.Type, out.Error = rpc.EventTurnEnd, m.Error.Message
		return out, true
	}
	return out, false
}
