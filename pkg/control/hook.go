package control

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// hookMetaPack marks ask-human tickets filed by a hook, keyed so a
// repeated guard call finds the same ticket.
const hookMetaPack = "hook"

// claimDenyDefer is how long a ticket a claim hook refused stays out of
// the queue, so the session's next claim gets other work.
const claimDenyDefer = 10 * time.Minute

// hookEvent is the JSON a hook reads on stdin.
type hookEvent struct {
	On      string          `json:"on"`
	Action  string          `json:"action,omitempty"`
	Target  string          `json:"target,omitempty"`
	Detail  string          `json:"detail,omitempty"`
	Session string          `json:"session,omitempty"`
	Pool    string          `json:"pool,omitempty"`
	Project string          `json:"project,omitempty"`
	Ticket  *store.Ticket   `json:"ticket,omitempty"`
	Extra   json.RawMessage `json:"extra,omitempty"`
}

type hookAnswer struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

// runHooks runs every hook for the event in config order and returns the
// first answer that is not allow. No hooks means allow.
func (c *Controller) runHooks(ctx context.Context, ev hookEvent) hookAnswer {
	for _, h := range c.config().Hooks {
		if !h.Applies(ev.On, ev.Action, ev.Pool) {
			continue
		}
		a := c.runHook(ctx, h, ev)
		if a.Decision != rpc.GuardAllow {
			c.log.InfoContext(ctx, "hook decision", "on", ev.On, "action", ev.Action, "session", ev.Session,
				"decision", a.Decision, "reason", a.Reason)
			return a
		}
	}
	return hookAnswer{Decision: rpc.GuardAllow}
}

// runHook fails closed: a crash, timeout or unreadable answer is deny.
func (c *Controller) runHook(ctx context.Context, h config.Hook, ev hookEvent) hookAnswer {
	in, err := json.Marshal(ev)
	if err != nil {
		return hookAnswer{rpc.GuardDeny, "hook event: " + err.Error()}
	}
	ctx, cancel := context.WithTimeout(ctx, config.Duration(h.Timeout, 30*time.Second))
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", h.Command)
	cmd.Dir = c.root.Dir
	cmd.Stdin = bytes.NewReader(in)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if ctx.Err() != nil {
		return hookAnswer{rpc.GuardDeny, "hook timed out: " + h.Command}
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return hookAnswer{rpc.GuardDeny, fmt.Sprintf("hook exited %d: %s", exit.ExitCode(), strings.TrimSpace(stderr.String()))}
	}
	if err != nil {
		return hookAnswer{rpc.GuardDeny, "hook: " + err.Error()}
	}
	out = bytes.TrimSpace(out)
	if len(out) == 0 {
		return hookAnswer{Decision: rpc.GuardAllow}
	}
	var a hookAnswer
	if err := json.Unmarshal(out, &a); err != nil {
		return hookAnswer{rpc.GuardDeny, "hook answer is not JSON: " + err.Error()}
	}
	switch a.Decision {
	case rpc.GuardAllow, rpc.GuardDeny, rpc.GuardAskHuman:
		return a
	}
	return hookAnswer{rpc.GuardDeny, fmt.Sprintf("hook decision %q is not allow, deny or ask_human", a.Decision)}
}

// guard is `jobshop guard`: a session asks before an outside action.
// ask_human files one needs-human ticket per session, action and target;
// a later guard call reads it: closed done allows, closed otherwise
// denies, open is still pending.
func (c *Controller) guard(ctx context.Context, sc scope, p rpc.GuardParams) (any, error) {
	if p.Action == "" {
		return nil, fmt.Errorf("%w: action is required", store.ErrInvalid)
	}
	on := config.HookAction
	if p.Action == config.HookMerge {
		on = config.HookMerge
	}
	ev := hookEvent{
		On: on, Action: p.Action, Target: p.Target, Detail: p.Detail,
		Session: sc.session, Pool: sc.pool, Project: sc.project, Extra: p.Extra,
	}
	if !sc.admin {
		if id, err := c.current(ctx, sc, ""); err == nil {
			if t, err := c.store.Get(ctx, id); err == nil {
				ev.Ticket = &t
			}
		}
	}
	key := strings.Join([]string{sc.actor, p.Action, p.Target}, "/")
	if r, ok := c.askedHuman(ctx, key); ok {
		return r, nil
	}
	a := c.runHooks(ctx, ev)
	r := rpc.GuardResult{Decision: a.Decision, Reason: a.Reason}
	if a.Decision == rpc.GuardAskHuman {
		t, err := c.fileAskHuman(ctx, ev, key, a.Reason)
		if err != nil {
			return nil, err
		}
		r.Ticket = t.ID
	}
	return r, nil
}

// askedHuman returns the answer of an earlier ask_human ticket for key.
func (c *Controller) askedHuman(ctx context.Context, key string) (rpc.GuardResult, bool) {
	seen, err := c.store.List(ctx, store.Filter{MetaPack: hookMetaPack, MetaKey: "key", MetaValue: key, Newest: true, Limit: 1})
	if err != nil || len(seen) == 0 {
		return rpc.GuardResult{}, false
	}
	t := seen[0]
	r := rpc.GuardResult{Ticket: t.ID}
	switch {
	case t.Status != store.StatusClosed:
		r.Decision, r.Reason = rpc.GuardAskHuman, "waiting for a human on "+t.ID
	case t.Resolution == "done":
		r.Decision, r.Reason = rpc.GuardAllow, "approved on "+t.ID
	default:
		r.Decision, r.Reason = rpc.GuardDeny, fmt.Sprintf("refused on %s (%s)", t.ID, t.Resolution)
	}
	return r, true
}

func (c *Controller) fileAskHuman(ctx context.Context, ev hookEvent, key, reason string) (store.Ticket, error) {
	meta, err := json.Marshal(map[string]map[string]string{hookMetaPack: {"key": key}})
	if err != nil {
		return store.Ticket{}, err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "A guardrail hook asks a human before session %s does %q", ev.Session, ev.Action)
	if ev.Target != "" {
		fmt.Fprintf(&b, " on %s", ev.Target)
	}
	fmt.Fprintf(&b, ".\n\nReason: %s\n", reason)
	if ev.Detail != "" {
		fmt.Fprintf(&b, "\nDetail: %s\n", ev.Detail)
	}
	if ev.Ticket != nil {
		fmt.Fprintf(&b, "\nSession ticket: %s %s\n", ev.Ticket.ID, ev.Ticket.Title)
	}
	b.WriteString("\nClose done to allow, close with any other resolution to deny.")
	prio := 1
	project := ev.Project
	if project == "" {
		project = c.config().DefaultProject
	}
	return c.store.Create(ctx, "controller", store.NewTicket{
		Project: project, Kind: store.KindTask, Title: "Approve " + ev.Action + " by " + ev.Session,
		Body: b.String(), Priority: &prio, Labels: []string{needsHumanLabel}, Meta: meta,
	})
}

// claimHook runs claim hooks for a session claim; any answer but allow
// refuses the claim. ask_human on claim is a deny: there is no ticket to
// wait on before the claim exists.
func (c *Controller) claimHook(ctx context.Context, sc scope, t store.Ticket) error {
	if sc.admin || len(c.config().Hooks) == 0 {
		return nil
	}
	a := c.runHooks(ctx, hookEvent{On: config.HookClaim, Session: sc.session, Pool: sc.pool, Project: sc.project, Ticket: &t})
	if a.Decision == rpc.GuardAllow {
		return nil
	}
	return forbidden("claim hook %s: %s", a.Decision, a.Reason)
}
