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
	"github.com/elee1766/acphost/pkg/proc"
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
	t.Cleanup(func() { stopDetachedHost(dir, time.Second) })
	assert.Equal(t, d.Pid, hostPid(dir), "host pid from supervisor.pid")
	assert.True(t, hostAlive(dir))
	_, ok := host.ReadExit(dir)
	assert.False(t, ok, "the previous exit.json is gone")

	nc, _, wel, err := host.Connect(t.Context(), dir)
	require.NoError(t, err, "the agent connects")
	_ = nc.Close()
	assert.Equal(t, d.Pid, wel.Pid)
	assert.Positive(t, wel.AgentPid)

	c.stopHost(dir)
	assert.False(t, hostAlive(dir), "stopped")
	assert.False(t, d.Running())
	ex, ok := host.ReadExit(dir)
	require.True(t, ok, "the host recorded the agent's exit")
	assert.NotEmpty(t, ex.Reason)
	assert.Equal(t, d.Pid, hostPid(dir), "the pid stays readable after the host ends")
}

// TestDetachedHostCgroupFallback: a sessions cgroup that cannot be opened
// does not fail the spawn; the host starts in the caller's cgroup.
func TestDetachedHostCgroupFallback(t *testing.T) {
	t.Parallel()
	c, logs := rulesController(t, rulesBase+"limits:\n  cgroup: {}\n")
	c.cgroupOnce.Do(func() { c.cgroup = cgroupSetup{hosts: filepath.Join(t.TempDir(), "missing")} })
	require.NotEmpty(t, c.hostsCgroup())
	dir := detachedHostDir(t)
	d, err := c.startDetachedHost("js-s1", dir)
	require.NoError(t, err)
	t.Cleanup(func() { stopDetachedHost(dir, time.Second) })
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
	require.True(t, hostAlive(dir))

	require.True(t, stopDetachedHost(dir, 100*time.Millisecond), "a host was running")
	select {
	case err := <-exited:
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "killed"), "SIGKILL ended it: %v", err)
	case <-time.After(time.Minute):
		t.Fatal("host not killed")
	}
	assert.False(t, hostAlive(dir))
	assert.False(t, stopDetachedHost(dir, time.Second), "nothing left to stop")
}

// TestLegacyHostLiveness (#904): a host from before supervisor.pid is
// still found and its pid read through host.pid.
func TestLegacyHostLiveness(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	assert.False(t, hostAlive(dir))
	assert.Zero(t, hostPid(dir))
	require.NoError(t, os.WriteFile(filepath.Join(dir, host.PidFile), []byte("4242\n"), 0o600))
	assert.Equal(t, 4242, hostPid(dir))
	assert.False(t, hostAlive(dir), "no lock held")
	_, err := proc.Open(dir)
	require.ErrorIs(t, err, os.ErrNotExist)
}
