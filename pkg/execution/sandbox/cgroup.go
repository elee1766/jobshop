package sandbox

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/gfx-labs/acphost/pkg/cgroup"

	"github.com/gfx-labs/jobshop/pkg/config"
)

// Delegation is the controller's delegated cgroup, set up at start when
// one is writable: the controller moves into <own>/controller and
// sessions go under <own>/sessions. Err says why limits cannot be used;
// a pool with limits then fails to spawn and is quarantined.
//
// Session hosts start in <own>/sessions/hosts (JOB-91), so a session
// without pool limits still counts against the sessions cap, not the
// controller's. An agent with pool limits gets its own child of sessions.
// The controller's own work (hooks, git) stays in <own>/controller.
type Delegation struct {
	Sessions string
	Hosts    string
	Err      error
}

// controllers are what session limits need.
var controllers = []string{"cpu", "memory", "pids"}

// Setup finds the controller's cgroup and, when it is delegated and
// writable, moves the controller into a leaf child and enables cpu,
// memory and pids for sessions. Only called when some pool sets limits
// or limits.cgroup asks for the split, so a host without delegation is
// never touched otherwise.
func Setup() Delegation {
	info, err := cgroup.Delegated()
	if err != nil {
		return Delegation{Err: err}
	}
	return SetupAt(info, os.Getpid())
}

// SetupAt is Setup for the delegated cgroup in info and the controller's
// pid.
func SetupAt(info cgroup.DelegatedInfo, pid int) Delegation {
	if !info.Writable {
		return Delegation{Err: fmt.Errorf("cgroup %s is not delegated to this user (set Delegate= on the service)", info.Path)}
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
	for _, want := range controllers {
		if !slices.Contains(info.Controllers, want) {
			return Delegation{Err: fmt.Errorf("cgroup %s: controller %s not available (cgroup.controllers: %s)", own, want, strings.Join(info.Controllers, " "))}
		}
	}
	leaf := filepath.Join(own, "controller")
	sessions := filepath.Join(own, "sessions")
	hosts := filepath.Join(sessions, "hosts")
	for _, d := range []string{leaf, sessions, hosts} {
		if err := os.Mkdir(d, 0o755); err != nil && !errors.Is(err, os.ErrExist) {
			return Delegation{Err: fmt.Errorf("cgroup: %w", err)}
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
		return Delegation{Err: fmt.Errorf("cgroup: move controller: %w", err)}
	}
	enable := "+" + strings.Join(controllers, " +")
	for _, d := range []string{own, sessions} {
		if err := os.WriteFile(filepath.Join(d, "cgroup.subtree_control"), []byte(enable), 0o644); err != nil {
			return Delegation{Err: fmt.Errorf("cgroup: enable %s in %s: %w", enable, d, err)}
		}
	}
	return Delegation{Sessions: sessions, Hosts: hosts}
}

// TuneAt writes the weights and the sessions cap under own. All writes
// are tried and the errors joined: a missing weight only loses the
// reserve, so the controller logs it and runs on.
func TuneAt(own string, hc config.HostCgroup) error {
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

// MoveHosts moves every process in <own>/controller except pid into the
// hosts cgroup. Session hosts that outlived a controller restart (or a
// deploy that turned limits.cgroup on) sit in the controller's cgroup and
// would keep its share. Start only: later, the controller's own children
// are in there too.
func MoveHosts(d Delegation, pid int) {
	leaf := filepath.Join(filepath.Dir(d.Sessions), "controller")
	b, err := os.ReadFile(filepath.Join(leaf, "cgroup.procs"))
	if err != nil {
		return
	}
	self := strconv.Itoa(pid)
	for f := range strings.FieldsSeq(string(b)) {
		if f != self {
			// A process may have exited since the read.
			_ = os.WriteFile(filepath.Join(d.Hosts, "cgroup.procs"), []byte(f), 0o644)
		}
	}
}

// For turns pool limits into an acphost cgroup under parent.
func For(l config.CgroupLimits, parent string) (*cgroup.Cgroup, error) {
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

// Stats are the usage numbers of one session cgroup.
type Stats struct {
	Memory, CPUUsec, OOMKills int64
}

// ReadStats reads memory.current, cpu.stat usage_usec and memory.events
// oom_kill from dir. Missing files read as zero.
func ReadStats(dir string) Stats {
	var s Stats
	if b, err := os.ReadFile(filepath.Join(dir, "memory.current")); err == nil {
		s.Memory, _ = strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
	}
	s.CPUUsec = keyedValue(filepath.Join(dir, "cpu.stat"), "usage_usec")
	s.OOMKills = keyedValue(filepath.Join(dir, "memory.events"), "oom_kill")
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
