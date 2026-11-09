package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"slices"
	"strings"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/attention"
	"github.com/elee1766/jobshop/pkg/control/chatlife"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Attention chats (#535, moved here from Desk by #606). Every attention
// item resolves in a chat: a ticket's escalation chat when it has one
// (#379), else a chat-pool session keyed attention:<kind>:<id>. When an
// item enters attention the controller starts that chat and lets its
// first turn run, so a click attaches to a session that already
// answered. The key is kept on tickets in meta.desk.attention_chat;
// other items use the default key. A chat ends when its item leaves
// attention.
//
// The rules (keys, prompt, which chats ended, read-only limits) are in
// pkg/control/chatlife and pkg/control/attention. This file starts,
// resumes, suspends and closes the sessions, and writes the meta.

const (
	attentionKeyPrefix = chatlife.AttentionPrefix
	deskMetaPack       = domain.MetaDesk
)

var attentionTicketKinds = attention.TicketKinds

func attentionKey(kind, id string) string { return attention.Key(kind, id) }

// attentionChat is attention.chat: the item's chat, started when it has
// none. A pre-started chat is attached to and nothing is sent.
func (c *Controller) attentionChat(ctx context.Context, sc scope, p rpc.AttentionChatParams) (any, error) {
	if err := requireAdmin(sc, "attention chat"); err != nil {
		return nil, err
	}
	if p.Kind == "" || p.ID == "" || strings.ContainsAny(p.Kind+p.ID, " \n:") {
		return nil, fmt.Errorf("%w: kind and id are required", domain.ErrInvalid)
	}
	res, err := c.ensureAttentionChat(ctx, c.config(), rpc.AttentionItem{Kind: p.Kind, ID: p.ID, Title: p.Title}, p.Why, false)
	if err != nil {
		return nil, err
	}
	if !res.Started && res.Session != "" {
		// A click on a suspended chat resumes it now, so the session is
		// back by the time the operator types.
		c.resumeSession(ctx, res.Session)
	}
	return res, nil
}

// ensureAttentionChat finds or starts the item's chat. prestart marks a
// chat the controller starts on its own: it is suspended when idle and
// closed when the item leaves attention.
func (c *Controller) ensureAttentionChat(ctx context.Context, cfg *config.Config, it rpc.AttentionItem, why string, prestart bool) (rpc.AttentionChatResult, error) {
	res := rpc.AttentionChatResult{Kind: it.Kind, ID: it.ID, Key: attentionKey(it.Kind, it.ID), Pool: cfg.AttentionChatPool()}
	var t domain.Ticket
	isTicket := slices.Contains(attentionTicketKinds, it.Kind)
	if isTicket {
		var err error
		if t, err = c.store.Get(ctx, it.ID); err != nil {
			return res, err
		}
		// The escalation chat answers for the ticket (#379).
		if e, ok := escalationMeta(t); ok && it.Kind != rpc.AttentionQuestion {
			res.Key, res.Session, res.Pool = e.Key, e.Session, e.Pool
			if e.Session == "" {
				return res, fmt.Errorf("%w: escalation chat for %s is still starting", domain.ErrConflict, t.ID)
			}
			return res, nil
		}
		if k := metaString(t.Meta, deskMetaPack, attention.MetaChat); k != "" {
			res.Key = k
		}
	}
	if res.Pool == "" {
		return res, fmt.Errorf("%w: no chat pool for attention chats (attention.prestart_pool, or an operator chat pool)", domain.ErrInvalid)
	}
	sess, ok, err := c.openKeyed(ctx, res.Key)
	if err != nil {
		return res, err
	}
	if ok {
		res.Session, res.Pool = sess.ID, sess.Pool
		return res, nil
	}
	if prestart {
		// Explicitly archived chats stay closed until the operator discusses the item again.
		closed, err := c.store.List(ctx, store.Filter{Kind: domain.KindSession, Status: domain.StatusClosed, Key: res.Key, Order: store.OrderNewest, Limit: 1})
		if err != nil {
			return res, err
		}
		if len(closed) > 0 {
			f, _ := domain.DecodeFields[domain.SessionFields](closed[0])
			if f.Archived != nil {
				return res, nil
			}
		}
	}
	sc := scope{admin: true, actor: "controller"}
	sess, started, err := c.keyedSession(ctx, cfg, sc, res.Pool, res.Key)
	if err != nil {
		return res, err
	}
	res.Session = sess.ID
	if !started {
		return res, nil
	}
	res.Started = true
	// Read-only until a human writes: the first turn reads the item and
	// proposes, it does not act (#606).
	c.setReadOnly(sess.ID, true)
	if err := c.updateSessionFields(ctx, sess.ID, func(f *domain.SessionFields) {
		f.ReadOnly, f.Prestart = true, prestart
	}); err != nil {
		c.log.WarnContext(ctx, "attention chat fields", "session", sess.ID, "err", err)
	}
	// The prompt is context for the model: the first visible turn is its
	// own summary, as for escalation chats.
	if err := c.sendMessage(ctx, sess, "", c.attentionPrompt(ctx, it, why, t)); err != nil {
		return res, err
	}
	if isTicket {
		if err := c.recordAttentionChat(ctx, t.ID, res.Key); err != nil {
			c.log.WarnContext(ctx, "record attention chat", "ticket", t.ID, "err", err)
		}
	}
	c.log.InfoContext(ctx, "attention chat opened", "item", it.Kind+":"+it.ID, "session", sess.ID, "prestart", prestart)
	return res, nil
}

// openKeyed is the open session for key.
func (c *Controller) openKeyed(ctx context.Context, key string) (domain.Ticket, bool, error) {
	open, err := c.store.List(ctx, store.Filter{Kind: domain.KindSession, Status: domain.StatusOpen})
	if err != nil {
		return domain.Ticket{}, false, err
	}
	for _, t := range open {
		if f, err := domain.DecodeFields[domain.SessionFields](t); err == nil && f.Key == key {
			return t, true, nil
		}
	}
	return domain.Ticket{}, false, nil
}

// recordAttentionChat keeps key in meta.desk.attention_chat.
func (c *Controller) recordAttentionChat(ctx context.Context, id, key string) error {
	return c.updateDeskMeta(ctx, id, attention.MetaChat, key)
}

// recordAttentionSummary keeps a read-only attention chat's reply as the
// item's row summary, meta.desk.attention_summary (#607).
func (c *Controller) recordAttentionSummary(ctx context.Context, session, said string) {
	t, err := c.store.Get(ctx, session)
	if err != nil {
		return
	}
	f, err := domain.DecodeFields[domain.SessionFields](t)
	if err != nil {
		return
	}
	id, s, ok := chatlife.Summarize(f.Key, said)
	if !ok {
		return
	}
	if s.Recommend != "" {
		if err := c.updateDeskMeta(ctx, id, attention.MetaRecommend, s.Recommend); err != nil {
			c.log.WarnContext(ctx, "record attention recommendation", "ticket", id, "err", err)
		}
	}
	if err := c.updateDeskMeta(ctx, id, attention.MetaContext, s.Context); err != nil {
		c.log.WarnContext(ctx, "record attention context", "ticket", id, "err", err)
	}
	if err := c.updateDeskMeta(ctx, id, attention.MetaSummary, s.Row); err != nil {
		c.log.WarnContext(ctx, "record attention summary", "ticket", id, "err", err)
	}
}

// updateDeskMeta sets meta.desk.<field>, keeping the pack's other keys.
func (c *Controller) updateDeskMeta(ctx context.Context, id, field, value string) error {
	for attempt := 0; ; attempt++ {
		t, err := c.store.Get(ctx, id)
		if err != nil || metaString(t.Meta, deskMetaPack, field) == value {
			return err
		}
		m := metaMap(t.Meta)
		d := map[string]any{}
		if raw, ok := m[deskMetaPack]; ok {
			_ = json.Unmarshal(raw, &d)
		}
		d[field] = value
		m[deskMetaPack], _ = json.Marshal(d)
		mb, err := json.Marshal(m)
		if err != nil {
			return err
		}
		_, err = c.store.Update(ctx, "controller", id, store.Update{Meta: mb, IfRevision: t.Revision})
		if errors.Is(err, domain.ErrConflict) && attempt < 5 {
			continue
		}
		return err
	}
}

// attentionPrompt is the context of a resolving chat's first turn.
func (c *Controller) attentionPrompt(ctx context.Context, it rpc.AttentionItem, why string, t domain.Ticket) string {
	p := chatlife.AttentionPrompt{Item: it, Why: why, Ticket: t, Actions: c.attentionActions(ctx, it)}
	if it.Kind == rpc.AttentionRun && t.Kind == domain.KindRun {
		if run, err := c.runResult(ctx, t); err == nil {
			p.Run = &run
		}
	}
	return p.Text()
}

// sweepAttentionChats starts chats for items that entered attention, up
// to attention.prestart_max open at once, oldest item first, and closes
// attention chats whose item left (#606).
func (c *Controller) sweepAttentionChats(ctx context.Context, cfg *config.Config) {
	if c.draining() {
		return
	}
	c.sweepQuestionFollows(ctx)
	pool := cfg.AttentionChatPool()
	if pool == "" {
		return
	}
	items, err := c.attentionItems(ctx, cfg)
	if err != nil {
		return
	}
	open, err := c.store.List(ctx, store.Filter{Kind: domain.KindSession, Status: domain.StatusOpen})
	if err != nil {
		return
	}
	// Answered is not settled: a question's chat follows the asker (JOB-52).
	left, prestarted := chatlife.Sweep(items, open, func(s domain.Ticket, f domain.SessionFields) bool { return c.questionChatOpen(ctx, s, f) })
	for _, l := range left {
		c.closeAttentionChat(ctx, cfg, l.Session, l.Item)
	}
	if !cfg.PrestartOn() {
		return
	}
	limit := cfg.Attention.PrestartLimit()
	p := cfg.Pools[pool]
	for _, it := range items {
		if prestarted >= limit {
			return
		}
		if p.MaxSessions > 0 {
			// Pre-started chats take pool slots, never the human one.
			n, err := c.store.List(ctx, store.Filter{Kind: domain.KindSession, Status: domain.StatusOpen, Pool: pool})
			if err != nil || len(n) >= p.MaxSessions {
				return
			}
		}
		if !c.prestartable(ctx, cfg, it) {
			continue
		}
		res, err := c.ensureAttentionChat(ctx, cfg, it, "", true)
		if err != nil {
			c.log.WarnContext(ctx, "prestart attention chat", "item", it.Kind+":"+it.ID, "err", err)
			continue
		}
		if res.Started {
			prestarted++
		}
	}
}

// prestartable: the item gets a pre-started chat. Escalation tickets
// get the escalation chat from sweepEscalationChats instead.
func (c *Controller) prestartable(ctx context.Context, cfg *config.Config, it rpc.AttentionItem) bool {
	if !slices.Contains(attentionTicketKinds, it.Kind) {
		return true
	}
	t, err := c.store.Get(ctx, it.ID)
	if err != nil {
		return false
	}
	if it.Kind == rpc.AttentionQuestion {
		return true
	}
	if _, ok := escalationMeta(t); ok {
		return false
	}
	if holdTicket(t) {
		if _, _, _, ok := c.escalationChat(ctx, cfg, t); ok {
			return false
		}
	}
	return true
}

// closeAttentionChat ends the chat of an item that left attention.
func (c *Controller) closeAttentionChat(ctx context.Context, cfg *config.Config, s domain.Ticket, item string) {
	c.log.InfoContext(ctx, "attention item left, closing its chat", "item", item, "session", s.ID)
	_, err := c.closeTicketIn(ctx, cfg, scope{admin: true, actor: "controller"}, rpc.CloseParams{ID: s.ID, Resolution: "resolved"})
	if err != nil && !errors.Is(err, domain.ErrClosed) {
		c.log.WarnContext(ctx, "close attention chat", "session", s.ID, "err", err)
	}
}

// setReadOnly records a session's read-only state for the socket gate.
func (c *Controller) setReadOnly(id string, on bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if on {
		c.readOnly[id] = true
	} else {
		delete(c.readOnly, id)
	}
}

func (c *Controller) isReadOnly(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.readOnly[id]
}

// checkReadOnly refuses a write from a read-only session. Which methods
// count as reads is chatlife.ReadOnlyMethods.
func (c *Controller) checkReadOnly(sc scope, method string) error {
	if sc.admin || sc.session == "" || !c.isReadOnly(sc.session) || chatlife.ReadOnlyMethod(method) {
		return nil
	}
	return forbidden("session %s is read-only until the operator sends a message (%s)", sc.session, method)
}

// humanWrote lifts a session's read-only state: a human sent it a message.
func (c *Controller) humanWrote(ctx context.Context, id string) {
	if !c.isReadOnly(id) {
		return
	}
	c.setReadOnly(id, false)
	if err := c.updateSessionFields(ctx, id, func(f *domain.SessionFields) { f.ReadOnly = false }); err != nil {
		c.log.WarnContext(ctx, "clear read-only", "session", id, "err", err)
	}
}

// resumeSession wakes a suspended session so its host is back before the
// next message.
func (c *Controller) resumeSession(ctx context.Context, id string) {
	t, err := c.store.Get(ctx, id)
	if err != nil {
		return
	}
	if f, err := domain.DecodeFields[domain.SessionFields](t); err != nil || !f.Suspended {
		return
	}
	if err := c.updateSessionFields(ctx, id, func(f *domain.SessionFields) { f.Suspended = false }); err != nil {
		c.log.WarnContext(ctx, "resume session", "session", id, "err", err)
	}
	c.wakeRunner(id)
}

// errSuspended: a pre-started chat's host was stopped while idle; the
// session stays open and the runner waits for a message (#606).
var errSuspended = errors.New("session suspended")

// prestarted: the session is a pre-started attention chat.
func (r *runner) prestarted() bool {
	t, err := r.c.store.Get(r.ctx, r.id)
	if err != nil {
		return false
	}
	f, err := domain.DecodeFields[domain.SessionFields](t)
	return err == nil && f.Prestart
}

// suspend stops an idle pre-started chat's host and keeps the session
// open, so it takes no memory until someone clicks or writes. The next
// session() resumes the conversation with session/load and the journal
// keeps the transcript Desk replays.
func (r *runner) suspend(nc net.Conn) error {
	if err := r.c.updateSessionFields(r.ctx, r.id, func(f *domain.SessionFields) { f.Suspended = true }); err != nil {
		return err
	}
	r.log.InfoContext(r.ctx, "suspending idle attention chat")
	r.lifecycle("suspend")
	_ = nc.Close()
	r.c.stopHost(r.dir)
	return errSuspended
}

// suspended waits while the session is suspended: until a message for
// it arrives, attention.chat resumes it, or it closes. False when the
// runner is stopping.
func (r *runner) suspended() bool {
	for {
		t, err := r.c.store.Get(r.ctx, r.id)
		if err == nil {
			f, _ := domain.DecodeFields[domain.SessionFields](t)
			if t.Status == domain.StatusClosed || !f.Suspended {
				return true
			}
			if len(r.inbox(r.ctx, false)) > 0 {
				if err := r.c.updateSessionFields(r.ctx, r.id, func(f *domain.SessionFields) { f.Suspended = false }); err == nil {
					return true
				}
			}
		}
		select {
		case <-r.ctx.Done():
			return false
		case <-r.restartWake:
		case <-r.c.after(r.c.opt.Poll):
		}
	}
}
