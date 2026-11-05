package control_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// The pr-work step gates as packs/github declares them (JOB-10).
const nonPRWorkYAML = `projects:
  js: {}
agents:
  fake:
    command: %q
    args: [fake-agent]
pools:
  dev: {project: js, agent: fake, max_sessions: 1, idle_timeout: 1m, dir: %q}
workflows:
  pr-work:
    vars: [task]
    steps:
      - {id: implement, pool: dev, title: "Implement {{.task}}", prompt: x}
      - {id: review-merge, pool: dev, title: "Review {{.task}}", prompt: x, when: {step: implement, outcome: pr}}
      - {id: post-merge, pool: dev, title: "Post-merge {{.task}}", prompt: x, when: {step: implement, outcome: pr}}
`

// TestNonPRWork (JOB-10): implement closing with outcome done skips
// review-merge and post-merge and the run closes done; outcome pr runs
// both.
func TestNonPRWork(t *testing.T) {
	root, start := poolRoot(t, "1m")
	require.NoError(t, os.WriteFile(root.Config(), fmt.Appendf(nil, nonPRWorkYAML, os.Args[0], root.Dir), 0o600))
	_, err := config.LoadFile(root.Config())
	require.NoError(t, err)
	outcome := filepath.Join(root.Dir, "outcome")
	require.NoError(t, os.WriteFile(outcome, []byte("done\n"), 0o600))
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())

	finished := func(id string) func() bool {
		return func() bool { return showRun(t, admin, id).Run.Status == store.StatusClosed }
	}
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "pr-work", Vars: map[string]string{"task": "run a check"}, Key: "no-pr"})
	passEventually(t, root.Dir, finished(r.Run.ID), "run with outcome done did not finish")
	got := showRun(t, admin, r.Run.ID)
	assert.Equal(t, "done", got.Run.Resolution)
	s := stepByName(t, got)
	assert.Equal(t, "done", s["implement"].Resolution)
	for _, n := range []string{"review-merge", "post-merge"} {
		assert.Equal(t, store.StepSkipped, s[n].Resolution, n)
	}

	// A code change: outcome pr opens both steps; the fake agent closes
	// them in turn (it keeps closing with pr, which later steps ignore).
	require.NoError(t, os.WriteFile(outcome, []byte("pr\n"), 0o600))
	r = startRun(t, admin, rpc.RunStartParams{Workflow: "pr-work", Vars: map[string]string{"task": "fix a bug"}, Key: "pr"})
	passEventually(t, root.Dir, finished(r.Run.ID), "run with outcome pr did not finish")
	got = showRun(t, admin, r.Run.ID)
	assert.Equal(t, "done", got.Run.Resolution)
	for n, st := range stepByName(t, got) {
		assert.Equal(t, "done", st.Resolution, n)
	}
}
