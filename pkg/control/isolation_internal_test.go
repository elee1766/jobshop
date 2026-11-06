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
