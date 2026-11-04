package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"

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

// escalationPolicy returns the chat pool and the rendered first message
// for escalation ticket t, or ok false when its workflow has no
// escalation chat (or t is not a run or pool escalation).
func escalationPolicy(cfg *config.Config, t store.Ticket) (pool, text string, ok bool, err error) {
	var (
		w      config.Workflow
		tmpl   string
		extras map[string]any
	)
	switch {
	case metaString(t.Meta, runMetaPack, "escalation") != "":
		w = cfg.Workflows[metaString(t.Meta, runMetaPack, "workflow")]
		tmpl = w.Escalate.Prompt
		extras = map[string]any{
			"ticket": t.ID, "run": metaString(t.Meta, runMetaPack, "id"), "workflow": metaString(t.Meta, runMetaPack, "workflow"),
			"issue_ref": issueRef(t), "body": t.Body,
		}
	case metaString(t.Meta, poolMetaPack, "name") != "":
		// The workflow whose escalate_pool received the ticket owns the
		// policy, as it does for the quarantine ticket's destination.
		for _, name := range slices.Sorted(maps.Keys(cfg.Workflows)) {
			x := cfg.Workflows[name]
			if t.Pool != "" && x.EscalatePool == t.Pool && x.Escalate.ChatPool != "" && x.Escalate.PoolPrompt != "" {
				w, tmpl = x, x.Escalate.PoolPrompt
				break
			}
		}
		extras = map[string]any{
			"ticket": t.ID, "pool": metaString(t.Meta, poolMetaPack, "name"),
			"quarantine": metaString(t.Meta, poolMetaPack, "quarantine"), "body": t.Body,
		}
	}
	if w.Escalate.ChatPool == "" || tmpl == "" {
		return "", "", false, nil
	}
	vars := maps.Clone(w.PackVars)
	if vars == nil {
		vars = map[string]any{}
	}
	maps.Copy(vars, extras)
	text, err = config.FillTemplate(tmpl, vars)
	return w.Escalate.ChatPool, text, err == nil, err
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
	if t.Status == store.StatusClosed {
		return nil
	}
	if _, ok := escalationChatOf(t); ok {
		return nil
	}
	pool, text, ok, err := escalationPolicy(c.config(), t)
	if err != nil {
		return fmt.Errorf("render first message: %w", err)
	}
	if !ok {
		return nil
	}
	key := escalationKeyPrefix + t.ID
	out, err := c.sessionSend(ctx, escalationAs, rpc.SessionSendParams{Pool: pool, Key: key, Text: text})
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	res := out.(rpc.SessionSendResult)
	c.log.Info("escalation chat opened", "ticket", t.ID, "pool", pool, "session", res.Session)
	return c.setMetaPack(ctx, t.ID, escalationMetaPack, escalationChat{Key: key, Session: res.Session, Pool: pool})
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
	for _, f := range []store.Filter{
		{Kind: store.KindTask, Status: store.StatusOpen, Label: needsHumanLabel, Limit: escalationChatSweepLimit},
		{Kind: store.KindTask, Status: store.StatusInProgress, Label: needsHumanLabel, Limit: escalationChatSweepLimit},
		{Kind: store.KindTask, Status: store.StatusClosed, Label: needsHumanLabel, Newest: true, Limit: escalationChatSweepLimit},
	} {
		ts, err := c.store.List(ctx, f)
		if err != nil {
			c.log.Warn("escalation chat: list", "err", err)
			return
		}
		for _, t := range ts {
			if t.Status == store.StatusClosed {
				err = c.finishEscalationChat(ctx, t)
			} else {
				err = c.openEscalationChat(ctx, t)
			}
			if err != nil {
				c.log.Warn("escalation chat", "ticket", t.ID, "err", err)
			}
		}
	}
}
