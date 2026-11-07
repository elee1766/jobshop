package control

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/elee1766/acphost/pkg/cgroup"
	"github.com/elee1766/acphost/pkg/userns"
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
	cg := setupCgroupAt(info, 7)
	require.NoError(t, cg.err)
	assert.Equal(t, filepath.Join(own, "sessions"), cg.sessions)
	assert.Equal(t, filepath.Join(own, "sessions", "hosts"), cg.hosts)
	b, err := os.ReadFile(filepath.Join(own, "controller", "cgroup.procs"))
	require.NoError(t, err)
	assert.Equal(t, "7", string(b), "the controller moved into its leaf")
	for _, d := range []string{own, cg.sessions} {
		b, err := os.ReadFile(filepath.Join(d, "cgroup.subtree_control"))
		require.NoError(t, err)
		assert.Equal(t, "+cpu +memory +pids", string(b))
	}

	// Restarted inside the leaf: the delegated root is its parent.
	again := setupCgroupAt(cgroup.DelegatedInfo{Path: filepath.Join(own, "controller"), Writable: true, Controllers: info.Controllers}, 8)
	require.NoError(t, again.err)
	assert.Equal(t, cg.sessions, again.sessions)

	// Started by systemd in the leaf (DelegateSubgroup=controller, JOB-21)
	// before anything is enabled: the leaf lists no controllers, the
	// parent's are what count.
	fresh := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(fresh, "cgroup.controllers"), []byte("cpu io memory pids\n"), 0o600))
	require.NoError(t, os.Mkdir(filepath.Join(fresh, "controller"), 0o755))
	sub := setupCgroupAt(cgroup.DelegatedInfo{Path: filepath.Join(fresh, "controller"), Writable: true}, 9)
	require.NoError(t, sub.err)
	assert.Equal(t, filepath.Join(fresh, "sessions"), sub.sessions)

	assert.ErrorContains(t, setupCgroupAt(cgroup.DelegatedInfo{Path: own}, 7).err, "not delegated")
	assert.ErrorContains(t, setupCgroupAt(cgroup.DelegatedInfo{Path: own, Writable: true, Controllers: []string{"cpu"}}, 7).err, "memory")
}

func TestCgroupFor(t *testing.T) {
	hc, err := cgroupFor(config.CgroupLimits{CPU: "4", Memory: "8G", MemoryHigh: "6G", Pids: 4096}, "/sys/fs/cgroup/x/sessions")
	require.NoError(t, err)
	assert.Equal(t, cgroup.Cgroup{
		Parent: "/sys/fs/cgroup/x/sessions", CPUMax: "400000 100000",
		MemoryMax: "8589934592", MemoryHigh: "6442450944", SwapMax: "0", PidsMax: 4096,
	}, *hc)
	hc, err = cgroupFor(config.CgroupLimits{Memory: "1G", Swap: "max"}, "/p")
	require.NoError(t, err)
	assert.Equal(t, "max", hc.SwapMax, "an explicit swap is kept")
	hc, err = cgroupFor(config.CgroupLimits{Pids: 10}, "/p")
	require.NoError(t, err)
	assert.Empty(t, hc.SwapMax, "no memory limit, no swap limit")
}

// TestApplyIsolationNoDelegation: a pool with limits and no delegated
// cgroup fails with a permanent error (pool quarantine); a pool without
// isolation gets nothing (#388).
func TestApplyIsolationNoDelegation(t *testing.T) {
	c, _, _ := stallControllerYAML(t, `projects:
  js: {}
agents:
  fake: {command: /bin/false}
pools:
  plain: {agent: fake, dir: %[1]q}
  limited: {agent: fake, dir: %[1]q, isolation: {limits: {memory: 1G}}}
  mapped: {agent: fake, dir: %[1]q, isolation: {user: {uid: 1000, map_uid: 100000, map_gid: 100000, groups: ["0"]}}}
`)
	c.cgroupOnce.Do(func() {})
	c.cgroup = cgroupSetup{err: os.ErrPermission}

	var s host.Spec
	require.NoError(t, c.applyIsolation(&s, "plain"))
	assert.Nil(t, s.User)
	assert.Nil(t, s.Cgroup)
	assert.Empty(t, c.isolationKey("plain"), "no isolation: env hash unchanged")

	err := c.applyIsolation(&s, "limited")
	require.Error(t, err)
	assert.True(t, host.Permanent(err))
	assert.Contains(t, err.Error(), "no delegated cgroup")

	s = host.Spec{}
	require.NoError(t, c.applyIsolation(&s, "mapped"))
	require.NotNil(t, s.User)
	assert.Equal(t, userns.User{UID: 1000, GID: 1000, MapUID: 100000, MapGID: 100000, Groups: []int{0}}, *s.User)
	assert.NotEmpty(t, c.isolationKey("mapped"))
}

// TestPoolCgroupSystemd (#944): with the systemd driver a limited pool
// gets a scope in the sessions slice and nothing touches the controller's
// cgroup; limits.cgroup alone also gives unlimited pools a scope, so
// their agents count against the slice.
func TestPoolCgroupSystemd(t *testing.T) {
	c, _, _ := stallControllerYAML(t, `projects:
  js: {}
agents:
  fake: {command: /bin/false}
pools:
  plain: {agent: fake, dir: %[1]q}
  limited: {agent: fake, dir: %[1]q, isolation: {limits: {memory: 1G, cpu: "2"}}}
limits:
  cgroup_driver: systemd
`)
	c.cgroupOnce.Do(func() {})
	c.cgroup = cgroupSetup{err: os.ErrPermission}
	cfg := c.config()
	assert.False(t, wantsCgroup(cfg), "no delegated cgroup set up")
	assert.Empty(t, c.hostsCgroup())

	var s host.Spec
	require.NoError(t, c.applyIsolation(&s, "limited"), "no delegation needed")
	require.NotNil(t, s.Cgroup)
	assert.Equal(t, cgroup.Cgroup{
		Driver: cgroup.DriverSystemd, Parent: config.DefaultSessionsSlice,
		CPUMax: "200000 100000", MemoryMax: "1073741824", SwapMax: "0",
	}, *s.Cgroup)
	assert.Contains(t, c.isolationKey("limited"), "JOBSHOP_ISOLATION_SCOPE="+config.DefaultSessionsSlice)

	s = host.Spec{}
	require.NoError(t, c.applyIsolation(&s, "plain"))
	assert.Nil(t, s.Cgroup, "no limits and no limits.cgroup: no scope")
	assert.Empty(t, c.isolationKey("plain"))

	cfg.Limits.Cgroup = &config.HostCgroup{}
	hc, err := c.poolCgroup(cfg, "plain")
	require.NoError(t, err)
	assert.Equal(t, &cgroup.Cgroup{Driver: cgroup.DriverSystemd, Parent: config.DefaultSessionsSlice}, hc)

	spec, ok := c.preflightSpec(cfg, "limited")
	require.True(t, ok)
	assert.Equal(t, cgroup.DriverSystemd, spec.Cgroup.Driver, "preflight runs the systemd checks")
}

// TestHostPlacement (#992): with systemd a session host starts in a scope
// in the sessions slice (InScope), with cgroupfs in the hosts cgroup
// (InCgroup) or the controller's.
func TestHostPlacement(t *testing.T) {
	for _, tc := range []struct {
		name   string
		limits string
		setup  cgroupSetup
		want   hostPlacement
	}{
		{"systemd default slice", "limits:\n  cgroup_driver: systemd\n", cgroupSetup{}, hostPlacement{systemd: true, slice: config.DefaultSessionsSlice}},
		{"systemd slice", "limits:\n  cgroup_driver: systemd\n  sessions_slice: x.slice\n  cgroup: {}\n", cgroupSetup{hosts: "/h"}, hostPlacement{systemd: true, slice: "x.slice"}},
		{"cgroupfs split", "limits:\n  cgroup: {}\n", cgroupSetup{hosts: "/h"}, hostPlacement{cgroup: "/h"}},
		{"cgroupfs no split", "", cgroupSetup{hosts: "/h"}, hostPlacement{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, _, _ := stallControllerYAML(t, "projects:\n  js: {}\nagents:\n  fake: {command: /bin/false}\npools:\n  plain: {agent: fake, dir: %[1]q}\n"+tc.limits)
			c.cgroupOnce.Do(func() { c.cgroup = tc.setup })
			assert.Equal(t, tc.want, c.hostPlacement())
		})
	}
}

func TestShareDir(t *testing.T) {
	d := t.TempDir()
	sub := filepath.Join(d, "a")
	require.NoError(t, os.MkdirAll(sub, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(sub, "f"), []byte("x"), 0o600))
	gid := os.Getegid()
	require.NoError(t, shareDir(d, gid, true))
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
	assert.Equal(t, cgroupStats{memory: 1234, cpuUsec: 99, oomKills: 1}, readCgroupStats(d))
	assert.Equal(t, cgroupStats{}, readCgroupStats(filepath.Join(d, "missing")))
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
	require.NoError(t, tuneCgroupAt(own, config.HostCgroup{SessionsCPU: "5"}))
	assert.Equal(t, "1000", read("controller/cpu.weight"))
	assert.Equal(t, "100", read("sessions/cpu.weight"))
	assert.Equal(t, "500000 100000", read("sessions/cpu.max"))

	require.NoError(t, tuneCgroupAt(own, config.HostCgroup{SessionsCPUWeight: 50, ControllerCPUWeight: 9000}))
	assert.Equal(t, "9000", read("controller/cpu.weight"))
	assert.Equal(t, "50", read("sessions/cpu.weight"))
	assert.Equal(t, "max", read("sessions/cpu.max"), "an unset cap lifts the old one on reload")

	assert.Error(t, tuneCgroupAt(t.TempDir(), config.HostCgroup{}), "missing children are reported")
}

// TestMoveHosts: hosts left in the controller's cgroup move to hosts,
// the controller stays (JOB-91).
func TestMoveHosts(t *testing.T) {
	own := t.TempDir()
	hosts := filepath.Join(own, "sessions", "hosts")
	require.NoError(t, os.MkdirAll(hosts, 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(own, "controller"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(own, "controller", "cgroup.procs"), []byte("7\n11\n12\n"), 0o600))
	moveHosts(cgroupSetup{sessions: filepath.Join(own, "sessions"), hosts: hosts}, 7)
	// A plain file keeps the last write; cgroupfs moves one pid per write.
	b, err := os.ReadFile(filepath.Join(hosts, "cgroup.procs"))
	require.NoError(t, err)
	assert.Equal(t, "12", string(b))
}
