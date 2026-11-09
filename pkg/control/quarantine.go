package control

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/elee1766/jobshop/pkg/conversation/attention"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/storage/store"

	"github.com/elee1766/jobshop/pkg/domain"
)

const (
	// crashReleaseReason marks a release by a crashed session in history,
	// so crashes per ticket can be counted.
	crashReleaseReason = "session crashed"
	quarantineLabel    = "quarantined"
	quarantineCrashes  = 3
	quarantineFor      = time.Hour
	poolCrashes        = 3
	poolCrashWindow    = 10 * time.Minute
	poolPause          = 15 * time.Minute
)

// crashRelease releases what a crashed session held with crashReleaseReason
// and quarantines tickets that have now crashed quarantineCrashes sessions.
func (c *Controller) crashRelease(ctx context.Context, sessionID string) {
	held, err := c.store.List(ctx, store.Filter{Holder: sessionID, Status: domain.StatusInProgress})
	if err != nil {
		return
	}
	for _, t := range held {
		cl := store.Claim{TicketID: t.ID, Holder: sessionID, Fence: t.Fence}
		if _, err := c.store.Release(ctx, "", cl.TicketID, &cl, store.Outcome{Reason: crashReleaseReason}); err != nil {
			c.log.WarnContext(ctx, "crash release", "ticket", t.ID, "err", err)
			continue
		}
		c.maybeQuarantine(ctx, t.ID)
		c.stepEvent(ctx, c.config(), t)
	}
}

// maybeQuarantine defers a ticket for quarantineFor and labels it once
// quarantineCrashes crashed sessions have released it since it was last
// quarantined (removing the label resets the count).
func (c *Controller) maybeQuarantine(ctx context.Context, id string) {
	hist, err := store.History(ctx, c.store, id)
	if err != nil {
		return
	}
	var sessions []string
	for _, h := range hist {
		var d struct {
			Reason string `json:"reason"`
			Label  string `json:"label"`
		}
		_ = json.Unmarshal(h.Diff, &d)
		switch {
		case (h.Op == "label_add" || h.Op == "label_remove") && d.Label == quarantineLabel:
			sessions = nil
		case h.Op == "release" && d.Reason == crashReleaseReason:
			sessions = append(sessions, h.Actor)
		}
	}
	if len(sessions) < quarantineCrashes {
		return
	}
	until := c.now().Add(quarantineFor)
	if _, err := c.store.Update(ctx, "controller", id, store.Update{DeferUntil: &until}); err != nil {
		c.log.ErrorContext(ctx, "quarantine", "ticket", id, "err", err)
		return
	}
	if err := store.AddLabel(ctx, c.store, "controller", id, quarantineLabel); err != nil {
		c.log.ErrorContext(ctx, "quarantine label", "ticket", id, "err", err)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Quarantined until %s: released by %d crashed sessions.\n", until.UTC().Format(time.RFC3339), len(sessions))
	for _, s := range sessions {
		reason := "unknown"
		if ex, ok := host.ReadExit(c.root.SessionDir(s)); ok {
			reason = fmt.Sprintf("%s (code %d %s)", ex.Reason, ex.ExitCode, ex.Signal)
		}
		fmt.Fprintf(&b, "- %s: %s\n", s, reason)
	}
	b.WriteString("Undo with `ticket update " + id + " --defer 0` and `label rm " + id + " quarantined`.")
	_, _ = store.AddComment(ctx, c.store, "controller", id, b.String())
	c.log.WarnContext(ctx, "quarantined ticket", "ticket", id, "sessions", sessions)
}

// poolCrashed records a crashed session of pool, ended by reason, and
// pauses the pool when poolCrashes happened within poolCrashWindow. The
// pause keeps the last reason and the pool's failed preflight checks
// (JOB-27), so attention and the finding detail name the error.
func (c *Controller) poolCrashed(pool, reason string) {
	if pool == "" {
		return
	}
	if !c.poolCrashLocked(pool) {
		return
	}
	why := strings.TrimSpace(reason)
	if why == "" {
		why = "sessions crash-looping"
	}
	if hint := c.preflightHint(pool); hint != "" {
		why += "\nisolation preflight:\n" + hint
	}
	c.mu.Lock()
	if c.poolPauseReason == nil {
		c.poolPauseReason = map[string]string{}
	}
	c.poolPauseReason[pool] = why
	until := c.poolPausedUntil[pool]
	c.mu.Unlock()
	c.log.Warn("pool quarantined: sessions crash-looping", "pool", pool, "until", until, "err", why)
}

// preflightHint is the failed checks of pool's isolation preflight, or
// "" when the pool is not isolated or the host passes.
func (c *Controller) preflightHint(pool string) string {
	cfg := c.config()
	if cfg == nil {
		return ""
	}
	spec, ok := c.preflightSpec(cfg, pool)
	if !ok {
		return ""
	}
	check := c.opt.Preflight
	if check == nil {
		check = host.Preflight
	}
	rep, err := check(spec)
	if err != nil || rep.OK() {
		return ""
	}
	var failed host.PreflightReport
	for _, ch := range rep.Checks {
		if !ch.OK {
			failed.Checks = append(failed.Checks, ch)
		}
	}
	return failed.String()
}

// poolPauseWhy is the reason pool was paused for crash-looping.
func (c *Controller) poolPauseWhy(pool string) string {
	if why := c.poolPauseReason[pool]; why != "" {
		return why
	}
	return "sessions crash-looping"
}

// poolCrashLocked logs a crash and reports whether it paused the pool.
func (c *Controller) poolCrashLocked(pool string) bool {
	now := c.now()
	c.mu.Lock()
	defer c.mu.Unlock()
	keep := c.poolCrashLog[pool][:0]
	for _, t := range c.poolCrashLog[pool] {
		if now.Sub(t) < poolCrashWindow {
			keep = append(keep, t)
		}
	}
	keep = append(keep, now)
	c.poolCrashLog[pool] = keep
	if len(keep) >= poolCrashes {
		c.poolPausedUntil[pool] = now.Add(poolPause)
		c.poolCrashLog[pool] = nil
		return true
	}
	return false
}

// poolPaused reports whether the supervisor should start no sessions for pool.
func (c *Controller) poolPaused(pool string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.poolQuarantined[pool]; ok {
		return true
	}
	return c.now().Before(c.poolPausedUntil[pool])
}

// poolMetaPack is the meta pack on a pool quarantine escalation:
// {"pool": {"name": <pool>, "quarantine": <error>}}.
const poolMetaPack = "pool"

// quarantinePool pauses pool until reload or unquarantine after a
// permanent spawn error (#146), and files one open escalation per pool.
// Retrying would hit the same bad path every time.
func (c *Controller) quarantinePool(ctx context.Context, pool, reason string) {
	if pool == "" {
		return
	}
	if !c.setQuarantine(pool, reason) {
		c.log.WarnContext(ctx, "pool quarantined: permanent spawn error", "pool", pool, "err", reason)
	}
	c.ensurePoolEscalation(ctx, pool, reason)
}

// setQuarantine records pool's permanent quarantine and reports whether
// it was quarantined already.
func (c *Controller) setQuarantine(pool, reason string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, already := c.poolQuarantined[pool]
	c.poolQuarantined[pool] = reason
	return already
}

// ensurePoolEscalation files "Pool <name> quarantined: <error>" unless an
// open one for the pool exists.
func (c *Controller) ensurePoolEscalation(ctx context.Context, pool, reason string) {
	for _, st := range []domain.Status{domain.StatusOpen, domain.StatusInProgress} {
		seen, err := c.store.List(ctx, store.Filter{MetaPack: poolMetaPack, MetaKey: "name", MetaValue: pool, Status: st, Limit: 1})
		if err != nil || len(seen) > 0 {
			return
		}
	}
	cfg := c.config()
	p := cfg.Pools[pool]
	proj := p.Project
	if proj == "" {
		proj = cfg.DefaultProject
	}
	// The escalation goes where run escalations go (a workflow's
	// escalate_pool in the same project), never to the quarantined pool.
	dest := ""
	for _, name := range slices.Sorted(maps.Keys(cfg.Workflows)) {
		ep := cfg.Workflows[name].EscalatePool
		if q, ok := cfg.Pools[ep]; ok && ep != pool && q.Project == proj {
			dest = ep
			break
		}
	}
	meta, err := json.Marshal(map[string]map[string]string{poolMetaPack: {"name": pool, "quarantine": reason}})
	if err != nil {
		return
	}
	body := fmt.Sprintf("Pool %s starts no sessions: a spawn failed with an error that will repeat on every attempt.\n\n%s\n\n"+
		"This is a host config problem (a repo or workspace path, an agent command). Fix the config, then `jobshop reload` "+
		"or `jobshop pool unquarantine %s`, and close this ticket.", pool, reason, pool)
	prio := 1
	t, err := store.Create(ctx, c.store, "controller", store.NewTicket{
		Project: proj, Kind: domain.KindTask, Title: clipTitle(fmt.Sprintf("Pool %s quarantined: %s", pool, reason)),
		Body: body, Priority: &prio, Pool: dest, Labels: []string{"needs-human"}, Meta: meta,
	})
	if err != nil {
		c.log.ErrorContext(ctx, "create pool escalation", "pool", pool, "err", err)
		return
	}
	c.log.InfoContext(ctx, "pool escalation filed", "pool", pool, "ticket", t.ID)
	c.ensureEscalationChat(ctx, cfg, t)
}

// clipTitle keeps the first line of s, at most 200 bytes.
func clipTitle(s string) string { return attention.ClipTitle(s) }

// unquarantinePools clears permanent quarantines: all of them, or one by
// name.
func (c *Controller) unquarantinePools(pool string) []string {
	return c.unquarantineExcept(pool, nil)
}

// unquarantineExcept is unquarantinePools keeping the pools in keep: a
// reload clears every quarantine (the config may be fixed) but the ones
// its preflight just set.
func (c *Controller) unquarantineExcept(pool string, keep map[string]string) []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []string
	for p := range c.poolQuarantined {
		if _, kept := keep[p]; !kept && (pool == "" || p == pool) {
			delete(c.poolQuarantined, p)
			delete(c.poolPausedUntil, p)
			delete(c.poolPauseReason, p)
			out = append(out, p)
		}
	}
	if pool != "" {
		delete(c.poolPausedUntil, pool)
		delete(c.poolPauseReason, pool)
		c.poolCrashLog[pool] = nil
	}
	return out
}

func (c *Controller) poolUnquarantine(ctx context.Context, sc scope, p rpc.PoolParams) (any, error) {
	if err := c.requireGrant(sc, "pool-unquarantine", "pool unquarantine"); err != nil {
		return nil, err
	}
	if _, ok := c.config().Pools[p.Pool]; !ok {
		return nil, fmt.Errorf("%w: unknown pool %q", domain.ErrInvalid, p.Pool)
	}
	c.unquarantinePools(p.Pool)
	c.log.InfoContext(ctx, "pool unquarantined", "pool", p.Pool)
	return c.poolStatus(), nil
}

// poolStatus lists pools that are quarantined or paused.
func (c *Controller) poolStatus() rpc.PoolStatusResult {
	c.mu.Lock()
	defer c.mu.Unlock()
	res := rpc.PoolStatusResult{Quarantined: map[string]string{}, Paused: map[string]time.Time{}}
	maps.Copy(res.Quarantined, c.poolQuarantined)
	now := c.now()
	for p, until := range c.poolPausedUntil {
		if now.Before(until) {
			res.Paused[p] = until
		}
	}
	return res
}

// poolShow explains a pool's merged sandbox and isolation (JOB-83).
func (c *Controller) poolShow(_ context.Context, sc scope, p rpc.PoolShowParams) (any, error) {
	if err := c.requireGrant(sc, "pool-list", "pool show"); err != nil {
		return nil, err
	}
	cfg := c.config()
	pool, ok := cfg.Pools[p.Pool]
	if !ok {
		return nil, fmt.Errorf("%w: pool %q", domain.ErrNotFound, p.Pool)
	}
	return rpc.PoolShowResult{
		Pool: p.Pool, Agent: pool.Agent, Project: pool.Project, Labels: pool.Labels,
		Sandbox: cfg.ExplainSandbox(p.Pool), Isolation: cfg.ExplainIsolation(p.Pool),
	}, nil
}

func (c *Controller) poolList(ctx context.Context, sc scope, p rpc.PoolListParams) (any, error) {
	if err := c.requireGrant(sc, "pool-list", "pool list"); err != nil {
		return nil, err
	}
	res := c.poolStatusCtx(ctx)
	if len(p.Label) == 0 {
		return res, nil
	}
	// Quarantined and paused pools not in the filter go too, so a
	// filtered view shows only its own.
	keep := map[string]bool{}
	cfg := c.config()
	for name, pool := range cfg.Pools {
		if pool.HasLabels(p.Label) {
			keep[name] = true
		}
	}
	res.Pools = slices.DeleteFunc(res.Pools, func(s rpc.PoolScale) bool { return !keep[s.Name] })
	maps.DeleteFunc(res.Quarantined, func(name, _ string) bool { return !keep[name] })
	maps.DeleteFunc(res.Paused, func(name string, _ time.Time) bool { return !keep[name] })
	return res, nil
}
