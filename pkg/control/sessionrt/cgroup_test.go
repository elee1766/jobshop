package sessionrt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/elee1766/acphost/pkg/cgroup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
)

// TestSetupCgroupAt runs the setup on a plain directory standing in for
// the delegated cgroup: leaf and sessions children are made, the pid is
// written to the leaf, and controllers are enabled on both levels.
func TestSetupCgroupAt(t *testing.T) {
	own := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(own, "cgroup.procs"), []byte("42\n"), 0o600))
	info := cgroup.DelegatedInfo{Path: own, Writable: true, Controllers: []string{"cpu", "io", "memory", "pids"}}
	cg := SetupAt(info, 7)
	require.NoError(t, cg.Err)
	assert.Equal(t, filepath.Join(own, "sessions"), cg.Sessions)
	assert.Equal(t, filepath.Join(own, "sessions", "hosts"), cg.Hosts)
	b, err := os.ReadFile(filepath.Join(own, "controller", "cgroup.procs"))
	require.NoError(t, err)
	assert.Equal(t, "7", string(b), "the controller moved into its leaf")
	for _, d := range []string{own, cg.Sessions} {
		b, err := os.ReadFile(filepath.Join(d, "cgroup.subtree_control"))
		require.NoError(t, err)
		assert.Equal(t, "+cpu +memory +pids", string(b))
	}

	// Restarted inside the leaf: the delegated root is its parent.
	again := SetupAt(cgroup.DelegatedInfo{Path: filepath.Join(own, "controller"), Writable: true, Controllers: info.Controllers}, 8)
	require.NoError(t, again.Err)
	assert.Equal(t, cg.Sessions, again.Sessions)

	// Started by systemd in the leaf (DelegateSubgroup=controller, JOB-21)
	// before anything is enabled: the leaf lists no controllers, the
	// parent's are what count.
	fresh := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(fresh, "cgroup.controllers"), []byte("cpu io memory pids\n"), 0o600))
	require.NoError(t, os.Mkdir(filepath.Join(fresh, "controller"), 0o755))
	sub := SetupAt(cgroup.DelegatedInfo{Path: filepath.Join(fresh, "controller"), Writable: true}, 9)
	require.NoError(t, sub.Err)
	assert.Equal(t, filepath.Join(fresh, "sessions"), sub.Sessions)

	assert.ErrorContains(t, SetupAt(cgroup.DelegatedInfo{Path: own}, 7).Err, "not delegated")
	assert.ErrorContains(t, SetupAt(cgroup.DelegatedInfo{Path: own, Writable: true, Controllers: []string{"cpu"}}, 7).Err, "memory")
}

func TestCgroupFor(t *testing.T) {
	hc, err := For(config.CgroupLimits{CPU: "4", Memory: "8G", MemoryHigh: "6G", Pids: 4096}, "/sys/fs/cgroup/x/sessions")
	require.NoError(t, err)
	assert.Equal(t, cgroup.Cgroup{
		Parent: "/sys/fs/cgroup/x/sessions", CPUMax: "400000 100000",
		MemoryMax: "8589934592", MemoryHigh: "6442450944", SwapMax: "0", PidsMax: 4096,
	}, *hc)
	hc, err = For(config.CgroupLimits{Memory: "1G", Swap: "max"}, "/p")
	require.NoError(t, err)
	assert.Equal(t, "max", hc.SwapMax, "an explicit swap is kept")
	hc, err = For(config.CgroupLimits{Pids: 10}, "/p")
	require.NoError(t, err)
	assert.Empty(t, hc.SwapMax, "no memory limit, no swap limit")
}

func TestShareDir(t *testing.T) {
	d := t.TempDir()
	sub := filepath.Join(d, "a")
	require.NoError(t, os.MkdirAll(sub, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(sub, "f"), []byte("x"), 0o600))
	gid := os.Getegid()
	require.NoError(t, ShareDir(d, gid, true))
	st, err := os.Stat(sub)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o770)|os.ModeSetgid|os.ModeDir, st.Mode())
	st, err = os.Stat(filepath.Join(sub, "f"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o660), st.Mode())
}

func TestReadCgroupStats(t *testing.T) {
	d := t.TempDir()
	write := func(f, s string) { require.NoError(t, os.WriteFile(filepath.Join(d, f), []byte(s), 0o600)) }
	write("memory.current", "1234\n")
	write("cpu.stat", "usage_usec 99\nuser_usec 50\n")
	write("memory.events", strings.Join([]string{"low 0", "high 3", "max 2", "oom 1", "oom_kill 1"}, "\n"))
	assert.Equal(t, Stats{Memory: 1234, CPUUsec: 99, OOMKills: 1}, ReadStats(d))
	assert.Equal(t, Stats{}, ReadStats(filepath.Join(d, "missing")))
}

// TestTuneCgroupAt: the controller outweighs sessions and sessions carry
// the cap (JOB-91).
func TestTuneCgroupAt(t *testing.T) {
	own := t.TempDir()
	for _, d := range []string{"controller", "sessions"} {
		require.NoError(t, os.Mkdir(filepath.Join(own, d), 0o755))
	}
	read := func(rel string) string {
		b, err := os.ReadFile(filepath.Join(own, rel))
		require.NoError(t, err)
		return string(b)
	}
	require.NoError(t, TuneAt(own, config.HostCgroup{SessionsCPU: "5"}))
	assert.Equal(t, "1000", read("controller/cpu.weight"))
	assert.Equal(t, "100", read("sessions/cpu.weight"))
	assert.Equal(t, "500000 100000", read("sessions/cpu.max"))

	require.NoError(t, TuneAt(own, config.HostCgroup{SessionsCPUWeight: 50, ControllerCPUWeight: 9000}))
	assert.Equal(t, "9000", read("controller/cpu.weight"))
	assert.Equal(t, "50", read("sessions/cpu.weight"))
	assert.Equal(t, "max", read("sessions/cpu.max"), "an unset cap lifts the old one on reload")

	assert.Error(t, TuneAt(t.TempDir(), config.HostCgroup{}), "missing children are reported")
}

// TestMoveHosts: hosts left in the controller's cgroup move to hosts,
// the controller stays (JOB-91).
func TestMoveHosts(t *testing.T) {
	own := t.TempDir()
	hosts := filepath.Join(own, "sessions", "hosts")
	require.NoError(t, os.MkdirAll(hosts, 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(own, "controller"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(own, "controller", "cgroup.procs"), []byte("7\n11\n12\n"), 0o600))
	MoveHosts(Delegation{Sessions: filepath.Join(own, "sessions"), Hosts: hosts}, 7)
	// A plain file keeps the last write; cgroupfs moves one pid per write.
	b, err := os.ReadFile(filepath.Join(hosts, "cgroup.procs"))
	require.NoError(t, err)
	assert.Equal(t, "12", string(b))
}
