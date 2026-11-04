package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/store"
)

// Escalation chat (#379). When the controller files a run or pool
// escalation ticket, it opens one session in the configured chat pool,
// keyed escalation:<ticket>, and sends the pack's escalation prompt as
// its first message. meta.escalation records the key and session with a
// conditional update, so a restart or a repeated sweep never opens a
// second chat. Closing the ticket sends one final message. No escalation
// chat config: the ticket only. Gateways and Desk attach to the key.

// escalationMetaPack is meta.escalation on an escalation ticket.
const escalationMetaPack = "escalation"

// escalationKeyPrefix starts every escalation chat session key.
const escalationKeyPrefix = "escalation:"

// EscalationMeta is meta.escalation.
type EscalationMeta struct {
	Key     string `json:"key"`
	Pool    string `json:"pool"`
	Session string `json:"session,omitempty"`
	Final   bool   `json:"final,omitempty"`
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

// escalationChat returns the chat policy and rendered first message for
// escalation ticket t, or ok false when t gets no chat.
func (c *Controller) escalationChat(t store.Ticket) (pool, prompt string, ok bool) {
	cfg := c.config()
	m := metaMap(t.Meta)
	if raw, isRun := m[runMetaPack]; isRun {
		var r struct{ ID, Workflow, Escalation string }
		if json.Unmarshal(raw, &r) != nil || r.Escalation == "" {
			return "", "", false
		}
		e := cfg.Escalation
		if w, ok := cfg.Workflows[r.Workflow]; ok && w.Escalate != nil {
			e = w.Escalate
		}
		if e == nil || e.ChatPool == "" || e.Prompt == "" {
			return "", "", false
		}
		out, err := config.Fill(e.Prompt, e.Template, map[string]any{
			"ticket": t.ID, "run": r.ID, "workflow": r.Workflow, "issue_ref": issueRef(t), "body": t.Body,
		})
		if err != nil {
			c.log.Warn("escalation prompt", "ticket", t.ID, "err", err)
			return "", "", false
		}
		return e.ChatPool, out, true
	}
	if raw, isPool := m[poolMetaPack]; isPool {
		var p struct{ Name, Quarantine string }
		if json.Unmarshal(raw, &p) != nil || p.Name == "" {
			return "", "", false
		}
		e := c.poolEscalationChat(cfg, t.Pool)
		if e == nil || e.ChatPool == "" || e.PoolPrompt == "" || e.ChatPool == p.Name {
			return "", "", false
		}
		out, err := config.Fill(e.PoolPrompt, e.Template, map[string]any{
			"ticket": t.ID, "pool": p.Name, "error": p.Quarantine, "body": t.Body,
		})
		if err != nil {
			c.log.Warn("escalation pool prompt", "ticket", t.ID, "err", err)
			return "", "", false
		}
		return e.ChatPool, out, true
	}
	return "", "", false
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

// issueRef tells the session which issue or PR to read, with its repo.
func issueRef(t store.Ticket) string {
	var g struct {
		Repo  string `json:"repo"`
		Issue any    `json:"issue"`
		PR    any    `json:"pr"`
	}
	if raw, ok := metaMap(t.Meta)["github"]; ok {
		_ = json.Unmarshal(raw, &g)
	}
	repo := ""
	if g.Repo != "" {
		repo = " --repo " + g.Repo
	}
	switch {
	case g.Issue != nil:
		return fmt.Sprintf("its GitHub issue (`gh issue view %v%s --comments`)", g.Issue, repo)
	case g.PR != nil:
		return fmt.Sprintf("its GitHub PR (`gh pr view %v%s --comments`)", g.PR, repo)
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

// ensureEscalationChat opens the chat for open escalation ticket t unless
// it has one. The key is recorded before the first send, so a crash
// between the two resends to the same keyed session, and the prompt goes
// only to a session this call started.
func (c *Controller) ensureEscalationChat(ctx context.Context, t store.Ticket) {
	if t.Status == store.StatusClosed {
		return
	}
	if e, ok := escalationMeta(t); ok && e.Session != "" {
		return
	}
	pool, prompt, ok := c.escalationChat(t)
	if !ok {
		return
	}
	key := escalationKeyPrefix + t.ID
	_, err := c.updateEscalationMeta(ctx, t.ID, func(_ store.Ticket, e *EscalationMeta, ok bool) error {
		if ok {
			return errSkipEscalation
		}
		*e = EscalationMeta{Key: key, Pool: pool}
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
		if err := c.sendText(ctx, sess, prompt); err != nil {
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
	if err := c.sendText(ctx, sess, msg); err != nil {
		c.log.Warn("escalation final", "ticket", t.ID, "err", err)
	}
}

// sendText queues text as a message to session sess.
func (c *Controller) sendText(ctx context.Context, sess store.Ticket, text string) error {
	fields, err := json.Marshal(store.MessageFields{To: sess.ID, From: "controller"})
	if err != nil {
		return err
	}
	title, body, _ := strings.Cut(text, "\n")
	_, err = c.store.Create(ctx, "controller", store.NewTicket{
		Project: sess.Project, Kind: store.KindMessage, Title: title, Body: body, Fields: fields,
	})
	return err
}

// sweepEscalationChats repairs chats on every tick: open escalation
// tickets without a session get one, and closed ones get their final
// message.
func (c *Controller) sweepEscalationChats(ctx context.Context) {
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
	ts, err := c.store.List(ctx, store.Filter{Label: needsHumanLabel, Status: store.StatusClosed, Newest: true, Limit: 50})
	if err != nil {
		return
	}
	for _, t := range ts {
		c.escalationClosed(ctx, t)
	}
}
