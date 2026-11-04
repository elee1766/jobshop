package control

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

// TestProcGone: a live process is not gone; an exited, unreaped child (a
// zombie) is, so stopSessionProcs does not wait out its deadline (#433).
func TestProcGone(t *testing.T) {
	assert.False(t, procGone(os.Getpid()))

	cmd := exec.Command("true")
	require.NoError(t, cmd.Start())
	pid := cmd.Process.Pid
	// Wait for the exit without reaping: WNOWAIT leaves the zombie.
	var info unix.Siginfo
	require.NoError(t, unix.Waitid(unix.P_PID, pid, &info, unix.WEXITED|unix.WNOWAIT, nil))
	assert.True(t, procGone(pid), "zombie counts as gone")
	require.NoError(t, cmd.Wait())
	assert.True(t, procGone(pid), "reaped")
}
