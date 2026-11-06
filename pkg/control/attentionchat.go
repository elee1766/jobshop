package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"slices"
	"strings"

	acp "github.com/coder/acp-go-sdk"

	"github.com/elee1766/jobshop/pkg/config"
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

const (
	attentionKeyPrefix = "attention:"
	deskMetaPack       = "desk"
)

// attentionTicketKinds are attention kinds whose ID is a ticket.
var attentionTicketKinds = []string{
	rpc.AttentionRun, rpc.AttentionTicket, rpc.AttentionQuestion, rpc.AttentionDecision,
	rpc.AttentionStale, rpc.AttentionWait,
}

func attentionKey(kind, id string) string { return attentionKeyPrefix + kind + ":" + id }

// attentionChat is attention.chat: the item's chat, started when it has
// none. A pre-started chat is attached to and nothing is sent.
func (c *Controller) attentionChat(ctx context.Context, sc scope, p rpc.AttentionChatParams) (any, error) {
	if err := requireAdmin(sc, "attention chat"); err != nil {
		return nil, err
	}
	if p.Kind == "" || p.ID == "" || strings.ContainsAny(p.Kind+p.ID, " \n:") {
		return nil, fmt.Errorf("%w: kind and id are required", store.ErrInvalid)
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
	var t store.Ticket
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
				return res, fmt.Errorf("%w: escalation chat for %s is still starting", store.ErrConflict, t.ID)
			}
			return res, nil
		}
		if k := metaString(t.Meta, deskMetaPack, "attention_chat"); k != "" {
			res.Key = k
		}
	}
	if res.Pool == "" {
		return res, fmt.Errorf("%w: no chat pool for attention chats (attention.prestart_pool, or an operator chat pool)", store.ErrInvalid)
	}
	sess, ok, err := c.openKeyed(ctx, res.Key)
	if err != nil {
		return res, err
	}
	if ok {
		res.Session, res.Pool = sess.ID, sess.Pool
		return res, nil
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
	if err := c.updateSessionFields(ctx, sess.ID, func(f *store.SessionFields) {
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
func (c *Controller) openKeyed(ctx context.Context, key string) (store.Ticket, bool, error) {
	open, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Status: store.StatusOpen})
	if err != nil {
		return store.Ticket{}, false, err
	}
	for _, t := range open {
		if f, err := store.DecodeFields[store.SessionFields](t); err == nil && f.Key == key {
			return t, true, nil
		}
	}
	return store.Ticket{}, false, nil
}

// recordAttentionChat keeps key in meta.desk.attention_chat.
func (c *Controller) recordAttentionChat(ctx context.Context, id, key string) error {
	return c.updateDeskMeta(ctx, id, "attention_chat", key)
}

// recordAttentionSummary keeps a read-only attention chat's reply as the
// item's row summary, meta.desk.attention_summary (#607).
func (c *Controller) recordAttentionSummary(ctx context.Context, session, said string) {
	said = strings.TrimSpace(said)
	if said == "" {
		return
	}
	t, err := c.store.Get(ctx, session)
	if err != nil {
		return
	}
	f, err := store.DecodeFields[store.SessionFields](t)
	if err != nil || !strings.HasPrefix(f.Key, attentionKeyPrefix) {
		return
	}
	kind, id, _ := strings.Cut(strings.TrimPrefix(f.Key, attentionKeyPrefix), ":")
	if !slices.Contains(attentionTicketKinds, kind) {
		return
	}
	if err := c.updateDeskMeta(ctx, id, "attention_summary", summaryOf(said)); err != nil {
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
		if errors.Is(err, store.ErrConflict) && attempt < 5 {
			continue
		}
		return err
	}
}

// attentionPrompt is the context of a resolving chat's first turn.
func (c *Controller) attentionPrompt(ctx context.Context, it rpc.AttentionItem, why string, t store.Ticket) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Desk attention item: %s %s\n", it.Kind, it.ID)
	title := it.Title
	if title == "" {
		title = t.Title
	}
	if title != "" {
		fmt.Fprintf(&b, "Title: %s\n", title)
	}
	if why != "" && why != it.Kind {
		fmt.Fprintf(&b, "Why: %s\n", why)
	}
	if t.ID != "" {
		var m struct {
			GitHub struct {
				Repo  string `json:"repo"`
				Issue any    `json:"issue"`
				PR    any    `json:"pr"`
			} `json:"github"`
		}
		_ = json.Unmarshal(t.Meta, &m)
		if m.GitHub.Repo != "" && m.GitHub.Issue != nil {
			fmt.Fprintf(&b, "Issue: https://github.com/%s/issues/%v\n", m.GitHub.Repo, m.GitHub.Issue)
		}
		if m.GitHub.Repo != "" && m.GitHub.PR != nil {
			fmt.Fprintf(&b, "PR: https://github.com/%s/pull/%v\n", m.GitHub.Repo, m.GitHub.PR)
		}
	}
	if it.Kind == rpc.AttentionRun && t.Kind == store.KindRun {
		if run, err := c.runResult(ctx, t); err == nil {
			out, _ := json.MarshalIndent(run, "", "  ")
			if len(out) > 16<<10 {
				out = append(out[:16<<10], "\n[cut]"...)
			}
			fmt.Fprintf(&b, "\n`jobshop run show %s`:\n```json\n%s\n```\n", it.ID, out)
		}
	}
	b.WriteString("\nThe operator opens this chat from Desk. Read the item with the jobshop CLI (read-only commands only: this session cannot write until the operator sends a message), then reply in a few lines: what blocks it, and the concrete actions available, as Desk's resolve buttons (approve, reject or dismiss for a decision; close, resume, unquarantine, clear wait) and the exact CLI commands. Do not act until the operator agrees.")
	if it.Kind == rpc.AttentionQuestion {
		qm, _ := questionOf(t)
		b.WriteString(questionChatPrompt(it.ID, qm))
	}
	return b.String()
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
	listed := map[string]bool{}
	for _, it := range items {
		listed[it.Kind+":"+it.ID] = true
		if slices.Contains(attentionTicketKinds, it.Kind) {
			listed["ticket:"+it.ID] = true
		}
	}
	open, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Status: store.StatusOpen})
	if err != nil {
		return
	}
	prestarted := 0
	for _, s := range open {
		f, err := store.DecodeFields[store.SessionFields](s)
		if err != nil || !strings.HasPrefix(f.Key, attentionKeyPrefix) {
			continue
		}
		kind, id, _ := strings.Cut(strings.TrimPrefix(f.Key, attentionKeyPrefix), ":")
		// A ticket listed under another kind (a question that became a
		// wait) still counts as in attention.
		inList := listed[kind+":"+id] || (slices.Contains(attentionTicketKinds, kind) && listed["ticket:"+id])
		if !inList && kind == rpc.AttentionQuestion && c.questionChatOpen(ctx, s, f) {
			// Answered is not settled: the chat follows the asker (JOB-52).
			inList = true
		}
		if !inList {
			c.closeAttentionChat(ctx, cfg, s, kind+":"+id)
			continue
		}
		if f.Prestart {
			prestarted++
		}
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
			n, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Status: store.StatusOpen, Pool: pool})
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
func (c *Controller) closeAttentionChat(ctx context.Context, cfg *config.Config, s store.Ticket, item string) {
	c.log.InfoContext(ctx, "attention item left, closing its chat", "item", item, "session", s.ID)
	_, err := c.closeTicketIn(ctx, cfg, scope{admin: true, actor: "controller"}, rpc.CloseParams{ID: s.ID, Resolution: "resolved"})
	if err != nil && !errors.Is(err, store.ErrClosed) {
		c.log.WarnContext(ctx, "close attention chat", "session", s.ID, "err", err)
	}
}

// readOnlyMethods are what a read-only session may call (#606).
var readOnlyMethods = []string{
	rpc.MethodWhoami, rpc.MethodVersion, rpc.MethodTicketShow, rpc.MethodTicketTrace, rpc.MethodTicketList,
	rpc.MethodReady, rpc.MethodRecall, rpc.MethodRunShow, rpc.MethodRunList, rpc.MethodBatchShow,
	rpc.MethodBatchList, rpc.MethodAttention, rpc.MethodConfigShow, rpc.MethodPoolList, rpc.MethodSessionList,
	rpc.MethodSessionJournal, rpc.MethodJournalSearch, rpc.MethodTriggerList, rpc.MethodActivityList,
	rpc.MethodLearnList, rpc.MethodMemoryShow, rpc.MethodReportIncidents, rpc.MethodLinearIntake,
	// Naming itself changes only its label (#626).
	rpc.MethodSessionRename,
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

// checkReadOnly refuses a write from a read-only session.
func (c *Controller) checkReadOnly(sc scope, method string) error {
	if sc.admin || sc.session == "" || !c.isReadOnly(sc.session) || slices.Contains(readOnlyMethods, method) {
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
	if err := c.updateSessionFields(ctx, id, func(f *store.SessionFields) { f.ReadOnly = false }); err != nil {
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
	if f, err := store.DecodeFields[store.SessionFields](t); err != nil || !f.Suspended {
		return
	}
	if err := c.updateSessionFields(ctx, id, func(f *store.SessionFields) { f.Suspended = false }); err != nil {
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
	f, err := store.DecodeFields[store.SessionFields](t)
	return err == nil && f.Prestart
}

// suspend stops an idle pre-started chat's host and keeps the session
// open, so it takes no memory until someone clicks or writes. The next
// session() resumes the conversation with session/load and the journal
// keeps the transcript Desk replays.
func (r *runner) suspend(nc net.Conn) error {
	if err := r.c.updateSessionFields(r.ctx, r.id, func(f *store.SessionFields) { f.Suspended = true }); err != nil {
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
			f, _ := store.DecodeFields[store.SessionFields](t)
			if t.Status == store.StatusClosed || !f.Suspended {
				return true
			}
			if len(r.inbox(r.ctx, false)) > 0 {
				if err := r.c.updateSessionFields(r.ctx, r.id, func(f *store.SessionFields) { f.Suspended = false }); err == nil {
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

// readOnlyTool: a tool call kind a read-only session may run. Execute
// stays allowed so the agent can run the jobshop CLI; its writes are
// refused at the session socket (checkReadOnly). Edit, delete and move
// are refused.
func readOnlyTool(k *acp.ToolKind) bool {
	if k == nil {
		return true
	}
	return !slices.Contains([]acp.ToolKind{acp.ToolKindEdit, acp.ToolKindDelete, acp.ToolKindMove}, *k)
}
