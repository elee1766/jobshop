package control_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// runRoot starts a controller whose pools never spawn sessions, with a
// pack declaring a three-step workflow.
func runRoot(t *testing.T) *rpc.Client {
	t.Helper()
	return runRootLease(t, "")
}

// runRootLease is runRoot with a lease (Go duration, "" for the default).
func runRootLease(t *testing.T, lease string) *rpc.Client {
	t.Helper()
	t.Parallel()
	dir, err := os.MkdirTemp("", "js") // short path for the unix socket
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	root := config.Root{Dir: dir}
	pack := filepath.Join(dir, "pk")
	require.NoError(t, os.MkdirAll(filepath.Join(pack, "prompts"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(pack, "prompts", "impl.md"), []byte("Implement #${issue} in ${repo}.\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(pack, config.PackFileName), []byte(`workflows:
  pr:
    vars: [issue]
    issue: issue
    escalate_pool: dev
    steps:
      - {id: implement, pool: dev, retries: 3, prompt_file: prompts/impl.md, title: "Implement #${issue}"}
      - {id: review, pool: refinery, prompt: "Review #${issue}", title: "Review #${issue}"}
      - {id: deploy, pool: refinery, prompt: "Deploy", title: "Deploy #${issue}"}
`), 0o600))
	cfg := fmt.Sprintf(`packs:
  - path: pk
    vars: {repo: /src}
lease: %q
projects:
  js: {}
agents:
  fake:
    command: %q
pools:
  dev: {project: js, agent: fake, max_sessions: 0, dir: %q}
  refinery: {project: js, agent: fake, max_sessions: 0, dir: %q}
`, lease, os.Args[0], dir, dir)
	require.NoError(t, os.WriteFile(root.Config(), []byte(cfg), 0o600))
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- control.Run(ctx, root, control.Options{Tick: 50 * time.Millisecond}) }()
	t.Cleanup(func() {
		cancel()
		if err := <-done; !errors.Is(err, context.Canceled) {
			require.NoError(t, err)
		}
	})
	return dial(t, root.AdminSocket())
}

func startRun(t *testing.T, admin *rpc.Client, p rpc.RunStartParams) rpc.RunResult {
	t.Helper()
	var r rpc.RunResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodRunStart, p, &r))
	return r
}

// readyIDs lists ready step tickets (escalation tasks are left out).
func readyIDs(t *testing.T, admin *rpc.Client) []string {
	t.Helper()
	var ts []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodReady, rpc.ReadyParams{Project: "js"}, &ts))
	out := []string{}
	for _, x := range ts {
		if x.Kind == store.KindStep {
			out = append(out, x.ID)
		}
	}
	return out
}

func showRun(t *testing.T, admin *rpc.Client, id string) rpc.RunResult {
	t.Helper()
	var r rpc.RunResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodRunShow, rpc.IDParams{ID: id}, &r))
	return r
}

func closeStep(t *testing.T, admin *rpc.Client, id, res string) {
	t.Helper()
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketClose, rpc.CloseParams{ID: id, Resolution: res}, nil))
}

func runState(t *testing.T, r rpc.RunResult) string {
	t.Helper()
	f, err := store.DecodeFields[store.RunFields](r.Run)
	require.NoError(t, err)
	return f.State
}

// TestRunStart creates the run and its steps in one graph: only step 1 is
// ready, templates are filled, meta carries the issue.
func TestRunStart(t *testing.T) {
	admin := runRoot(t)
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "42"}})

	assert.Equal(t, store.KindRun, r.Run.Kind)
	assert.Empty(t, r.Run.Pool, "runs are never claimed")
	assert.Equal(t, store.RunRunning, runState(t, r))
	require.Len(t, r.Steps, 3)
	s0 := r.Steps[0]
	assert.Equal(t, store.KindStep, s0.Kind)
	assert.Equal(t, "dev", s0.Pool)
	assert.Equal(t, r.Run.ID, s0.ParentID)
	assert.Equal(t, "Implement #42", s0.Title)
	assert.Equal(t, "Implement #42 in /src.\n", s0.Body, "run vars and pack vars filled")
	assert.JSONEq(t, `{"github":{"issue":42}}`, string(s0.Meta))
	assert.JSONEq(t, `{"github":{"issue":42}}`, string(r.Run.Meta))
	sf, err := store.DecodeFields[store.StepFields](s0)
	require.NoError(t, err)
	assert.Equal(t, store.StepFields{Run: r.Run.ID, Step: "implement", Index: 0, Retries: 3}, sf)
	sf, err = store.DecodeFields[store.StepFields](r.Steps[1])
	require.NoError(t, err)
	assert.Equal(t, config.DefaultStepRetries, sf.Retries)

	assert.Equal(t, []string{s0.ID}, readyIDs(t, admin), "only the first step is ready")
	closeStep(t, admin, s0.ID, "done")
	assert.Equal(t, []string{r.Steps[1].ID}, readyIDs(t, admin))

	var runs []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodRunList, rpc.RunListParams{State: store.RunRunning}, &runs))
	require.Len(t, runs, 1)
	assert.Equal(t, r.Run.ID, runs[0].ID)
}

func TestRunStartErrors(t *testing.T) {
	admin := runRoot(t)
	for _, tt := range []struct {
		name string
		p    rpc.RunStartParams
		want string
	}{
		{"unknown workflow", rpc.RunStartParams{Workflow: "nope"}, `unknown workflow "nope"`},
		{"missing var", rpc.RunStartParams{Workflow: "pr"}, "needs var issue"},
		{"unknown var", rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "1", "x": "2"}}, `no var "x"`},
		{"other project", rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "1"}, Project: "zz"}, "runs in project js"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := admin.Call(t.Context(), rpc.MethodRunStart, tt.p, nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
	err := admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "x", Kind: store.KindStep}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "run start")
}

// TestRunKeyDedup returns the open run for a repeated key, and starts a
// new one once that run is closed.
func TestRunKeyDedup(t *testing.T) {
	admin := runRoot(t)
	p := rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "7"}, Key: "sha1"}
	a := startRun(t, admin, p)
	b := startRun(t, admin, p)
	assert.Equal(t, a.Run.ID, b.Run.ID)
	assert.True(t, b.Existing)
	assert.False(t, a.Existing)
	other := startRun(t, admin, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "7"}, Key: "sha2"})
	assert.NotEqual(t, a.Run.ID, other.Run.ID)

	require.NoError(t, admin.Call(t.Context(), rpc.MethodRunCancel, rpc.IDParams{ID: a.Run.ID}, nil))
	c := startRun(t, admin, p)
	assert.NotEqual(t, a.Run.ID, c.Run.ID)
	assert.False(t, c.Existing)
}

// TestRunCompletes closes the run done when its last step closes done.
func TestRunCompletes(t *testing.T) {
	admin := runRoot(t)
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "1"}})
	for i, s := range r.Steps {
		closeStep(t, admin, s.ID, "done")
		if i < len(r.Steps)-1 {
			got := showRun(t, admin, r.Run.ID)
			assert.Equal(t, store.StatusOpen, got.Run.Status)
		}
	}
	got := showRun(t, admin, r.Run.ID)
	assert.Equal(t, store.StatusClosed, got.Run.Status)
	assert.Equal(t, "done", got.Run.Resolution)
	assert.Equal(t, store.RunDone, runState(t, got))
}

// TestRunCancel closes open steps and the run as cancelled.
func TestRunCancel(t *testing.T) {
	admin := runRoot(t)
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "1"}})
	closeStep(t, admin, r.Steps[0].ID, "done")
	var got rpc.RunResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodRunCancel, rpc.IDParams{ID: r.Run.ID}, &got))
	assert.Equal(t, store.StatusClosed, got.Run.Status)
	assert.Equal(t, store.RunCancelled, got.Run.Resolution)
	assert.Equal(t, store.RunCancelled, runState(t, got))
	assert.Equal(t, "done", got.Steps[0].Resolution, "finished steps keep their resolution")
	for _, s := range got.Steps[1:] {
		assert.Equal(t, store.StatusClosed, s.Status)
		assert.Equal(t, store.RunCancelled, s.Resolution)
	}
	assert.Empty(t, readyIDs(t, admin))
	err := admin.Call(t.Context(), rpc.MethodRunCancel, rpc.IDParams{ID: r.Run.ID}, nil)
	assert.Error(t, err, "cancel twice")
}

// TestRunResume escalates at the limit, then resume reopens the failed
// step with failures reset and the run back to running.
func TestRunResume(t *testing.T) {
	admin := runRoot(t)
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "1"}})
	closeStep(t, admin, r.Steps[0].ID, "done")
	// review has the default 2 retries.
	closeStep(t, admin, r.Steps[1].ID, "failed")
	closeStep(t, admin, r.Steps[1].ID, "failed")
	got := showRun(t, admin, r.Run.ID)
	require.Equal(t, store.RunNeedsHuman, runState(t, got))
	assert.Empty(t, readyIDs(t, admin), "later steps stay blocked")

	require.NoError(t, admin.Call(t.Context(), rpc.MethodRunResume, rpc.RunResumeParams{ID: r.Run.ID}, &got))
	assert.Equal(t, store.RunRunning, runState(t, got))
	assert.Equal(t, store.StatusOpen, got.Steps[1].Status)
	f, err := store.DecodeFields[store.StepFields](got.Steps[1])
	require.NoError(t, err)
	assert.Equal(t, 0, f.Failures)
	assert.Equal(t, []string{r.Steps[1].ID}, readyIDs(t, admin))

	// A resumed step gets its full retries again: one failure only reopens.
	closeStep(t, admin, r.Steps[1].ID, "failed")
	assert.Equal(t, store.RunRunning, runState(t, showRun(t, admin, r.Run.ID)))

	// --step reopens a finished step; the step after it is blocked again.
	require.NoError(t, admin.Call(t.Context(), rpc.MethodRunResume, rpc.RunResumeParams{ID: r.Run.ID, Step: "implement"}, &got))
	assert.Equal(t, store.StatusOpen, got.Steps[0].Status)
	assert.Equal(t, []string{r.Steps[0].ID}, readyIDs(t, admin))

	err = admin.Call(t.Context(), rpc.MethodRunResume, rpc.RunResumeParams{ID: r.Run.ID, Step: "nope"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `no step "nope"`)
}

// TestSelfdevPRWork loads packs/selfdev and starts pr-work as the planner
// does: three steps in dev, refinery, refinery, only implement ready, and
// a repeated start with the same key returns the same run.
func TestSelfdevPRWork(t *testing.T) {
	t.Parallel()
	dir, err := os.MkdirTemp("", "js") // short path for the unix socket
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	root := config.Root{Dir: dir}
	pack, err := filepath.Abs("../../packs/selfdev")
	require.NoError(t, err)
	// dev_sessions 0 and no master-red check: the test starts no agents.
	cfg := fmt.Sprintf(`projects:
  js: {}
agents:
  fake:
    command: %q
packs:
  - path: %s
    vars: {repo: %q, agent: fake, dev_sessions: 0}
pools:
  plan: {max_sessions: 0}
  refinery: {max_sessions: 0}
triggers:
  master-red:
    when: {check: "false"}
`, os.Args[0], pack, dir)
	require.NoError(t, os.WriteFile(root.Config(), []byte(cfg), 0o600))
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- control.Run(ctx, root, control.Options{Tick: time.Hour}) }()
	t.Cleanup(func() {
		cancel()
		if err := <-done; !errors.Is(err, context.Canceled) {
			require.NoError(t, err)
		}
	})
	admin := dial(t, root.AdminSocket())

	p := rpc.RunStartParams{Workflow: "pr-work", Vars: map[string]string{"issue": "93"}, Key: "issue-93"}
	r := startRun(t, admin, p)
	require.Len(t, r.Steps, 3)
	var pools, titles []string
	for _, s := range r.Steps {
		pools = append(pools, s.Pool)
		titles = append(titles, s.Title)
		assert.JSONEq(t, `{"github":{"issue":93}}`, string(s.Meta))
	}
	assert.Equal(t, []string{"dev", "refinery", "refinery"}, pools)
	assert.Equal(t, "Implement #93", titles[0])
	assert.Contains(t, r.Steps[0].Body, "gh issue view 93 --comments")
	assert.Contains(t, r.Steps[2].Body, "jobshop version")
	assert.Equal(t, []string{r.Steps[0].ID}, readyIDs(t, admin), "only implement is ready")

	again := startRun(t, admin, p)
	assert.True(t, again.Existing)
	assert.Equal(t, r.Run.ID, again.Run.ID)
}
