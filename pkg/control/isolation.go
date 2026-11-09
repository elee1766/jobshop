package control

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/elee1766/acphost/pkg/cgroup"
	"go.opentelemetry.io/otel/metric"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/execution/sandbox"
	"github.com/elee1766/jobshop/pkg/execution/session"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// Per-pool isolation (#388): a pool with isolation.user runs its agents
// as a mapped host uid/gid, one with isolation.limits in its own cgroup
// v2 with limits, and sandbox.network pasta gives a filtered network.
// acphost does the work and pkg/execution/sandbox holds the mechanics;
// this file reads the pool's config and builds the spec. With no
// isolation config nothing here runs.

// resolutionOOM ends a session whose agent the memory limit killed.
const resolutionOOM = "oom"

// wantsCgroup reports whether the controller should set up its delegated
// cgroup: some pool has limits, or limits.cgroup asks for the split
// (JOB-91). A host without delegation is never touched otherwise. With
// the systemd driver there is nothing to set up (#944).
func wantsCgroup(cfg *config.Config) bool {
	return !cfg.Limits.Systemd() && (cfg.Limits.Cgroup != nil || anyLimits(cfg))
}

func (c *Controller) hostPlacement() session.Placement {
	if l := c.config().Limits; l.Systemd() {
		return session.Placement{Systemd: true, Slice: l.Slice()}
	}
	return session.Placement{Cgroup: c.hostsCgroup()}
}

// hostsCgroup is the cgroup directory a new session host starts in, ""
// when the split is not wanted or the cgroup is unavailable (the host then
// stays in the controller's, as before). With the systemd driver it is
// "": hosts start in a scope instead (hostPlacement).
func (c *Controller) hostsCgroup() string {
	if l := c.config().Limits; l.Cgroup == nil || l.Systemd() {
		return ""
	}
	return c.ensureCgroup().Hosts
}

// tuneCgroup applies cfg's limits.cgroup to the controller's cgroup, at
// start and on each reload. At start it also moves live hosts out of the
// controller's cgroup.
func (c *Controller) tuneCgroup(cfg *config.Config) {
	hc := cfg.Limits.Cgroup
	if hc == nil || cfg.Limits.Systemd() {
		// systemd: the cap and weights are slice and unit properties.
		return
	}
	cg := c.ensureCgroup()
	if cg.Err != nil {
		return
	}
	if err := sandbox.TuneAt(filepath.Dir(cg.Sessions), *hc); err != nil {
		c.log.Warn("isolation: sessions cgroup tuning not applied", "err", err)
	}
	if c.hostsMoved.CompareAndSwap(false, true) {
		sandbox.MoveHosts(cg, os.Getpid())
	}
}

// anyLimits reports whether some pool sets isolation limits.
func anyLimits(cfg *config.Config) bool {
	for name := range cfg.Pools {
		if cfg.IsolationFor(name).Limits != nil {
			return true
		}
	}
	return false
}

// ensureCgroup sets up the delegated cgroup once, the first time a pool
// with limits needs it (start, or a reload that adds limits).
func (c *Controller) ensureCgroup() sandbox.Delegation {
	c.cgroupOnce.Do(func() {
		c.cgroup = sandbox.Setup()
		if c.cgroup.Err != nil {
			c.log.Warn("isolation limits unavailable", "err", c.cgroup.Err)
		} else {
			c.log.Info("isolation: sessions cgroup ready", "path", c.cgroup.Sessions)
		}
	})
	return c.cgroup
}

// applyIsolation sets spec.User and spec.Cgroup for pool. A pool with
// limits on a host without a delegated cgroup gets a permanent error, so
// the spawn quarantines that pool with one escalation (#146).
func (c *Controller) applyIsolation(spec *host.Spec, pool string) error {
	iso := c.config().IsolationFor(pool)
	if u := iso.User; u != nil {
		hu, err := sandbox.UserFor(*u)
		if err != nil {
			return host.Permanentf("%w", err)
		}
		spec.User = hu
	}
	hc, err := c.poolCgroup(c.config(), pool)
	if err != nil {
		return host.Permanentf("%w", err)
	}
	spec.Cgroup = hc
	return nil
}

// poolCgroup is the acphost cgroup for pool's agents, nil when they get
// none. With cgroupfs that is a child of the controller's sessions
// cgroup when the pool sets limits. With systemd (#944) it is a scope in
// the sessions slice when the pool sets limits, and also without limits
// when limits.cgroup asks for the split, so every agent counts against
// the slice and not the controller's unit (JOB-91).
func (c *Controller) poolCgroup(cfg *config.Config, pool string) (*cgroup.Cgroup, error) {
	l := cfg.IsolationFor(pool).Limits
	if cfg.Limits.Systemd() {
		if l == nil && cfg.Limits.Cgroup == nil {
			return nil, nil
		}
		hc := &cgroup.Cgroup{}
		if l != nil {
			var err error
			if hc, err = sandbox.For(*l, ""); err != nil {
				return nil, fmt.Errorf("isolation limits: %w", err)
			}
		}
		hc.Driver, hc.Parent = cgroup.DriverSystemd, cfg.Limits.Slice()
		return hc, nil
	}
	if l == nil {
		return nil, nil
	}
	cg := c.ensureCgroup()
	if cg.Err != nil {
		return nil, fmt.Errorf("isolation limits: no delegated cgroup: %w", cg.Err)
	}
	hc, err := sandbox.For(*l, cg.Sessions)
	if err != nil {
		return nil, fmt.Errorf("isolation limits: %w", err)
	}
	return hc, nil
}

// isolationKey is what of the isolation config feeds EnvHash, so a reload
// that changes it respawns idle sessions. Empty without isolation.
func (c *Controller) isolationKey(pool string) []string {
	iso := c.config().IsolationFor(pool)
	var key []string
	if u := iso.User; u != nil {
		key = append(key, fmt.Sprintf("JOBSHOP_ISOLATION_USER=%d:%d>%d:%d groups=%s per_session=%d", u.UID, u.GIDOrUID(), u.MapUID, u.MapGID, strings.Join(u.Groups, ","), u.Slots()))
	}
	if l := iso.Limits; l != nil {
		key = append(key, fmt.Sprintf("JOBSHOP_ISOLATION_LIMITS=cpu=%s mem=%s high=%s swap=%s pids=%d", l.CPU, l.Memory, l.MemoryHigh, l.Swap, l.Pids))
	}
	// A driver change moves idle agents to the new kind of cgroup. The
	// cgroupfs default adds nothing, so existing hashes stay.
	if cfg := c.config(); cfg.Limits.Systemd() && (iso.Limits != nil || cfg.Limits.Cgroup != nil) {
		key = append(key, "JOBSHOP_ISOLATION_SCOPE="+cfg.Limits.Slice())
	}
	return key
}

// sessionCgroupDir is session id's agent cgroup: under the sessions
// cgroup with cgroupfs, the scope the host recorded in isolation.json
// with systemd (its name holds a pid, so it is not derived). "" when
// unknown.
func (c *Controller) sessionCgroupDir(cfg *config.Config, id string) string {
	if !cfg.Limits.Systemd() {
		if c.cgroup.Sessions == "" {
			return ""
		}
		return filepath.Join(c.cgroup.Sessions, id)
	}
	iso, ok, err := host.ReadIsolation(c.root.SessionDir(id))
	if !ok || err != nil || iso.Cgroup == nil {
		return ""
	}
	return iso.Cgroup.Dir
}

// recordCgroupMetrics sums the live session cgroups per pool. Called each
// tick; a no-op without a sessions cgroup.
func (c *Controller) recordCgroupMetrics(ctx context.Context, cfg *config.Config, pools map[string][]string) {
	if c.metrics == nil || (c.cgroup.Sessions == "" && !cfg.Limits.Systemd()) {
		return
	}
	for pool, ids := range pools {
		if cfg.IsolationFor(pool).Limits == nil {
			continue
		}
		var sum sandbox.Stats
		for _, id := range ids {
			dir := c.sessionCgroupDir(cfg, id)
			if dir == "" {
				continue
			}
			s := sandbox.ReadStats(dir)
			sum.Memory += s.Memory
			sum.CPUUsec += s.CPUUsec
			sum.OOMKills += s.OOMKills
		}
		attrs := metric.WithAttributes(telemetry.AttrPool.String(pool))
		c.metrics.CgroupMemory.Record(ctx, sum.Memory, attrs)
		c.metrics.CgroupCPU.Record(ctx, sum.CPUUsec, attrs)
		c.metrics.CgroupOOMKills.Record(ctx, sum.OOMKills, attrs)
	}
}
