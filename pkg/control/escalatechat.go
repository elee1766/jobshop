package control

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/store"
)

// Escalation chat (#379). When work is on hold for a human (a run or pool
// escalation ticket, a wait past its timeout, a guard ask_human, any task
// or step labeled needs-human), the controller opens one session in the
// configured chat pool, keyed escalation:<ticket>, and sends the pack's
// prompt for the cause as context for the model, not as a visible message.
// meta.escalation records the key and session with a conditional update,
// so a restart or a repeated sweep never opens a second chat. Closing the
// ticket sends one final message. Questions keep their own answer path and
// get no chat. No escalation chat config: the ticket only. Gateways and
// Desk attach to the key.

// escalationMetaPack is meta.escalation on an escalation ticket.
const escalationMetaPack = "escalation"

// escalationKeyPrefix starts every escalation chat session key.
const escalationKeyPrefix = "escalation:"

// EscalationMeta is meta.escalation.
type EscalationMeta struct {
	Key     string `json:"key"`
	Pool    string `json:"pool"`
	Session string `json:"session,omitempty"`
	// Cause is why the chat exists (cause* constants) and Held the ticket
	// on hold, for Desk's list.
	Cause string `json:"cause,omitempty"`
	Held  string `json:"held,omitempty"`
	Final bool   `json:"final,omitempty"`
}

func metaMap(raw json.RawMessage) map[string]json.RawMessage {
	m := map[string]json.RawMessage{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &m)
	}
	return m
}

func escalationMeta(t store.Ticket) (EscalationMeta, bool) {
	var e EscalationMeta
	raw, ok := metaMap(t.Meta)[escalationMetaPack]
	if !ok || json.Unmarshal(raw, &e) != nil || e.Key == "" {
		return e, false
	}
	return e, true
}

// Escalation causes (meta.escalation.cause, the cause template var).
const (
	causeRun   = "run"   // a run step failed at its retry limit
	causePool  = "pool"  // a pool is quarantined
	causeWait  = "wait"  // a wait passed its timeout
	causeGuard = "guard" // a guardrail hook asked a human
	causeLabel = "label" // anything else labeled needs-human
)

// escCause is what put work on hold for a human, as the prompt template
// sees it (config.EscalationVars). Strings are empty when the cause has
// no such part.
type escCause struct {
	kind, reason, held, heldTitle, worker      string
	run, workflow, step, issueRef, pool, quarn string
}

// maxReason bounds the reason text a prompt carries.
const maxReason = 1000

// holdTicket reports whether t is a hold that wants an escalation chat:
// a task or step labeled needs-human that is not a question. Questions
// (report --question, admission and intake approvals) keep their own
// answer path and get no chat.
func holdTicket(t store.Ticket) bool {
	return (t.Kind == store.KindTask || t.Kind == store.KindStep) && slices.Contains(t.Labels, needsHumanLabel) &&
		!slices.Contains(t.Labels, questionLabel) && metaString(t.Meta, questionMetaPack, "ticket") == ""
}

// causeOf classifies the hold ticket t and gathers its cause. A run
// escalation names the failed step and the session that last held it; a
// timed-out wait, a guard ask and a plain label name the held ticket.
func (c *Controller) causeOf(ctx context.Context, t store.Ticket) escCause {
	switch {
	case metaString(t.Meta, runMetaPack, "escalation") != "":
		x := escCause{
			kind: causeRun, run: metaString(t.Meta, runMetaPack, "id"), workflow: metaString(t.Meta, runMetaPack, "workflow"),
			issueRef: issueRef(t), reason: "a step failed at its retry limit",
		}
		if steps, err := store.Children(ctx, c.store, x.run); err == nil {
			for i := len(steps) - 1; i >= 0; i-- {
				if !exhaustedStep(steps[i]) {
					continue
				}
				st := steps[i]
				x.held, x.heldTitle, x.worker = st.ID, st.Title, c.lastHolder(ctx, st.ID)
				if sf, err := store.DecodeFields[store.StepFields](st); err == nil {
					x.step = sf.Step
					x.reason = fmt.Sprintf("step %s failed %d/%d times", sf.Step, sf.Failures, sf.Retries)
				}
				if r := failedReason(st); r != "" {
					x.reason += ": " + clip(r, maxReason)
				} else if cs, err := store.Comments(ctx, c.store, st.ID); err == nil && len(cs) > 0 {
					x.reason += ". Last comment by " + cs[len(cs)-1].Author + ": " + clip(cs[len(cs)-1].Body, maxReason)
				}
				break
			}
		}
		return x
	case metaString(t.Meta, poolMetaPack, "name") != "":
		q := metaString(t.Meta, poolMetaPack, "quarantine")
		return escCause{kind: causePool, pool: metaString(t.Meta, poolMetaPack, "name"), quarn: q, reason: q}
	case metaString(t.Meta, hookMetaPack, "session") != "":
		x := escCause{
			kind: causeGuard, worker: metaString(t.Meta, hookMetaPack, "session"), held: metaString(t.Meta, hookMetaPack, "ticket"),
			reason: metaString(t.Meta, hookMetaPack, "reason"), issueRef: issueRef(t),
		}
		if x.held != "" {
			if h, err := c.store.Get(ctx, x.held); err == nil {
				x.heldTitle = h.Title
			}
		}
		return x
	}
	x := escCause{kind: causeLabel, held: t.ID, heldTitle: t.Title, worker: t.Holder, issueRef: issueRef(t), reason: clip(t.Body, maxReason)}
	if w := waitOf(t); w != nil && w.TimedOut {
		x.kind = causeWait
		x.reason = fmt.Sprintf("wait (%s) passed its timeout %s; owner %s", w.Reason, w.Timeout.Format(time.RFC3339), cmp.Or(w.Owner, "unset"))
		x.worker = cmp.Or(w.Session, x.worker)
	} else if hist, err := store.History(ctx, c.store, t.ID); err == nil {
		for i := len(hist) - 1; i >= 0; i-- {
			if hist[i].Op == "label_add" && strings.Contains(string(hist[i].Diff), needsHumanLabel) {
				if hist[i].Actor != "controller" && hist[i].Actor != "admin" {
					x.worker = hist[i].Actor
				}
				break
			}
		}
		if cs, err := store.Comments(ctx, c.store, t.ID); err == nil && len(cs) > 0 {
			x.reason = clip(cs[len(cs)-1].Body, maxReason)
		}
	}
	if x.worker != "" {
		if s, err := c.store.Get(ctx, x.worker); err != nil || s.Kind != store.KindSession {
			x.worker = "" // an actor like admin, not a session to message
		}
	}
	return x
}

// escalationChat returns the chat pool and the rendered prompt for hold
// ticket t, or ok false when t gets no chat: no policy for its cause, or
// a pool quarantine of the chat pool itself.
func (c *Controller) escalationChat(ctx context.Context, t store.Ticket) (pool, prompt string, x escCause, ok bool) {
	cfg := c.config()
	x = c.causeOf(ctx, t)
	var e *config.EscalationChat
	var tmpl string
	switch x.kind {
	case causeRun:
		e = cfg.Escalation
		if w, has := cfg.Workflows[x.workflow]; has && w.Escalate != nil {
			e = w.Escalate
		}
		if e != nil {
			tmpl = e.Prompt
		}
	case causePool:
		if e = c.poolEscalationChat(cfg, t.Pool); e != nil && e.ChatPool != x.pool {
			tmpl = e.PoolPrompt
		}
	default:
		e = c.holdEscalationChat(cfg, t.Project)
		if e != nil {
			tmpl = e.HoldPrompt
		}
	}
	if e == nil || e.ChatPool == "" || tmpl == "" {
		return "", "", x, false
	}
	out, err := config.Fill(tmpl, map[string]any{
		"ticket": t.ID, "title": t.Title, "cause": x.kind, "reason": x.reason, "body": t.Body,
		"held": x.held, "held_title": x.heldTitle, "worker": x.worker,
		"run": x.run, "workflow": x.workflow, "step": x.step, "issue_ref": x.issueRef,
		"pool": x.pool, "quarantine": x.quarn, "error": x.quarn,
	})
	if err != nil {
		c.log.Warn("escalation prompt", "ticket", t.ID, "cause", x.kind, "err", err)
		return "", "", x, false
	}
	return e.ChatPool, out, x, true
}

// holdEscalationChat is the policy for a hold that is neither a run nor a
// pool escalation: no run names the workflow, so the first one (by name)
// with a hold prompt whose chat pool is in the ticket's project owns it,
// else the top-level block.
func (c *Controller) holdEscalationChat(cfg *config.Config, project string) *config.EscalationChat {
	ok := func(e *config.EscalationChat) bool {
		if e == nil || e.HoldPrompt == "" {
			return false
		}
		p, has := cfg.Pools[e.ChatPool]
		return has && p.Project == project
	}
	for _, name := range slices.Sorted(maps.Keys(cfg.Workflows)) {
		if e := cfg.Workflows[name].Escalate; ok(e) {
			return e
		}
	}
	if ok(cfg.Escalation) {
		return cfg.Escalation
	}
	return nil
}

// poolEscalationChat is the policy for a pool quarantine escalation filed
// in pool dest: the escalate block of a workflow escalating to dest that
// has a pool prompt, else the top-level block.
func (c *Controller) poolEscalationChat(cfg *config.Config, dest string) *config.EscalationChat {
	for _, name := range slices.Sorted(maps.Keys(cfg.Workflows)) {
		w := cfg.Workflows[name]
		if w.Escalate != nil && w.Escalate.PoolPrompt != "" && w.EscalatePool == dest {
			return w.Escalate
		}
	}
	return cfg.Escalation
}

// issueRef tells the chat session which issue to read, with its repo.
func issueRef(t store.Ticket) string {
	repo := ""
	if r := metaString(t.Meta, "github", "repo"); r != "" {
		repo = " --repo " + r
	}
	var m map[string]map[string]any
	_ = json.Unmarshal(t.Meta, &m)
	if k := metaString(t.Meta, "linear", "issue"); k != "" {
		return fmt.Sprintf("its Linear issue (`jobshop linear issue %s`)", k)
	}
	g := m["github"]
	switch {
	case g["issue"] != nil:
		return fmt.Sprintf("its GitHub issue (`gh issue view %v%s --comments`)", g["issue"], repo)
	case g["pr"] != nil:
		return fmt.Sprintf("its GitHub PR (`gh pr view %v%s --comments`)", g["pr"], repo)
	}
	return "any GitHub issue it names (the run's vars)"
}

// updateEscalationMeta reads ticket id, lets fn change its
// meta.escalation, and writes it conditional on the revision read,
// retrying on conflicts. fn returning errSkipEscalation ends without a
// write.
func (c *Controller) updateEscalationMeta(ctx context.Context, id string, fn func(t store.Ticket, e *EscalationMeta, ok bool) error) (store.Ticket, error) {
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
		if errors.Is(err, store.ErrConflict) && attempt < 5 {
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
func (c *Controller) ensureEscalationChat(ctx context.Context, t store.Ticket) {
	if t.Status == store.StatusClosed || !holdTicket(t) {
		return
	}
	if e, ok := escalationMeta(t); ok && e.Session != "" {
		return
	}
	pool, prompt, cause, ok := c.escalationChat(ctx, t)
	if !ok {
		return
	}
	key := escalationKeyPrefix + t.ID
	_, err := c.updateEscalationMeta(ctx, t.ID, func(_ store.Ticket, e *EscalationMeta, ok bool) error {
		if ok {
			return errSkipEscalation
		}
		*e = EscalationMeta{Key: key, Pool: pool, Cause: cause.kind, Held: cause.held}
		return nil
	})
	if err != nil && !errors.Is(err, errSkipEscalation) {
		c.log.Warn("record escalation chat", "ticket", t.ID, "err", err)
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
	sess, started, err := c.keyedSession(ctx, sc, e.Pool, e.Key)
	if err != nil {
		c.log.Warn("escalation chat session", "ticket", t.ID, "pool", e.Pool, "err", err)
		return
	}
	if started {
		// The pack prompt is context for the model, not a visible message:
		// the session's first visible turn is its own summary.
		if err := c.sendMessage(ctx, sess, "", prompt); err != nil {
			c.log.Warn("escalation chat prompt", "ticket", t.ID, "err", err)
			return
		}
	}
	_, err = c.updateEscalationMeta(ctx, t.ID, func(_ store.Ticket, e *EscalationMeta, _ bool) error {
		e.Session = sess.ID
		return nil
	})
	if err != nil {
		c.log.Warn("record escalation session", "ticket", t.ID, "err", err)
		return
	}
	c.log.Info("escalation chat opened", "ticket", t.ID, "session", sess.ID, "key", e.Key)
}

// escalationClosed sends the final message once ticket t closed.
func (c *Controller) escalationClosed(ctx context.Context, t store.Ticket) {
	e, ok := escalationMeta(t)
	if !ok || e.Final || e.Session == "" || t.Status != store.StatusClosed {
		return
	}
	var sent EscalationMeta
	cur, err := c.updateEscalationMeta(ctx, t.ID, func(t store.Ticket, e *EscalationMeta, ok bool) error {
		if !ok || e.Final || t.Status != store.StatusClosed {
			return errSkipEscalation
		}
		e.Final = true
		sent = *e
		return nil
	})
	if err != nil {
		if !errors.Is(err, errSkipEscalation) {
			c.log.Warn("escalation final", "ticket", t.ID, "err", err)
		}
		return
	}
	sess, err := c.openSessionTicket(ctx, sent.Session)
	if err != nil {
		return // the chat already ended
	}
	res := cur.Resolution
	if res == "" {
		res = "done"
	}
	msg := fmt.Sprintf("jobshop: escalation %s closed (%s). Nothing more is needed unless the operator writes again.", t.ID, res)
	if err := c.sendMessage(ctx, sess, msg, ""); err != nil {
		c.log.Warn("escalation final", "ticket", t.ID, "err", err)
	}
}

// sendMessage queues a message to session sess: text is shown, msgContext
// goes to the model only (session.send context).
func (c *Controller) sendMessage(ctx context.Context, sess store.Ticket, text, msgContext string) error {
	fields, err := json.Marshal(store.MessageFields{To: sess.ID, From: "controller", Context: strings.TrimSpace(msgContext)})
	if err != nil {
		return err
	}
	title, body, _ := strings.Cut(text, "\n")
	_, err = c.store.Create(ctx, "controller", store.NewTicket{
		Project: sess.Project, Kind: store.KindMessage, Title: title, Body: body, Fields: fields,
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
	c.ensureEscalationChat(ctx, t)
}

// sweepEscalationChats repairs chats on every tick: open hold tickets
// without a session get one, and closed ones get their final message.
func (c *Controller) sweepEscalationChats(ctx context.Context) {
	if c.draining() {
		return
	}
	cfg := c.config()
	if cfg.Escalation == nil && !slices.ContainsFunc(slices.Sorted(maps.Keys(cfg.Workflows)), func(n string) bool { return cfg.Workflows[n].Escalate != nil }) {
		return
	}
	for _, st := range []store.Status{store.StatusOpen, store.StatusInProgress} {
		ts, err := c.store.List(ctx, store.Filter{Label: needsHumanLabel, Status: st})
		if err != nil {
			return
		}
		for _, t := range ts {
			c.ensureEscalationChat(ctx, t)
		}
	}
	// Recently closed escalations that still owe a final message.
	ts, err := c.store.List(ctx, store.Filter{Label: needsHumanLabel, Status: store.StatusClosed, Order: store.OrderNewest, Limit: 50})
	if err != nil {
		return
	}
	for _, t := range ts {
		c.escalationClosed(ctx, t)
	}
}
