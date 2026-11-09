package control_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// TestAgentPoolPack loads packs/agent (#634), the port of the gascity
// agent pool, and runs its pr-work workflow end to end with the fake
// agent: deliver, review and merge each close done.
func TestAgentPoolPack(t *testing.T) {
	root, start := poolRoot(t, "1m")
	repo := filepath.Join(root.Dir, "app")
	for _, args := range [][]string{
		{"init", "-q", "-b", "main", repo},
		{"-C", repo, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "init"},
		{"-C", repo, "update-ref", "refs/remotes/origin/main", "HEAD"},
	} {
		out, err := exec.Command("git", args...).CombinedOutput()
		require.NoError(t, err, string(out))
	}
	pack, err := filepath.Abs("../../packs/agent")
	require.NoError(t, err)
	write := func(extra string) {
		cfg := fmt.Sprintf("projects:\n  js: {}\nagents:\n  fake:\n    command: %q\n    args: [fake-agent]\npacks:\n  - path: %q\n    prefix: app-\n    vars: {agent: fake, checkout: %q, repo: acme/app%s}\n",
			os.Args[0], pack, repo, extra)
		require.NoError(t, os.WriteFile(root.Config(), []byte(cfg), 0o600))
	}

	// Unisolated defaults for the run, so it works on any CI host. Only
	// the workflow name is assumed (JOB-2: no pack content asserts).
	write("")
	_, err = config.LoadFile(root.Config())
	require.NoError(t, err)

	stop := start()
	defer stop()
	ctx := t.Context()
	admin := dial(t, root.AdminSocket())
	runs, _ := subscribe(ctx, t, root.AdminSocket(), rpc.EventsSubscribeParams{Kinds: []string{rpc.EventKindRun}})
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "app-pr-work", Vars: map[string]string{"task": "add a widget"}, Key: "task-1"})
	require.NotEmpty(t, r.Steps)
	done := func() bool {
		for _, ev := range runs.get() {
			if ev.ID == r.Run.ID && ev.Status == domain.StatusClosed {
				return true
			}
		}
		return false
	}
	passEventually(t, root.Dir, done, "the agent pack's pr-work run did not finish")
	got := showRun(t, admin, r.Run.ID)
	require.Equal(t, "done", got.Run.Resolution)
	for _, s := range got.Steps {
		require.Equal(t, domain.StatusClosed, s.Status, s.Title)
	}
}
