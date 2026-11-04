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
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Escalation chat (#379). When the controller files an escalation ticket
// (a run escalation or a pool quarantine) and the workflow's `escalate`
// block names a chat pool, it starts a session of that pool keyed
// escalation:<ticket id> and sends the rendered prompt as its first
// message. The prompt is pack policy (workflows.<name>.escalate); the
// controller only renders and sends it. The session is recorded in
// meta.escalation so there is one chat per ticket across restarts, and
// Desk (or any other client) sends to it by key. Questions keep their own
// answer path and get no chat.

// escalationMetaPack is meta.escalation on an escalation ticket.
const escalationMetaPack = "escalation"

// escalationKeyPrefix starts the session key of an escalation chat.
const escalationKeyPrefix = "escalation:"

// escalationChat is meta.escalation: the chat the controller opened.
type escalationChat struct {
	Key     string `json:"key"`
	Session string `json:"session"`
	Pool    string `json:"pool,omitempty"`
	// Cause is why the chat exists (cause* constants) and Held the ticket
	// on hold, for Desk's list.
	Cause string `json:"cause,omitempty"`
	Held  string `json:"held,omitempty"`
	// Final is set once the closed message was sent (or the session was
	// already gone).
	Final bool `json:"final,omitempty"`
}

func escalationChatOf(t store.Ticket) (escalationChat, bool) {
	var m map[string]json.RawMessage
	var e escalationChat
	if json.Unmarshal(t.Meta, &m) != nil || json.Unmarshal(m[escalationMetaPack], &e) != nil || e.Session == "" {
		return e, false
	}
	return e, true
}

// escalationAs is the identity the chat's messages come from.
var escalationAs = scope{admin: true, actor: "controller"}

// Escalation causes (meta.escalation.cause, the {{.cause}} template var).
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
// an open task labeled needs-human that is not a question. Questions
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
		if steps, err := c.store.ListChildren(ctx, x.run); err == nil {
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
				} else if cs, err := c.store.Comments(ctx, st.ID); err == nil && len(cs) > 0 {
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
	} else if hist, err := c.store.History(ctx, t.ID); err == nil {
		for i := len(hist) - 1; i >= 0; i-- {
			if hist[i].Op == "label_add" && strings.Contains(string(hist[i].Diff), needsHumanLabel) {
				if hist[i].Actor != "controller" && hist[i].Actor != "admin" {
					x.worker = hist[i].Actor
				}
				break
			}
		}
		if cs, err := c.store.Comments(ctx, t.ID); err == nil && len(cs) > 0 {
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

// escalationOpening returns the chat pool and the rendered context for
// hold ticket t, or ok false when the workflow that owns it has no chat
// for this cause.
func (c *Controller) escalationOpening(ctx context.Context, t store.Ticket) (pool, text string, x escCause, ok bool, err error) {
	cfg := c.config()
	x = c.causeOf(ctx, t)
	var w config.Workflow
	var tmpl string
	pick := func(w config.Workflow) string {
		switch x.kind {
		case causeRun:
			return w.Escalate.Prompt
		case causePool:
			return w.Escalate.PoolPrompt
		}
		return w.Escalate.HoldPrompt
	}
	switch x.kind {
	case causeRun:
		w = cfg.Workflows[x.workflow]
		tmpl = pick(w)
	default:
		// No run names the workflow, so the first one (by name) with a
		// chat for this cause in the ticket's project owns the policy: for
		// a quarantine the one whose escalate_pool received the ticket.
		for _, name := range slices.Sorted(maps.Keys(cfg.Workflows)) {
			cand := cfg.Workflows[name]
			cp, has := cfg.Pools[cand.Escalate.ChatPool]
			if !has || pick(cand) == "" {
				continue
			}
			if x.kind == causePool && (t.Pool == "" || cand.EscalatePool != t.Pool) {
				continue
			}
			if x.kind != causePool && cp.Project != t.Project {
				continue
			}
			w, tmpl = cand, pick(cand)
			break
		}
	}
	if w.Escalate.ChatPool == "" || tmpl == "" {
		return "", "", x, false, nil
	}
	vars := maps.Clone(w.PackVars)
	if vars == nil {
		vars = map[string]any{}
	}
	maps.Copy(vars, map[string]any{
		"ticket": t.ID, "title": t.Title, "cause": x.kind, "reason": x.reason, "body": t.Body,
		"held": x.held, "held_title": x.heldTitle, "worker": x.worker,
		"run": x.run, "workflow": x.workflow, "step": x.step, "issue_ref": x.issueRef,
		"pool": x.pool, "quarantine": x.quarn,
	})
	text, err = w.Render(tmpl, vars)
	return w.Escalate.ChatPool, text, x, err == nil, err
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

// openEscalationChat starts the chat for t unless it has one or its
// workflow configures none. The ticket is read again under escMu, so the
// filing call and the sweep cannot both send the first message.
func (c *Controller) openEscalationChat(ctx context.Context, t store.Ticket) error {
	c.escMu.Lock()
	defer c.escMu.Unlock()
	t, err := c.store.Get(ctx, t.ID)
	if err != nil {
		return err
	}
	if t.Status == store.StatusClosed || !slices.Contains(t.Labels, needsHumanLabel) {
		return nil
	}
	if _, ok := escalationChatOf(t); ok {
		return nil
	}
	pool, text, cause, ok, err := c.escalationOpening(ctx, t)
	if err != nil {
		return fmt.Errorf("render first message: %w", err)
	}
	if !ok {
		return nil
	}
	key := escalationKeyPrefix + t.ID
	// The pack prompt is context for the model, not a visible message: the
	// session's first visible turn is its own summary.
	out, err := c.sessionSend(ctx, escalationAs, rpc.SessionSendParams{Pool: pool, Key: key, Context: text})
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	res := out.(rpc.SessionSendResult)
	c.log.Info("escalation chat opened", "ticket", t.ID, "pool", pool, "session", res.Session)
	return c.setMetaPack(ctx, t.ID, escalationMetaPack, escalationChat{Key: key, Session: res.Session, Pool: pool, Cause: cause.kind, Held: cause.held})
}

// holdChat opens the chat for ticket id after something put it on hold
// (a needs-human label, a guard ask). A failure is logged: the sweep
// retries it.
func (c *Controller) holdChat(ctx context.Context, id string) {
	t, err := c.store.Get(ctx, id)
	if err != nil || !holdTicket(t) || !c.escalationChatsConfigured() {
		return
	}
	if err := c.openEscalationChat(ctx, t); err != nil {
		c.log.Warn("escalation chat", "ticket", id, "err", err)
	}
}

// finishEscalationChat sends the one closed message to the chat of a
// closed escalation ticket and sets meta.escalation.final. A session that
// already ended gets none.
func (c *Controller) finishEscalationChat(ctx context.Context, t store.Ticket) error {
	if e, ok := escalationChatOf(t); !ok || e.Final || t.Status != store.StatusClosed {
		return nil
	}
	c.escMu.Lock()
	defer c.escMu.Unlock()
	t, err := c.store.Get(ctx, t.ID)
	if err != nil {
		return err
	}
	e, ok := escalationChatOf(t)
	if !ok || e.Final || t.Status != store.StatusClosed {
		return nil
	}
	res := t.Resolution
	if res == "" {
		res = "done"
	}
	text := fmt.Sprintf("Ticket %s is closed (%s). Nothing more is needed unless the operator writes again.", t.ID, res)
	_, err = c.sessionSend(ctx, escalationAs, rpc.SessionSendParams{ID: e.Session, Text: text})
	if err != nil && !errors.Is(err, store.ErrClosed) && !errors.Is(err, store.ErrNotFound) {
		return fmt.Errorf("send: %w", err)
	}
	e.Final = true
	return c.setMetaPack(ctx, t.ID, escalationMetaPack, e)
}

// escalationChatsConfigured reports whether any workflow opens chats.
func (c *Controller) escalationChatsConfigured() bool {
	for _, w := range c.config().Workflows {
		if w.Escalate.ChatPool != "" {
			return true
		}
	}
	return false
}

// escalationChatSweepLimit bounds the tickets one sweep reads per status.
const escalationChatSweepLimit = 200

// sweepEscalationChats opens a chat for each open escalation ticket that
// lacks one (the filing call failed, or the config gained a chat pool)
// and sends the closed message for tickets closed since. It runs each
// tick.
func (c *Controller) sweepEscalationChats(ctx context.Context) {
	if c.draining() || !c.escalationChatsConfigured() {
		return
	}
	var filters []store.Filter
	for _, k := range []store.Kind{store.KindTask, store.KindStep} {
		filters = append(filters,
			store.Filter{Kind: k, Status: store.StatusOpen, Label: needsHumanLabel, Limit: escalationChatSweepLimit},
			store.Filter{Kind: k, Status: store.StatusInProgress, Label: needsHumanLabel, Limit: escalationChatSweepLimit},
			store.Filter{Kind: k, Status: store.StatusClosed, Label: needsHumanLabel, Newest: true, Limit: escalationChatSweepLimit})
	}
	for _, f := range filters {
		ts, err := c.store.List(ctx, f)
		if err != nil {
			c.log.Warn("escalation chat: list", "err", err)
			return
		}
		for _, t := range ts {
			switch {
			case t.Status == store.StatusClosed:
				err = c.finishEscalationChat(ctx, t)
			case holdTicket(t):
				err = c.openEscalationChat(ctx, t)
			default:
				continue
			}
			if err != nil {
				c.log.Warn("escalation chat", "ticket", t.ID, "err", err)
			}
		}
	}
}
