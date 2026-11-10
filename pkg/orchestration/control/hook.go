package control

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/orchestration/hooks"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// hookMetaPack marks ask-human tickets filed by a hook, keyed so a
// repeated guard call finds the same ticket.
const hookMetaPack = "hook"

// claimDenyDefer is how long a ticket a claim hook refused stays out of
// the queue, so the session's next claim gets other work.
const claimDenyDefer = 10 * time.Minute

// runHooks runs every hook for the event in config order and returns the
// first answer that is not allow. No hooks means allow.
func (c *Controller) runHooks(ctx context.Context, cfg *config.Config, ev hooks.GuardEvent) hooks.Answer {
	for _, h := range cfg.Hooks {
		if !h.Applies(ev.On, ev.Action, ev.Pool) {
			continue
		}
		var a hooks.Answer
		if h.Handler != "" {
			a = c.runHookHandler(ctx, cfg, h, ev)
		} else {
			a = (hooks.Runner{Dir: c.root.Dir, After: c.after, Executor: hookCommands{c: c, cfg: cfg}}).Run(ctx, h, ev)
		}
		if a.Decision != domain.GuardAllow {
			c.log.InfoContext(ctx, "hook decision", "on", ev.On, "action", ev.Action, "session", ev.Session,
				"decision", a.Decision, "reason", a.Reason)
			return a
		}
	}
	return hooks.Answer{Decision: domain.GuardAllow}
}

// guard is `jobshop guard`: a session asks before an outside action.
// ask_human files one needs-human ticket per session, action and target;
// a later guard call reads it: closed done allows, closed otherwise
// denies, open is still pending.
func (c *Controller) guard(ctx context.Context, sc scope, p rpc.GuardParams) (any, error) {
	if p.Action == "" {
		return nil, fmt.Errorf("%w: action is required", domain.ErrInvalid)
	}
	on := config.HookAction
	if p.Action == config.HookMerge {
		on = config.HookMerge
	}
	ev := hooks.GuardEvent{
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
	a := c.runHooks(ctx, c.config(), ev)
	r := rpc.GuardResult{Decision: a.Decision, Reason: a.Reason}
	if a.Decision == domain.GuardAskHuman {
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
	seen, err := c.store.List(ctx, store.Filter{MetaPack: hookMetaPack, MetaKey: "key", MetaValue: key, Order: store.OrderNewest, Limit: 1})
	if err != nil || len(seen) == 0 {
		return rpc.GuardResult{}, false
	}
	t := seen[0]
	r := rpc.GuardResult{Ticket: t.ID}
	switch {
	case t.Status != domain.StatusClosed:
		r.Decision, r.Reason = domain.GuardAskHuman, "waiting for a human on "+t.ID
	case t.Resolution == "done":
		r.Decision, r.Reason = domain.GuardAllow, "approved on "+t.ID
	default:
		r.Decision, r.Reason = domain.GuardDeny, fmt.Sprintf("refused on %s (%s)", t.ID, t.Resolution)
	}
	return r, true
}

func (c *Controller) fileAskHuman(ctx context.Context, ev hooks.GuardEvent, key, reason string) (domain.Ticket, error) {
	// session, ticket and reason tell an escalation chat who asked and why.
	hm := map[string]string{"key": key, "session": ev.Session, "reason": clip(reason, maxReasonMeta)}
	if ev.Ticket != nil {
		hm["ticket"] = ev.Ticket.ID
	}
	meta, err := json.Marshal(map[string]map[string]string{hookMetaPack: hm})
	if err != nil {
		return domain.Ticket{}, err
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
	t, err := store.Create(ctx, c.store, "controller", store.NewTicket{
		Project: project, Kind: domain.KindTask, Title: "Approve " + ev.Action + " by " + ev.Session,
		Body: b.String(), Priority: &prio, Labels: []string{needsHumanLabel}, Meta: meta,
	})

	if err == nil {
		c.holdChat(ctx, t.ID)
	}
	return t, err
}

// maxReasonMeta bounds a hook reason kept in meta.hook.
const maxReasonMeta = 500

// claimHook runs claim hooks for a session claim; any answer but allow
// refuses the claim. ask_human on claim is a deny: there is no ticket to
// wait on before the claim exists.
func (c *Controller) claimHook(ctx context.Context, sc scope, t domain.Ticket) error {
	cfg := c.config()
	if sc.admin || len(cfg.Hooks) == 0 {
		return nil
	}
	a := c.runHooks(ctx, cfg, hooks.GuardEvent{On: config.HookClaim, Session: sc.session, Pool: sc.pool, Project: sc.project, Ticket: &t})
	if a.Decision == domain.GuardAllow {
		return nil
	}
	return forbidden("claim hook %s: %s", a.Decision, a.Reason)
}
