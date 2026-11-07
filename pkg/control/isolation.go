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
	"syscall"

	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/elee1766/acphost/pkg/cgroup"
	"github.com/elee1766/acphost/pkg/userns"
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
//
// Session hosts start in <own>/sessions/hosts (JOB-91), so a session
// without pool limits still counts against the sessions cap, not the
// controller's. An agent with pool limits gets its own child of sessions.
// The controller's own work (hooks, git) stays in <own>/controller.
type cgroupSetup struct {
	sessions string
	hosts    string
	err      error
}

// wantsCgroup reports whether the controller should set up its delegated
// cgroup: some pool has limits, or limits.cgroup asks for the split
// (JOB-91). A host without delegation is never touched otherwise. With
// the systemd driver there is nothing to set up (#944).
func wantsCgroup(cfg *config.Config) bool {
	return !cfg.Limits.Systemd() && (cfg.Limits.Cgroup != nil || anyLimits(cfg))
}

// hostsCgroup is the cgroup directory a new session host starts in, ""
// when the split is not wanted or the cgroup is unavailable (the host then
// stays in the controller's, as before). With the systemd driver hosts
// stay in the controller's unit (KillMode=process keeps them across a
// restart); their agents get scopes in the sessions slice (poolCgroup).
func (c *Controller) hostsCgroup() string {
	if l := c.config().Limits; l.Cgroup == nil || l.Systemd() {
		return ""
	}
	return c.ensureCgroup().hosts
}

// tuneCgroupAt writes the weights and the sessions cap under own. All
// writes are tried and the errors joined: a missing weight only loses the
// reserve, so the controller logs it and runs on.
func tuneCgroupAt(own string, hc config.HostCgroup) error {
	capMax, err := hc.SessionsCPUMax()
	if err != nil {
		return err
	}
	var errs []error
	for _, w := range []struct{ file, value string }{
		{filepath.Join(own, "controller", "cpu.weight"), strconv.Itoa(hc.ControllerWeight())},
		{filepath.Join(own, "sessions", "cpu.weight"), strconv.Itoa(hc.SessionsWeight())},
		{filepath.Join(own, "sessions", "cpu.max"), capMax},
	} {
		if err := os.WriteFile(w.file, []byte(w.value), 0o644); err != nil {
			errs = append(errs, fmt.Errorf("cgroup: %w", err))
		}
	}
	return errors.Join(errs...)
}

// moveHosts moves every process in <own>/controller except pid into the
// hosts cgroup. Session hosts that outlived a controller restart (or a
// deploy that turned limits.cgroup on) sit in the controller's cgroup and
// would keep its share. Start only: later, the controller's own children
// are in there too.
func moveHosts(cg cgroupSetup, pid int) {
	leaf := filepath.Join(filepath.Dir(cg.sessions), "controller")
	b, err := os.ReadFile(filepath.Join(leaf, "cgroup.procs"))
	if err != nil {
		return
	}
	self := strconv.Itoa(pid)
	for f := range strings.FieldsSeq(string(b)) {
		if f != self {
			// A process may have exited since the read.
			_ = os.WriteFile(filepath.Join(cg.hosts, "cgroup.procs"), []byte(f), 0o644)
		}
	}
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
	if cg.err != nil {
		return
	}
	if err := tuneCgroupAt(filepath.Dir(cg.sessions), *hc); err != nil {
		c.log.Warn("isolation: sessions cgroup tuning not applied", "err", err)
	}
	if c.hostsMoved.CompareAndSwap(false, true) {
		moveHosts(cg, os.Getpid())
	}
}

// setupCgroup finds the controller's cgroup and, when it is delegated
// and writable, moves the controller into a leaf child and enables
// cpu, memory and pids for sessions. Only called when some pool sets
// limits, so a host without delegation is never touched otherwise.
func setupCgroup() cgroupSetup {
	info, err := cgroup.Delegated()
	if err != nil {
		return cgroupSetup{err: err}
	}
	return setupCgroupAt(info, os.Getpid())
}

// cgroupControllers are what session limits need.
var cgroupControllers = []string{"cpu", "memory", "pids"}

func setupCgroupAt(info cgroup.DelegatedInfo, pid int) cgroupSetup {
	if !info.Writable {
		return cgroupSetup{err: fmt.Errorf("cgroup %s is not delegated to this user (set Delegate= on the service)", info.Path)}
	}
	own := info.Path
	// Started inside our own leaf (a restart, or systemd's
	// DelegateSubgroup=controller, JOB-21): the parent is the delegated
	// root, and its controllers are the ones to check. The leaf lists only
	// what the parent enabled, which may be nothing yet.
	if filepath.Base(own) == "controller" {
		own = filepath.Dir(own)
		if b, err := os.ReadFile(filepath.Join(own, "cgroup.controllers")); err == nil {
			info.Controllers = strings.Fields(string(b))
		}
	}
	for _, want := range cgroupControllers {
		if !slices.Contains(info.Controllers, want) {
			return cgroupSetup{err: fmt.Errorf("cgroup %s: controller %s not available (cgroup.controllers: %s)", own, want, strings.Join(info.Controllers, " "))}
		}
	}
	leaf := filepath.Join(own, "controller")
	sessions := filepath.Join(own, "sessions")
	hosts := filepath.Join(sessions, "hosts")
	for _, d := range []string{leaf, sessions, hosts} {
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
	return cgroupSetup{sessions: sessions, hosts: hosts}
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
		hu := &userns.User{UID: u.UID, GID: u.GIDOrUID(), MapUID: u.MapUID, MapGID: u.MapGID}
		for _, g := range u.Groups {
			gid, err := config.LookupGroup(g)
			if err != nil {
				return host.Permanentf("isolation user: %w", err)
			}
			hu.Groups = append(hu.Groups, gid)
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
			if hc, err = cgroupFor(*l, ""); err != nil {
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
	if cg.err != nil {
		return nil, fmt.Errorf("isolation limits: no delegated cgroup: %w", cg.err)
	}
	hc, err := cgroupFor(*l, cg.sessions)
	if err != nil {
		return nil, fmt.Errorf("isolation limits: %w", err)
	}
	return hc, nil
}

// cgroupFor turns pool limits into an acphost cgroup under parent.
func cgroupFor(l config.CgroupLimits, parent string) (*cgroup.Cgroup, error) {
	cpu, err := l.CPUMax()
	if err != nil {
		return nil, err
	}
	hc := &cgroup.Cgroup{Parent: parent, CPUMax: cpu, PidsMax: l.Pids}
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

// shareDir gives gid group ownership of dir, group rw (rwx and setgid on
// directories, so new files keep the group), and with recursive the same
// for everything under it. Unix sockets in dir are made group writable
// too. Used for pools with isolation.user (#388).
func shareDir(dir string, gid int, recursive bool) error {
	fix := func(p string, d fs.DirEntry) error {
		// Files the mapped user made (a respawn after it wrote the
		// worktree) are not ours to chown; the setgid dir gave them the
		// group already.
		var st syscall.Stat_t
		if err := syscall.Lstat(p, &st); err != nil {
			return err
		}
		if int(st.Uid) != os.Geteuid() {
			return nil
		}
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

// sessionCgroupDir is session id's agent cgroup: under the sessions
// cgroup with cgroupfs, the scope the host recorded in isolation.json
// with systemd (its name holds a pid, so it is not derived). "" when
// unknown.
func (c *Controller) sessionCgroupDir(cfg *config.Config, id string) string {
	if !cfg.Limits.Systemd() {
		if c.cgroup.sessions == "" {
			return ""
		}
		return filepath.Join(c.cgroup.sessions, id)
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
	if c.metrics == nil || (c.cgroup.sessions == "" && !cfg.Limits.Systemd()) {
		return
	}
	for pool, ids := range pools {
		if cfg.IsolationFor(pool).Limits == nil {
			continue
		}
		var sum cgroupStats
		for _, id := range ids {
			dir := c.sessionCgroupDir(cfg, id)
			if dir == "" {
				continue
			}
			s := readCgroupStats(dir)
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
