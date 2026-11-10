package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"time"

	"go.opentelemetry.io/otel/metric"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/orchestration/scaleplan"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/support/telemetry"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// Pool scaling (#281). Each tick the supervisor computes a desired session
// count per ticket pool and reconciles toward it: it starts sessions up to
// desired, and cordons the surplus instead of stopping it. A cordoned
// session claims nothing new, finishes what it holds and then ends.

const (
	// scaleFile persists manual scale targets and pool drains in the root,
	// so a controller restart keeps them.
	scaleFile            = "scale.json"
	resolutionScaledDown = "scaled_down"
)

// scaleState is scaleFile.
type scaleState struct {
	Targets map[string]int `json:"targets,omitempty"`
	Drained []string       `json:"drained,omitempty"`
}

// loadScale reads scaleFile into c.scaleTargets and c.drained.
func (c *Controller) loadScale() {
	path := filepath.Join(c.root.Dir, scaleFile)
	b, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			c.log.Warn("scale: state unreadable, starting with none", "file", path, "err", err)
		}
		return
	}
	var st scaleState
	if err := json.Unmarshal(b, &st); err != nil {
		c.log.Warn("scale: state unreadable, starting with none", "file", path, "err", err)
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	maps.Copy(c.scaleTargets, st.Targets)
	for _, p := range st.Drained {
		c.drained[p] = true
	}
}

// saveScaleLocked writes scaleFile. Caller holds c.mu.
func (c *Controller) saveScaleLocked() error {
	st := scaleState{Targets: c.scaleTargets, Drained: slices.Sorted(maps.Keys(c.drained))}
	b, err := json.Marshal(st)
	if err != nil {
		return err
	}
	path := filepath.Join(c.root.Dir, scaleFile)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// poolScaleView is one pool's state from the last reconcile, for pool list.
type poolScaleView struct {
	desired int
}

// reconcile is the desired-state loop: per ticket pool, compute desired,
// cordon or uncordon, then start sessions within the brakes and host cap.
func (c *Controller) reconcile(ctx context.Context, cfg *config.Config) {
	now := c.now()
	backingOff := c.backingOffPools()
	var reqs []scaleplan.CapRequest
	open := 0
	for _, name := range slices.Sorted(maps.Keys(cfg.Pools)) {
		p := cfg.Pools[name]
		if p.Chat() {
			continue // chat sessions are started by session.send
		}
		sessions, working, err := c.scaleSessions(ctx, name)
		if err != nil {
			continue
		}
		open += len(sessions)
		if p.MaxSessions == 0 {
			continue // manual sessions only
		}
		ready, err := c.store.List(ctx, store.Filter{Project: p.Project, Pool: name, Ready: true, Limit: p.MaxSessions})
		if err != nil {
			continue
		}
		desired := c.desired(name, p, len(ready)+working, len(sessions), now)
		c.metrics.PoolDesired.Record(ctx, int64(desired), metric.WithAttributes(telemetry.AttrPool.String(name)))
		cordoned := 0
		for _, s := range sessions {
			if s.Cordon != "" {
				cordoned++
			}
		}
		c.metrics.PoolCordoned.Record(ctx, int64(cordoned), metric.WithAttributes(telemetry.AttrPool.String(name)))
		if c.poolDrained(name) {
			continue // the drain path ends its sessions as drained
		}
		plan := scaleplan.ForPool(desired, sessions)
		for _, id := range plan.Cordon {
			if err := c.setCordon(ctx, "controller", id, &domain.Cordon{Reason: domain.CordonScale, By: "controller", At: now}); err != nil {
				c.log.WarnContext(ctx, "scale: cordon", "session", id, "err", err)
				continue
			}
			c.scaleEvent(ctx, name, "cordon", id, desired)
		}
		for _, id := range plan.Uncordon {
			if err := c.setCordon(ctx, "controller", id, nil); err != nil {
				c.log.WarnContext(ctx, "scale: uncordon", "session", id, "err", err)
				continue
			}
			c.scaleEvent(ctx, name, "uncordon", id, desired)
		}
		// Backoff (rate limit), a crash-loop pause and spawn quarantine
		// block scale up only: they never cordon.
		if plan.Start > 0 && !backingOff[name] && !c.poolPaused(name) {
			start := plan.Start
			if free := c.slotCapacity(ctx, cfg, name); free >= 0 {
				// A full per-session id range is a capacity limit (#540).
				start = min(start, free)
			}
			if start > 0 {
				reqs = append(reqs, scaleplan.CapRequest{Pool: name, Priority: p.Priority, Start: start})
			}
		}
	}
	grants := scaleplan.CapStarts(reqs, open, cfg.Limits.MaxSessions)
	for _, name := range slices.Sorted(maps.Keys(grants)) {
		p := cfg.Pools[name]
		for range grants[name] {
			if _, err := c.newSession(ctx, cfg, "controller", p.Project, name, p.Agent, "", "", ""); err != nil {
				c.log.ErrorContext(ctx, "start session", "pool", name, "err", err)
				break
			}
			c.scaleEvent(ctx, name, "up", "", c.poolScale(name).desired)
		}
	}
}

// scaleSessions lists a pool's open sessions and how many hold a ticket.
// Idle is read before the caller reads ready: a claim in between then
// drops the ticket from ready instead of counting it twice (#102).
func (c *Controller) scaleSessions(ctx context.Context, pool string) ([]scaleplan.Session, int, error) {
	open, err := c.store.List(ctx, store.Filter{Kind: domain.KindSession, Status: domain.StatusOpen, Pool: pool})
	if err != nil {
		return nil, 0, err
	}
	out := make([]scaleplan.Session, 0, len(open))
	working := 0
	for _, t := range open {
		s := scaleplan.Session{ID: t.ID, Created: t.CreatedAt, LastClaim: t.CreatedAt}
		if f, err := domain.DecodeFields[domain.SessionFields](t); err == nil && f.Cordoned != nil {
			s.Cordon = f.Cordoned.Reason
		}
		held, err := c.store.List(ctx, store.Filter{Holder: t.ID, Status: domain.StatusInProgress, Limit: 1})
		s.Idle = err == nil && len(held) == 0
		if !s.Idle {
			working++
		}
		c.mu.Lock()
		if r := c.runners[t.ID]; r != nil {
			if lc := r.lastClaimAt(); !lc.IsZero() {
				s.LastClaim = lc
			}
		}
		c.mu.Unlock()
		out = append(out, s)
	}
	return out, working, nil
}

// desired is the pool's target this tick: 0 when drained, the manual
// target (capped by max_sessions) when set, else autoDesired.
//
// The first tick for a pool (a controller start or a new pool) seeds the
// window with the open count, so adopted sessions get the same
// scale_down_after grace as after a burst.
func (c *Controller) desired(name string, p config.Pool, demand, open int, now time.Time) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, seen := c.scaleHist[name]; !seen {
		c.scaleHist[name] = []scaleplan.Sample{{At: now, N: min(open, p.MaxSessions)}}
	}
	var d int
	target, manual := c.scaleTargets[name]
	switch {
	case c.drained[drainAll] || c.drained[name]:
		d = 0
	case manual:
		d = min(max(target, 0), p.MaxSessions)
	default:
		d, c.scaleHist[name] = scaleplan.Desired(p, demand, now, c.scaleHist[name])
	}
	c.scaleViews[name] = poolScaleView{desired: d}
	return d
}

func (c *Controller) poolScale(name string) poolScaleView {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.scaleViews[name]
}

// warmFloor is how many open sessions of pool are kept past idle_timeout:
// the manual target when set, else min_sessions. 0 when drained.
func (c *Controller) warmFloor(pool string) int {
	p, ok := c.config().Pools[pool]
	if !ok {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.drained[drainAll] || c.drained[pool] {
		return 0
	}
	if t, ok := c.scaleTargets[pool]; ok {
		return min(max(t, 0), p.MaxSessions)
	}
	return p.MinSessions
}

// keepWarm reports whether an idle session of pool should outlive its
// idle timeout: the pool's sessions in rotation are at or below the floor.
func (c *Controller) keepWarm(ctx context.Context, pool string) bool {
	floor := c.warmFloor(pool)
	if floor == 0 {
		return false
	}
	sessions, _, err := c.scaleSessions(ctx, pool)
	if err != nil {
		return false
	}
	active := 0
	for _, s := range sessions {
		if s.Cordon == "" {
			active++
		}
	}
	return active <= floor
}

// setCordon writes a session's cordon (nil lifts it) with a revision check,
// so a concurrent fields write by the runner is not lost.
func (c *Controller) setCordon(ctx context.Context, actor, id string, cd *domain.Cordon) error {
	for range 5 {
		t, err := c.store.Get(ctx, id)
		if err != nil {
			return err
		}
		f, err := domain.DecodeFields[domain.SessionFields](t)
		if err != nil {
			return err
		}
		f.Cordoned = cd
		b, err := json.Marshal(f)
		if err != nil {
			return err
		}
		_, err = c.store.Update(ctx, actor, id, store.Update{Fields: b, IfRevision: t.Revision})
		if !errors.Is(err, domain.ErrConflict) {
			if err == nil {
				c.wakeRunner(id)
			}
			return err
		}
	}
	return fmt.Errorf("%w: session %s kept changing", domain.ErrConflict, id)
}

// wakeRunner ends a runner's idle wait so it sees a cordon change, or a
// new message (#603), now instead of at its next poll.
func (c *Controller) wakeRunner(id string) {
	c.mu.Lock()
	r := c.runners[id]
	c.mu.Unlock()
	if r == nil {
		return
	}
	select {
	case r.restartWake <- struct{}{}:
	default:
	}
}

// sessionCordon reads a session's cordon from the store; nil when in
// rotation or unreadable.
func (c *Controller) sessionCordon(ctx context.Context, id string) *domain.Cordon {
	t, err := c.store.Get(ctx, id)
	if err != nil {
		return nil
	}
	f, err := domain.DecodeFields[domain.SessionFields](t)
	if err != nil {
		return nil
	}
	return f.Cordoned
}

// scaleEvent counts a scale change and publishes it on events.subscribe.
func (c *Controller) scaleEvent(ctx context.Context, pool, op, session string, desired int) {
	c.metrics.ScaleEvents.Add(ctx, 1, metric.WithAttributes(telemetry.AttrPool.String(pool), telemetry.AttrReason.String(op)))
	c.log.InfoContext(ctx, "pool scale", "pool", pool, "op", op, "session", session, "desired", desired)
	c.events.publish([]rpc.Event{{Kind: rpc.EventKindPool, Op: op, Pool: pool, ID: session, Desired: desired, Time: c.now()}})
}

// poolScaleRPC sets or clears a pool's manual target (pool scale).
func (c *Controller) poolScaleRPC(ctx context.Context, sc scope, p rpc.PoolScaleParams) (any, error) {
	if err := c.requireGrant(sc, "pool-scale", "pool scale"); err != nil {
		return nil, err
	}
	pool, ok := c.config().Pools[p.Pool]
	if !ok {
		return nil, fmt.Errorf("%w: unknown pool %q", domain.ErrInvalid, p.Pool)
	}
	if pool.Chat() {
		return nil, fmt.Errorf("%w: pool %s is a chat pool; it does not scale", domain.ErrInvalid, p.Pool)
	}
	if !p.Auto && (p.Target == nil || *p.Target < 0) {
		return nil, fmt.Errorf("%w: give a target >= 0 or auto", domain.ErrInvalid)
	}
	c.mu.Lock()
	if p.Auto {
		delete(c.scaleTargets, p.Pool)
	} else {
		c.scaleTargets[p.Pool] = *p.Target
	}
	err := c.saveScaleLocked()
	c.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("save %s: %w", scaleFile, err)
	}
	op, d := "manual", 0
	if p.Auto {
		op = "auto"
	} else {
		d = min(*p.Target, pool.MaxSessions)
	}
	c.scaleEvent(ctx, p.Pool, op, "", d)
	return c.poolStatusCtx(ctx), nil
}

// sessionCordonRPC cordons, uncordons or drains one session.
func (c *Controller) sessionCordonRPC(ctx context.Context, sc scope, p rpc.SessionCordonParams) (any, error) {
	if err := c.requireGrant(sc, "session-cordon", "session cordon"); err != nil {
		return nil, err
	}
	t, err := c.store.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if t.Kind != domain.KindSession {
		return nil, fmt.Errorf("%w: %s is a %s, not a session", domain.ErrInvalid, p.ID, t.Kind)
	}
	if t.Status == domain.StatusClosed {
		return nil, fmt.Errorf("%w: session %s", domain.ErrClosed, p.ID)
	}
	var cd *domain.Cordon
	op := "uncordon"
	if !p.Uncordon {
		reason := domain.CordonManual
		if p.Drain {
			reason = domain.CordonDrain
		}
		cd = &domain.Cordon{Reason: reason, By: sc.actor, At: c.now()}
		op = "cordon"
	}
	if err := c.setCordon(ctx, sc.actor, p.ID, cd); err != nil {
		return nil, err
	}
	c.scaleEvent(ctx, t.Pool, op, p.ID, c.poolScale(t.Pool).desired)
	return c.sessionInfo(ctx, p.ID)
}

// poolStatusCtx is poolStatus plus each ticket pool's scale columns.
func (c *Controller) poolStatusCtx(ctx context.Context) rpc.PoolStatusResult {
	res := c.poolStatus()
	cfg := c.config()
	for _, name := range slices.Sorted(maps.Keys(cfg.Pools)) {
		p := cfg.Pools[name]
		if p.Chat() {
			continue
		}
		row := rpc.PoolScale{
			Name: name, Project: p.Project, Labels: p.Labels, Min: p.MinSessions, Max: p.MaxSessions, Priority: p.Priority,
			Desired: c.poolScale(name).desired, Drained: c.poolDrained(name),
		}
		c.mu.Lock()
		if t, ok := c.scaleTargets[name]; ok {
			row.Target = &t
		}
		c.mu.Unlock()
		if sessions, working, err := c.scaleSessions(ctx, name); err == nil {
			row.Open, row.Working = len(sessions), working
			for _, s := range sessions {
				if s.Cordon != "" {
					row.Cordoned++
				} else if s.Idle {
					row.Idle++
				}
			}
		}
		res.Pools = append(res.Pools, row)
	}
	res.HostMax = cfg.Limits.MaxSessions
	res.Disabled = cfg.Disabled.Pools
	return res
}

// checkAssign refuses to pre-assign work to a cordoned session unless
// forced: it would sit unclaimed, or keep a scaled-down session alive.
func (c *Controller) checkAssign(ctx context.Context, holder string, force bool) error {
	if holder == "" || force {
		return nil
	}
	if cd := c.sessionCordon(ctx, holder); cd != nil {
		return fmt.Errorf("%w: session %s is cordoned (%s); pass --force to assign anyway", domain.ErrConflict, holder, cd.Reason)
	}
	return nil
}

// cordonedWork is what a cordoned session may still take: its in-progress
// ticket, then a ticket pre-assigned to it (only with --force). Zero id
// means nothing.
func (c *Controller) cordonedWork(ctx context.Context, id string) string {
	for _, st := range []domain.Status{domain.StatusInProgress, domain.StatusOpen} {
		held, err := c.store.List(ctx, store.Filter{Holder: id, Status: st, Limit: 1})
		if err == nil && len(held) > 0 {
			return held[0].ID
		}
	}
	return ""
}
