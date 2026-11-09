package session

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBackgroundJobs: a child with another executable than the agent is
// a job; an unknown agent pid has none (JOB-70).
func TestBackgroundJobs(t *testing.T) {
	t.Parallel()
	assert.Nil(t, BackgroundJobs(0, ""))

	cmd := exec.Command("sleep", "60")
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})
	// This test binary plays the agent and sleep is the job it left.
	want := strconv.Itoa(cmd.Process.Pid) + " "
	var found bool
	for _, j := range BackgroundJobs(os.Getpid(), "") {
		if strings.HasPrefix(j, want) {
			found = true
		}
	}
	assert.True(t, found, "sleep child is listed as a job")
}
