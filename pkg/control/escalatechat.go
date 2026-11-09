package control

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/chatlife"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/store"
)

// Escalation chat (#379). The rules (which holds get a chat, the cause
// and prompt, the meta.escalation shape) are in pkg/control/chatlife
// (JOB-147). This file opens the session, writes meta.escalation with a
// conditional update, so a restart or a repeated sweep never opens a
// second chat, and sends the final message when the ticket closes.
// Closing the ticket sends one final message. Questions keep their own
// answer path and get no chat. No escalation chat config: the ticket
// only. Gateways and Desk attach to the key.

// escalationMetaPack is meta.escalation on an escalation ticket.
const escalationMetaPack = domain.MetaEscalation

// escalationKeyPrefix starts every escalation chat session key.
const escalationKeyPrefix = chatlife.EscalationPrefix

// EscalationMeta is meta.escalation.
type EscalationMeta = chatlife.Escalation

func metaMap(raw json.RawMessage) map[string]json.RawMessage { return domain.MetaMap(raw) }

func escalationMeta(t domain.Ticket) (EscalationMeta, bool) { return chatlife.EscalationOf(t) }

// holdTicket reports whether t is a hold that wants an escalation chat.
func holdTicket(t domain.Ticket) bool { return chatlife.Hold(t) }

// escalationChat returns the chat pool and the rendered prompt for hold
// ticket t, or ok false when t gets no chat: no policy for its cause, or
// a pool quarantine of the chat pool itself.
func (c *Controller) escalationChat(ctx context.Context, cfg *config.Config, t domain.Ticket) (pool, prompt string, x chatlife.Cause, ok bool) {
	x = chatlife.ClassifyHold(ctx, c.store, t)
	pool, prompt, ok, err := chatlife.Policy(cfg, t, x)
	if err != nil {
		c.log.WarnContext(ctx, "escalation prompt", "ticket", t.ID, "cause", x.Kind, "err", err)
		return "", "", x, false
	}
	return pool, prompt, x, ok
}

// updateEscalationMeta reads ticket id, lets fn change its
// meta.escalation, and writes it conditional on the revision read,
// retrying on conflicts. fn returning errSkipEscalation ends without a
// write.
func (c *Controller) updateEscalationMeta(ctx context.Context, id string, fn func(t domain.Ticket, e *EscalationMeta, ok bool) error) (domain.Ticket, error) {
	for attempt := 0; ; attempt++ {
		t, err := c.store.Get(ctx, id)
		if err != nil {
			return t, err
		}
		e, ok := escalationMeta(t)
		if err := fn(t, &e, ok); err != nil {
			return t, err
		}
		m := metaMap(t.Meta)
		m[escalationMetaPack], _ = json.Marshal(e)
		mb, err := json.Marshal(m)
		if err != nil {
			return t, err
		}
		t, err = c.store.Update(ctx, "controller", id, store.Update{Meta: mb, IfRevision: t.Revision})
		if errors.Is(err, domain.ErrConflict) && attempt < 5 {
			continue
		}
		return t, err
	}
}

var errSkipEscalation = errors.New("skip")

// ensureEscalationChat opens the chat for open hold ticket t unless it
// has one. The key is recorded before the first send, so a crash between
// the two resends to the same keyed session, and the prompt goes only to
// a session this call started.
func (c *Controller) ensureEscalationChat(ctx context.Context, cfg *config.Config, t domain.Ticket) {
	if t.Status == domain.StatusClosed || !holdTicket(t) {
		return
	}
	if e, ok := escalationMeta(t); ok && e.Session != "" {
		return
	}
	pool, prompt, cause, ok := c.escalationChat(ctx, cfg, t)
	if !ok {
		return
	}
	key := escalationKeyPrefix + t.ID
	_, err := c.updateEscalationMeta(ctx, t.ID, func(_ domain.Ticket, e *EscalationMeta, ok bool) error {
		if ok {
			return errSkipEscalation
		}
		*e = EscalationMeta{Key: key, Pool: pool, Cause: cause.Kind, Held: cause.Held}
		return nil
	})
	if err != nil && !errors.Is(err, errSkipEscalation) {
		c.log.WarnContext(ctx, "record escalation chat", "ticket", t.ID, "err", err)
		return
	}
	cur, err := c.store.Get(ctx, t.ID)
	if err != nil {
		return
	}
	e, _ := escalationMeta(cur)
	if e.Session != "" {
		return
	}
	sc := scope{admin: true, actor: "controller"}
	sess, started, err := c.keyedSession(ctx, cfg, sc, e.Pool, e.Key)
	if err != nil {
		c.log.WarnContext(ctx, "escalation chat session", "ticket", t.ID, "pool", e.Pool, "err", err)
		return
	}
	if started {
		// The pack prompt is context for the model, not a visible message:
		// the session's first visible turn is its own summary.
		if err := c.sendMessage(ctx, sess, "", prompt); err != nil {
			c.log.WarnContext(ctx, "escalation chat prompt", "ticket", t.ID, "err", err)
			return
		}
	}
	_, err = c.updateEscalationMeta(ctx, t.ID, func(_ domain.Ticket, e *EscalationMeta, _ bool) error {
		e.Session = sess.ID
		return nil
	})
	if err != nil {
		c.log.WarnContext(ctx, "record escalation session", "ticket", t.ID, "err", err)
		return
	}
	c.log.InfoContext(ctx, "escalation chat opened", "ticket", t.ID, "session", sess.ID, "key", e.Key)
}

// escalationClosed sends the final message once ticket t closed.
func (c *Controller) escalationClosed(ctx context.Context, t domain.Ticket) {
	e, ok := escalationMeta(t)
	if !ok || e.Final || e.Session == "" || t.Status != domain.StatusClosed {
		return
	}
	if session, err := c.store.Get(ctx, e.Session); err == nil {
		if fields, err := domain.DecodeFields[domain.SessionFields](session); err == nil && fields.Named {
			return // A converted chat has an independent lifecycle.
		}
	}
	var sent EscalationMeta
	cur, err := c.updateEscalationMeta(ctx, t.ID, func(t domain.Ticket, e *EscalationMeta, ok bool) error {
		if !ok || e.Final || t.Status != domain.StatusClosed {
			return errSkipEscalation
		}
		e.Final = true
		sent = *e
		return nil
	})
	if err != nil {
		if !errors.Is(err, errSkipEscalation) {
			c.log.WarnContext(ctx, "escalation final", "ticket", t.ID, "err", err)
		}
		return
	}
	sess, err := c.openSessionTicket(ctx, sent.Session)
	if err != nil {
		return // the chat already ended
	}
	msg := chatlife.FinalMessage(t.ID, cur.Resolution)
	if err := c.sendMessage(ctx, sess, msg, ""); err != nil {
		c.log.WarnContext(ctx, "escalation final", "ticket", t.ID, "err", err)
	}
}

// sendMessage queues a message to session sess: text is shown, msgContext
// goes to the model only (session.send context).
func (c *Controller) sendMessage(ctx context.Context, sess domain.Ticket, text, msgContext string) error {
	fields, err := json.Marshal(domain.MessageFields{To: sess.ID, From: "controller", Context: strings.TrimSpace(msgContext)})
	if err != nil {
		return err
	}
	title, body, _ := strings.Cut(text, "\n")
	_, err = store.Create(ctx, c.store, "controller", store.NewTicket{
		Project: sess.Project, Kind: domain.KindMessage, Title: title, Body: body, Fields: fields,
	})

	if err == nil {
		c.wakeRunner(sess.ID)
	}
	return err
}

// holdChat opens the chat for ticket id after something put it on hold (a
// needs-human label, a guard ask, a wait timeout). The sweep retries what
// fails here.
func (c *Controller) holdChat(ctx context.Context, id string) {
	t, err := c.store.Get(ctx, id)
	if err != nil {
		return
	}
	c.ensureEscalationChat(ctx, c.config(), t)
}

// sweepEscalationChats repairs chats on every tick: open hold tickets
// without a session get one, and closed ones get their final message.
func (c *Controller) sweepEscalationChats(ctx context.Context, cfg *config.Config) {
	if c.draining() {
		return
	}
	if !chatlife.Configured(cfg) {
		return
	}
	for _, st := range []domain.Status{domain.StatusOpen, domain.StatusInProgress} {
		ts, err := c.store.List(ctx, store.Filter{Label: needsHumanLabel, Status: st})
		if err != nil {
			return
		}
		for _, t := range ts {
			c.ensureEscalationChat(ctx, cfg, t)
		}
	}
	// Recently closed escalations that still owe a final message.
	ts, err := c.store.List(ctx, store.Filter{Label: needsHumanLabel, Status: domain.StatusClosed, Order: store.OrderNewest, Limit: 50})
	if err != nil {
		return
	}
	for _, t := range ts {
		c.escalationClosed(ctx, t)
	}
}
