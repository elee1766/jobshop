package control

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	host "github.com/elee1766/acphost"
	"go.opentelemetry.io/otel/metric"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// Per-pool isolation (#388): a pool with isolation.user runs its agents
// as a mapped host uid/gid, one with isolation.limits in its own cgroup
// v2 with limits, and sandbox.network pasta gives a filtered network.
// acphost does the work; the controller builds the spec. With no
// isolation config nothing here runs.

// resolutionOOM ends a session whose agent the memory limit killed.
const resolutionOOM = "oom"

// cgroupSetup is the controller's delegated cgroup, set up at start when
// one is writable: the controller moves into <own>/controller and
// sessions go under <own>/sessions. err says why limits cannot be used;
// a pool with limits then fails to spawn and is quarantined.
type cgroupSetup struct {
	sessions string
	err      error
}

// setupCgroup finds the controller's cgroup and, when it is delegated
// and writable, moves the controller into a leaf child and enables
// cpu, memory and pids for sessions. Only called when some pool sets
// limits, so a host without delegation is never touched otherwise.
func setupCgroup() cgroupSetup {
	info, err := host.DelegatedCgroup()
	if err != nil {
		return cgroupSetup{err: err}
	}
	return setupCgroupAt(info, os.Getpid())
}

// cgroupControllers are what session limits need.
var cgroupControllers = []string{"cpu", "memory", "pids"}

func setupCgroupAt(info host.DelegatedInfo, pid int) cgroupSetup {
	if !info.Writable {
		return cgroupSetup{err: fmt.Errorf("cgroup %s is not delegated to this user (set Delegate= on the service)", info.Path)}
	}
	for _, want := range cgroupControllers {
		if !slices.Contains(info.Controllers, want) {
			return cgroupSetup{err: fmt.Errorf("cgroup %s: controller %s not available (cgroup.controllers: %s)", info.Path, want, strings.Join(info.Controllers, " "))}
		}
	}
	own := info.Path
	// Restarted inside our own leaf: the parent is the delegated root.
	if filepath.Base(own) == "controller" {
		own = filepath.Dir(own)
	}
	leaf := filepath.Join(own, "controller")
	sessions := filepath.Join(own, "sessions")
	for _, d := range []string{leaf, sessions} {
		if err := os.Mkdir(d, 0o755); err != nil && !errors.Is(err, os.ErrExist) {
			return cgroupSetup{err: fmt.Errorf("cgroup: %w", err)}
		}
	}
	// A cgroup with processes cannot enable controllers for children
	// (no internal processes rule): move every process of own into leaf
	// first. Only this process is expected; others (a helper started by
	// systemd) move too.
	if b, err := os.ReadFile(filepath.Join(own, "cgroup.procs")); err == nil {
		for f := range strings.FieldsSeq(string(b)) {
			_ = os.WriteFile(filepath.Join(leaf, "cgroup.procs"), []byte(f), 0o644)
		}
	}
	if err := os.WriteFile(filepath.Join(leaf, "cgroup.procs"), []byte(strconv.Itoa(pid)), 0o644); err != nil {
		return cgroupSetup{err: fmt.Errorf("cgroup: move controller: %w", err)}
	}
	enable := "+" + strings.Join(cgroupControllers, " +")
	for _, d := range []string{own, sessions} {
		if err := os.WriteFile(filepath.Join(d, "cgroup.subtree_control"), []byte(enable), 0o644); err != nil {
			return cgroupSetup{err: fmt.Errorf("cgroup: enable %s in %s: %w", enable, d, err)}
		}
	}
	return cgroupSetup{sessions: sessions}
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
func (c *Controller) ensureCgroup() cgroupSetup {
	c.cgroupOnce.Do(func() {
		c.cgroup = setupCgroup()
		if c.cgroup.err != nil {
			c.log.Warn("isolation limits unavailable", "err", c.cgroup.err)
		} else {
			c.log.Info("isolation: sessions cgroup ready", "path", c.cgroup.sessions)
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
		hu := &host.User{UID: u.UID, GID: u.GIDOrUID(), MapUID: u.MapUID, MapGID: u.MapGID}
		for _, g := range u.Groups {
			gid, err := config.LookupGroup(g)
			if err != nil {
				return host.Permanentf("isolation user: %w", err)
			}
			hu.Groups = append(hu.Groups, gid)
		}
		spec.User = hu
	}
	if l := iso.Limits; l != nil {
		cg := c.ensureCgroup()
		if cg.err != nil {
			return host.Permanentf("isolation limits: no delegated cgroup: %w", cg.err)
		}
		hc, err := cgroupFor(*l, cg.sessions)
		if err != nil {
			return host.Permanentf("isolation limits: %w", err)
		}
		spec.Cgroup = hc
	}
	return nil
}

// cgroupFor turns pool limits into an acphost cgroup under parent.
func cgroupFor(l config.CgroupLimits, parent string) (*host.Cgroup, error) {
	cpu, err := l.CPUMax()
	if err != nil {
		return nil, err
	}
	hc := &host.Cgroup{Parent: parent, CPUMax: cpu, PidsMax: l.Pids}
	if hc.MemoryMax, err = config.ParseBytes(l.Memory); err != nil {
		return nil, err
	}
	if hc.MemoryHigh, err = config.ParseBytes(l.MemoryHigh); err != nil {
		return nil, err
	}
	if hc.SwapMax, err = config.ParseBytes(l.Swap); err != nil {
		return nil, err
	}
	if hc.SwapMax == "" && hc.MemoryMax != "" && hc.MemoryMax != "max" {
		// Without this an agent at memory.max swaps instead of being
		// OOM killed, and the session never ends oom.
		hc.SwapMax = "0"
	}
	return hc, nil
}

// isolationKey is what of the isolation config feeds EnvHash, so a reload
// that changes it respawns idle sessions. Empty without isolation.
func (c *Controller) isolationKey(pool string) []string {
	iso := c.config().IsolationFor(pool)
	var key []string
	if u := iso.User; u != nil {
		key = append(key, fmt.Sprintf("JOBSHOP_ISOLATION_USER=%d:%d>%d:%d groups=%s", u.UID, u.GIDOrUID(), u.MapUID, u.MapGID, strings.Join(u.Groups, ",")))
	}
	if l := iso.Limits; l != nil {
		key = append(key, fmt.Sprintf("JOBSHOP_ISOLATION_LIMITS=cpu=%s mem=%s high=%s swap=%s pids=%d", l.CPU, l.Memory, l.MemoryHigh, l.Swap, l.Pids))
	}
	return key
}

// shareDir gives gid group ownership of dir, group rw (rwx and setgid on
// directories, so new files keep the group), and with recursive the same
// for everything under it. Unix sockets in dir are made group writable
// too. Used for pools with isolation.user (#388).
func shareDir(dir string, gid int, recursive bool) error {
	fix := func(p string, d fs.DirEntry) error {
		if err := os.Lchown(p, -1, gid); err != nil {
			return err
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		mode := info.Mode()
		perm := mode.Perm() | 0o060
		if mode.IsDir() {
			perm |= 0o010
			return os.Chmod(p, perm|fs.ModeSetgid)
		}
		return os.Chmod(p, perm)
	}
	if recursive {
		return filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			return fix(p, d)
		})
	}
	st, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if err := fix(dir, fs.FileInfoToDirEntry(st)); err != nil {
		return err
	}
	ents, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range ents {
		if e.Type()&fs.ModeSocket != 0 {
			if err := fix(filepath.Join(dir, e.Name()), e); err != nil {
				return err
			}
		}
	}
	return nil
}

// cgroupStats are the usage numbers of one session cgroup.
type cgroupStats struct {
	memory, cpuUsec, oomKills int64
}

// readCgroupStats reads memory.current, cpu.stat usage_usec and
// memory.events oom_kill from dir. Missing files read as zero.
func readCgroupStats(dir string) cgroupStats {
	var s cgroupStats
	if b, err := os.ReadFile(filepath.Join(dir, "memory.current")); err == nil {
		s.memory, _ = strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
	}
	s.cpuUsec = keyedValue(filepath.Join(dir, "cpu.stat"), "usage_usec")
	s.oomKills = keyedValue(filepath.Join(dir, "memory.events"), "oom_kill")
	return s
}

// keyedValue reads "key value" lines and returns key's value.
func keyedValue(path, key string) int64 {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		k, v, ok := strings.Cut(sc.Text(), " ")
		if ok && k == key {
			n, _ := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
			return n
		}
	}
	return 0
}

// recordCgroupMetrics sums the live session cgroups per pool. Called each
// tick; a no-op without a sessions cgroup.
func (c *Controller) recordCgroupMetrics(ctx context.Context, pools map[string][]string) {
	if c.metrics == nil || c.cgroup.sessions == "" {
		return
	}
	for pool, ids := range pools {
		if c.config().IsolationFor(pool).Limits == nil {
			continue
		}
		var sum cgroupStats
		for _, id := range ids {
			s := readCgroupStats(filepath.Join(c.cgroup.sessions, id))
			sum.memory += s.memory
			sum.cpuUsec += s.cpuUsec
			sum.oomKills += s.oomKills
		}
		attrs := metric.WithAttributes(telemetry.AttrPool.String(pool))
		c.metrics.CgroupMemory.Record(ctx, sum.memory, attrs)
		c.metrics.CgroupCPU.Record(ctx, sum.cpuUsec, attrs)
		c.metrics.CgroupOOMKills.Record(ctx, sum.oomKills, attrs)
	}
}
