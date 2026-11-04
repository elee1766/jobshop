package control

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	host "github.com/elee1766/acphost"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
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
	held, err := c.store.List(ctx, store.Filter{Holder: sessionID, Status: store.StatusInProgress})
	if err != nil {
		return
	}
	for _, t := range held {
		cl := store.Claim{TicketID: t.ID, Holder: sessionID, Fence: t.Fence}
		if _, err := c.store.Release(ctx, cl, crashReleaseReason, time.Time{}); err != nil {
			c.log.Warn("crash release", "ticket", t.ID, "err", err)
			continue
		}
		c.maybeQuarantine(ctx, t.ID)
		c.stepEvent(ctx, t)
	}
}

// maybeQuarantine defers a ticket for quarantineFor and labels it once
// quarantineCrashes crashed sessions have released it since it was last
// quarantined (removing the label resets the count).
func (c *Controller) maybeQuarantine(ctx context.Context, id string) {
	hist, err := c.store.History(ctx, id)
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
		c.log.Error("quarantine", "ticket", id, "err", err)
		return
	}
	if err := c.store.AddLabel(ctx, "controller", id, quarantineLabel); err != nil {
		c.log.Error("quarantine label", "ticket", id, "err", err)
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
	_, _ = c.store.AddComment(ctx, "controller", id, b.String())
	c.log.Warn("quarantined ticket", "ticket", id, "sessions", sessions)
}

// poolCrashed records a crashed session of pool and pauses the pool when
// poolCrashes happened within poolCrashWindow.
func (c *Controller) poolCrashed(pool string) {
	if pool == "" {
		return
	}
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
		c.log.Warn("pool quarantined: sessions crash-looping", "pool", pool, "crashes", len(keep), "until", now.Add(poolPause))
	}
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
	c.mu.Lock()
	_, already := c.poolQuarantined[pool]
	c.poolQuarantined[pool] = reason
	c.mu.Unlock()
	if !already {
		c.log.Warn("pool quarantined: permanent spawn error", "pool", pool, "err", reason)
	}
	c.ensurePoolEscalation(ctx, pool, reason)
}

// ensurePoolEscalation files "Pool <name> quarantined: <error>" unless an
// open one for the pool exists.
func (c *Controller) ensurePoolEscalation(ctx context.Context, pool, reason string) {
	for _, st := range []store.Status{store.StatusOpen, store.StatusInProgress} {
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
	t, err := c.store.Create(ctx, "controller", store.NewTicket{
		Project: proj, Kind: store.KindTask, Title: clipTitle(fmt.Sprintf("Pool %s quarantined: %s", pool, reason)),
		Body: body, Priority: &prio, Pool: dest, Labels: []string{"needs-human"}, Meta: meta,
	})
	if err != nil {
		c.log.Error("create pool escalation", "pool", pool, "err", err)
		return
	}
	c.log.Info("pool escalation filed", "pool", pool, "ticket", t.ID)
}

// clipTitle keeps the first line of s, at most 200 bytes.
func clipTitle(s string) string {
	s, _, _ = strings.Cut(s, "\n")
	if len(s) > 200 {
		s = s[:200]
	}
	return s
}

// unquarantinePools clears permanent quarantines: all of them on reload
// (the config may be fixed), or one by name.
func (c *Controller) unquarantinePools(pool string) []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []string
	for p := range c.poolQuarantined {
		if pool == "" || p == pool {
			delete(c.poolQuarantined, p)
			delete(c.poolPausedUntil, p)
			out = append(out, p)
		}
	}
	if pool != "" {
		delete(c.poolPausedUntil, pool)
		c.poolCrashLog[pool] = nil
	}
	return out
}

func (c *Controller) poolUnquarantine(_ context.Context, sc scope, p rpc.PoolParams) (any, error) {
	if err := c.requireGrant(sc, "pool-unquarantine", "pool unquarantine"); err != nil {
		return nil, err
	}
	if _, ok := c.config().Pools[p.Pool]; !ok {
		return nil, fmt.Errorf("%w: unknown pool %q", store.ErrInvalid, p.Pool)
	}
	c.unquarantinePools(p.Pool)
	c.log.Info("pool unquarantined", "pool", p.Pool)
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

func (c *Controller) poolList(ctx context.Context, sc scope, _ struct{}) (any, error) {
	if err := c.requireGrant(sc, "pool-list", "pool list"); err != nil {
		return nil, err
	}
	return c.poolStatusCtx(ctx), nil
}
