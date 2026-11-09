package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCheckPushed (JOB-70): release and handoff refuse while the current
// branch has a commit no remote has, unless forced; a pushed branch, a
// detached HEAD and a non-git dir pass.
func TestCheckPushed(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("no git")
	}
	t.Parallel()
	base := t.TempDir()
	remote, work := filepath.Join(base, "remote.git"), filepath.Join(base, "work")
	run := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
	}
	run(base, "init", "-q", "--bare", remote)
	run(base, "init", "-q", "-b", "main", work)
	run(work, "remote", "add", "origin", remote)
	run(work, "commit", "-q", "--allow-empty", "-m", "first")
	run(work, "push", "-q", "origin", "main")
	assert.NoError(t, checkPushed(work, "release", false), "everything pushed")

	run(work, "switch", "-q", "-c", "issue-1")
	run(work, "commit", "-q", "--allow-empty", "-m", "work in progress")
	err := checkPushed(work, "release", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "work in progress")
	assert.Contains(t, err.Error(), "--force")
	assert.NoError(t, checkPushed(work, "release", true), "forced")

	run(work, "push", "-q", "origin", "issue-1")
	assert.NoError(t, checkPushed(work, "handoff", false), "pushed")

	run(work, "commit", "-q", "--allow-empty", "-m", "later")
	run(work, "checkout", "-q", "--detach")
	assert.NoError(t, checkPushed(work, "release", false), "detached HEAD is not checked")
	assert.NoError(t, checkPushed(t.TempDir(), "release", false), "not a git dir")
}
