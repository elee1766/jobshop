package control

import (
	"cmp"
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

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// Pool scaling (#281). Each tick the supervisor computes a desired session
// count per ticket pool and reconciles toward it: it starts sessions up to
// desired, and cordons the surplus instead of stopping it. A cordoned
// session claims nothing new, finishes what it holds and then ends.

const (
	// scaleFile persists manual scale targets and pool drains in the root,
	// so a controller restart keeps them.
	scaleFile             = "scale.json"
	defaultScaleDownAfter = 10 * time.Minute
	resolutionScaledDown  = "scaled_down"
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

// scaleSample is one raw recommendation, kept for the stabilization window.
type scaleSample struct {
	at time.Time
	n  int
}

// autoDesired is clamp(min, max, demand) held at its highest value over
// the scale-down window, so a burst does not start and stop sessions in a
// loop. Scale up is immediate. hist is the pool's samples, pruned and
// appended in place.
func autoDesired(p config.Pool, demand int, now time.Time, hist []scaleSample) (int, []scaleSample) {
	rec := min(max(demand, p.MinSessions), p.MaxSessions)
	window := config.Duration(p.ScaleDownAfter, defaultScaleDownAfter)
	hist = slices.DeleteFunc(hist, func(s scaleSample) bool { return now.Sub(s.at) >= window })
	hist = append(hist, scaleSample{at: now, n: rec})
	out := 0
	for _, s := range hist {
		out = max(out, s.n)
	}
	// A reload that lowers max_sessions applies at once, not after the window.
	return min(out, p.MaxSessions), hist
}

// scaleSession is the reconciler's view of one open session.
type scaleSession struct {
	ID        string
	Idle      bool // holds no ticket
	LastClaim time.Time
	Created   time.Time
	Cordon    string // store.Cordon reason, "" when in rotation
}

// scalePlan is what one pool needs this tick.
type scalePlan struct {
	Cordon   []string // move to CordonScale
	Uncordon []string // scale cordons to lift
	Start    int      // new sessions wanted, before brakes and the host cap
}

// planPool compares desired with the sessions in rotation. Surplus is
// cordoned idle first, then least recent claim, then oldest host. A
// shortfall first lifts scale cordons (the session is already warm) and
// starts the rest. Manual and drain cordons are never lifted here.
func planPool(desired int, sessions []scaleSession) scalePlan {
	var active, scaled []scaleSession
	for _, s := range sessions {
		switch s.Cordon {
		case "":
			active = append(active, s)
		case store.CordonScale:
			scaled = append(scaled, s)
		}
	}
	var p scalePlan
	if n := len(active) - desired; n > 0 {
		slices.SortStableFunc(active, func(a, b scaleSession) int {
			if a.Idle != b.Idle {
				if a.Idle {
					return -1
				}
				return 1
			}
			return cmp.Or(a.LastClaim.Compare(b.LastClaim), a.Created.Compare(b.Created), cmp.Compare(a.ID, b.ID))
		})
		for _, s := range active[:n] {
			p.Cordon = append(p.Cordon, s.ID)
		}
		return p
	}
	short := desired - len(active)
	// Lift the busiest first: a working session is the furthest from ending.
	slices.SortStableFunc(scaled, func(a, b scaleSession) int {
		if a.Idle != b.Idle {
			if a.Idle {
				return 1
			}
			return -1
		}
		return b.LastClaim.Compare(a.LastClaim)
	})
	for _, s := range scaled {
		if short == 0 {
			break
		}
		p.Uncordon = append(p.Uncordon, s.ID)
		short--
	}
	p.Start = short
	return p
}

// capRequest is one pool's wish to start sessions under the host cap.
type capRequest struct {
	Pool     string
	Priority int
	Start    int
}

// capStarts grants starts in priority order (lower first, then name) until
// open plus granted reaches limit. limit 0 means no cap.
func capStarts(reqs []capRequest, open, limit int) map[string]int {
	out := map[string]int{}
	slices.SortStableFunc(reqs, func(a, b capRequest) int {
		return cmp.Or(cmp.Compare(a.Priority, b.Priority), cmp.Compare(a.Pool, b.Pool))
	})
	room := limit - open
	for _, r := range reqs {
		n := r.Start
		if limit > 0 {
			n = max(min(n, room), 0)
			room -= n
		}
		if n > 0 {
			out[r.Pool] = n
		}
	}
	return out
}

// poolScaleView is one pool's state from the last reconcile, for pool list.
type poolScaleView struct {
	desired int
}

// reconcile is the desired-state loop: per ticket pool, compute desired,
// cordon or uncordon, then start sessions within the brakes and host cap.
func (c *Controller) reconcile() {
	ctx := c.ctx
	cfg := c.config()
	now := c.now()
	backingOff := c.backingOffPools()
	var reqs []capRequest
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
		ready, err := c.store.Ready(ctx, p.Project, name, p.MaxSessions)
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
		plan := planPool(desired, sessions)
		for _, id := range plan.Cordon {
			if err := c.setCordon(ctx, "controller", id, &store.Cordon{Reason: store.CordonScale, By: "controller", At: now}); err != nil {
				c.log.Warn("scale: cordon", "session", id, "err", err)
				continue
			}
			c.scaleEvent(ctx, name, "cordon", id, desired)
		}
		for _, id := range plan.Uncordon {
			if err := c.setCordon(ctx, "controller", id, nil); err != nil {
				c.log.Warn("scale: uncordon", "session", id, "err", err)
				continue
			}
			c.scaleEvent(ctx, name, "uncordon", id, desired)
		}
		// Backoff (rate limit), a crash-loop pause and spawn quarantine
		// block scale up only: they never cordon.
		if plan.Start > 0 && !backingOff[name] && !c.poolPaused(name) {
			start := plan.Start
			if free := c.slotCapacity(ctx, name); free >= 0 {
				// A full per-session id range is a capacity limit (#540).
				start = min(start, free)
			}
			if start > 0 {
				reqs = append(reqs, capRequest{Pool: name, Priority: p.Priority, Start: start})
			}
		}
	}
	grants := capStarts(reqs, open, cfg.Limits.MaxSessions)
	for _, name := range slices.Sorted(maps.Keys(grants)) {
		p := cfg.Pools[name]
		for range grants[name] {
			if _, err := c.newSession(ctx, "controller", p.Project, name, p.Agent, "", "", ""); err != nil {
				c.log.Error("start session", "pool", name, "err", err)
				break
			}
			c.scaleEvent(ctx, name, "up", "", c.poolScale(name).desired)
		}
	}
}

// scaleSessions lists a pool's open sessions and how many hold a ticket.
// Idle is read before the caller reads ready: a claim in between then
// drops the ticket from ready instead of counting it twice (#102).
func (c *Controller) scaleSessions(ctx context.Context, pool string) ([]scaleSession, int, error) {
	open, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Status: store.StatusOpen, Pool: pool})
	if err != nil {
		return nil, 0, err
	}
	out := make([]scaleSession, 0, len(open))
	working := 0
	for _, t := range open {
		s := scaleSession{ID: t.ID, Created: t.CreatedAt, LastClaim: t.CreatedAt}
		if f, err := store.DecodeFields[store.SessionFields](t); err == nil && f.Cordoned != nil {
			s.Cordon = f.Cordoned.Reason
		}
		held, err := c.store.List(ctx, store.Filter{Holder: t.ID, Status: store.StatusInProgress, Limit: 1})
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
		c.scaleHist[name] = []scaleSample{{at: now, n: min(open, p.MaxSessions)}}
	}
	var d int
	target, manual := c.scaleTargets[name]
	switch {
	case c.drained[drainAll] || c.drained[name]:
		d = 0
	case manual:
		d = min(max(target, 0), p.MaxSessions)
	default:
		d, c.scaleHist[name] = autoDesired(p, demand, now, c.scaleHist[name])
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
func (c *Controller) setCordon(ctx context.Context, actor, id string, cd *store.Cordon) error {
	for range 5 {
		t, err := c.store.Get(ctx, id)
		if err != nil {
			return err
		}
		f, err := store.DecodeFields[store.SessionFields](t)
		if err != nil {
			return err
		}
		f.Cordoned = cd
		b, err := json.Marshal(f)
		if err != nil {
			return err
		}
		_, err = c.store.Update(ctx, actor, id, store.Update{Fields: b, IfRevision: t.Revision})
		if !errors.Is(err, store.ErrConflict) {
			if err == nil {
				c.wakeRunner(id)
			}
			return err
		}
	}
	return fmt.Errorf("%w: session %s kept changing", store.ErrConflict, id)
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
func (c *Controller) sessionCordon(ctx context.Context, id string) *store.Cordon {
	t, err := c.store.Get(ctx, id)
	if err != nil {
		return nil
	}
	f, err := store.DecodeFields[store.SessionFields](t)
	if err != nil {
		return nil
	}
	return f.Cordoned
}

// scaleEvent counts a scale change and publishes it on events.subscribe.
func (c *Controller) scaleEvent(ctx context.Context, pool, op, session string, desired int) {
	c.metrics.ScaleEvents.Add(ctx, 1, metric.WithAttributes(telemetry.AttrPool.String(pool), telemetry.AttrReason.String(op)))
	c.log.Info("pool scale", "pool", pool, "op", op, "session", session, "desired", desired)
	c.events.publish([]rpc.Event{{Kind: rpc.EventKindPool, Op: op, Pool: pool, ID: session, Desired: desired, Time: c.now()}})
}

// poolScaleRPC sets or clears a pool's manual target (pool scale).
func (c *Controller) poolScaleRPC(ctx context.Context, sc scope, p rpc.PoolScaleParams) (any, error) {
	if err := c.requireGrant(sc, "pool-scale", "pool scale"); err != nil {
		return nil, err
	}
	pool, ok := c.config().Pools[p.Pool]
	if !ok {
		return nil, fmt.Errorf("%w: unknown pool %q", store.ErrInvalid, p.Pool)
	}
	if pool.Chat() {
		return nil, fmt.Errorf("%w: pool %s is a chat pool; it does not scale", store.ErrInvalid, p.Pool)
	}
	if !p.Auto && (p.Target == nil || *p.Target < 0) {
		return nil, fmt.Errorf("%w: give a target >= 0 or auto", store.ErrInvalid)
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
	if t.Kind != store.KindSession {
		return nil, fmt.Errorf("%w: %s is a %s, not a session", store.ErrInvalid, p.ID, t.Kind)
	}
	if t.Status == store.StatusClosed {
		return nil, fmt.Errorf("%w: session %s", store.ErrClosed, p.ID)
	}
	var cd *store.Cordon
	op := "uncordon"
	if !p.Uncordon {
		reason := store.CordonManual
		if p.Drain {
			reason = store.CordonDrain
		}
		cd = &store.Cordon{Reason: reason, By: sc.actor, At: c.now()}
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
			Name: name, Min: p.MinSessions, Max: p.MaxSessions, Priority: p.Priority,
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
	return res
}

// checkAssign refuses to pre-assign work to a cordoned session unless
// forced: it would sit unclaimed, or keep a scaled-down session alive.
func (c *Controller) checkAssign(ctx context.Context, holder string, force bool) error {
	if holder == "" || force {
		return nil
	}
	if cd := c.sessionCordon(ctx, holder); cd != nil {
		return fmt.Errorf("%w: session %s is cordoned (%s); pass --force to assign anyway", store.ErrConflict, holder, cd.Reason)
	}
	return nil
}

// cordonedWork is what a cordoned session may still take: its in-progress
// ticket, then a ticket pre-assigned to it (only with --force). Zero id
// means nothing.
func (c *Controller) cordonedWork(ctx context.Context, id string) string {
	for _, st := range []store.Status{store.StatusInProgress, store.StatusOpen} {
		held, err := c.store.List(ctx, store.Filter{Holder: id, Status: st, Limit: 1})
		if err == nil && len(held) > 0 {
			return held[0].ID
		}
	}
	return ""
}
