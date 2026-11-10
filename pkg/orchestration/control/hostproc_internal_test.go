package control

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	host "github.com/elee1766/acphost/pkg/acphost"

	"github.com/gfx-labs/jobshop/pkg/execution/sandbox"
	"github.com/gfx-labs/jobshop/pkg/execution/session"
)

// detachedHostDir writes a spec for a fake-agent host in a fresh session
// dir. The test binary is the agent (fake-agent in TestMain) and, through
// acphost.MaybeChild in TestMain, the detached host.
func detachedHostDir(t *testing.T) string {
	t.Helper()
	// Short: the host's acp.sock must fit the unix socket path limit.
	base, err := os.MkdirTemp("", "jsh")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(base) })
	dir := filepath.Join(base, "s")
	require.NoError(t, os.Mkdir(dir, 0o700))
	require.NoError(t, host.WriteSpec(dir, host.Spec{
		Protocol: host.Protocol, Session: "js-s1", Command: os.Args[0], Args: []string{"fake-agent"},
		Dir: base, LeaseMS: time.Minute.Milliseconds(),
	}))
	return dir
}

// TestDetachedHostSpawnStop (#904): a host started through
// acphost.StartDetached records itself in supervisor.pid, its agent
// answers, and stopHost stops it through proc's Stop and Wait, leaving
// the agent's exit recorded.
func TestDetachedHostSpawnStop(t *testing.T) {
	t.Parallel()
	c, _ := rulesController(t, rulesBase)
	dir := detachedHostDir(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, host.ExitFile), []byte(`{"reason":"old"}`), 0o600))

	d, err := c.startDetachedHost("js-s1", dir)
	require.NoError(t, err)
	t.Cleanup(func() { session.StopDetached(dir, time.Second) })
	assert.Equal(t, d.Pid, session.HostPid(dir), "host pid from supervisor.pid")
	assert.True(t, session.HostAlive(dir))
	_, ok := host.ReadExit(dir)
	assert.False(t, ok, "the previous exit.json is gone")

	nc, _, wel, err := host.Connect(t.Context(), dir)
	require.NoError(t, err, "the agent connects")
	_ = nc.Close()
	assert.Equal(t, d.Pid, wel.Pid)
	assert.Positive(t, wel.AgentPid)

	c.stopHost(dir)
	assert.False(t, session.HostAlive(dir), "stopped")
	assert.False(t, d.Running())
	ex, ok := host.ReadExit(dir)
	require.True(t, ok, "the host recorded the agent's exit")
	assert.NotEmpty(t, ex.Reason)
	assert.Equal(t, d.Pid, session.HostPid(dir), "the pid stays readable after the host ends")
}

// TestDetachedHostCgroupFallback: a sessions cgroup that cannot be opened
// does not fail the spawn; the host starts in the caller's cgroup.
func TestDetachedHostCgroupFallback(t *testing.T) {
	t.Parallel()
	c, logs := rulesController(t, rulesBase+"limits:\n  cgroup: {}\n")
	c.cgroupOnce.Do(func() { c.cgroup = sandbox.Delegation{Hosts: filepath.Join(t.TempDir(), "missing")} })
	require.NotEmpty(t, c.hostsCgroup())
	dir := detachedHostDir(t)
	d, err := c.startDetachedHost("js-s1", dir)
	require.NoError(t, err)
	t.Cleanup(func() { session.StopDetached(dir, time.Second) })
	assert.True(t, d.Running())
	assert.Contains(t, logs.b.String(), "session host stays in the controller cgroup")
}

// TestDetachedHostScopeFallback (#992): with the systemd driver and no
// manager to start the scope, the host starts in the caller's cgroup.
func TestDetachedHostScopeFallback(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root uses the system manager")
	}
	// No manager socket there: the scope start fails before the host.
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	c, logs := rulesController(t, rulesBase+"limits:\n  cgroup_driver: systemd\n")
	dir := detachedHostDir(t)
	d, err := c.startDetachedHost("js-s1", dir)
	require.NoError(t, err)
	t.Cleanup(func() { session.StopDetached(dir, time.Second) })
	assert.True(t, d.Running())
	assert.Contains(t, logs.b.String(), "session host stays in the controller cgroup")
}

// TestStopHostKills (#904): a host that ignores SIGTERM is killed once
// the wait runs out.
func TestStopHostKills(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	holder := exec.Command(os.Args[0], "hold-lock", dir)
	holder.Env = append(os.Environ(), "HOLD_LOCK_IGNORE_TERM=1")
	out, err := holder.StdoutPipe()
	require.NoError(t, err)
	require.NoError(t, holder.Start())
	exited := make(chan error, 1)
	line, _ := bufio.NewReader(out).ReadString('\n')
	go func() { exited <- holder.Wait() }()
	t.Cleanup(func() { _ = holder.Process.Kill() })
	require.Equal(t, "ready\n", line)
	require.True(t, session.HostAlive(dir))

	require.True(t, session.StopDetached(dir, 100*time.Millisecond), "a host was running")
	select {
	case err := <-exited:
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "killed"), "SIGKILL ended it: %v", err)
	case <-time.After(time.Minute):
		t.Fatal("host not killed")
	}
	assert.False(t, session.HostAlive(dir))
	assert.False(t, session.StopDetached(dir, time.Second), "nothing left to stop")
}

// TestNoSupervisorPid (#918): a session dir with no supervisor.pid has no
// host, even with the old host.pid file (acphost v0.20 dropped it).
func TestNoSupervisorPid(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	assert.False(t, session.HostAlive(dir))
	assert.Zero(t, session.HostPid(dir))
	assert.False(t, session.StopDetached(dir, time.Second))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "host.pid"), []byte("4242\n"), 0o600))
	assert.Zero(t, session.HostPid(dir))
	assert.False(t, session.HostAlive(dir))
	assert.False(t, session.StopDetached(dir, time.Second))
}
