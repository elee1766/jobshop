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

	"github.com/elee1766/jobshop/pkg/clock"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// runRoot starts a controller whose pools never spawn sessions, with a
// pack declaring a three-step workflow.
func runRoot(t *testing.T) *rpc.Client {
	t.Helper()
	return runRootLease(t, "", nil)
}

// runRootLease is runRoot with a lease (Go duration, "" for the default)
// and clock (nil for the real one).
func runRootLease(t *testing.T, lease string, clk clock.Clock) *rpc.Client {
	t.Helper()
	t.Parallel()
	dir, err := os.MkdirTemp("", "js") // short path for the unix socket
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	root := config.Root{Dir: dir}
	pack := filepath.Join(dir, "pk")
	require.NoError(t, os.MkdirAll(filepath.Join(pack, "prompts"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(pack, "prompts", "impl.md"), []byte("Implement #{{.issue}} in {{.repo}}.\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(pack, config.PackFileName), []byte(`workflows:
  pr:
    vars: [issue]
    issue: issue
    repo: acme/widget
    escalate_pool: dev
    steps:
      - {id: implement, pool: dev, retries: 3, prompt_file: prompts/impl.md, title: "Implement #{{.issue}}"}
      - {id: review, pool: refinery, prompt: "Review #{{.issue}}", title: "Review #{{.issue}}"}
      - {id: deploy, pool: refinery, prompt: "Deploy", title: "Deploy #{{.issue}}"}
  routed:
    vars: [review_pool]
    steps:
      - {id: review, pool: "{{.review_pool}}", prompt: "Review", title: "Review"}
  locked:
    vars: [pr]
    steps:
      - {id: work, pool: dev, prompt: "Work", title: "Work {{.pr}}", needs_resources: ["pr:acme#{{.pr}}", emu]}
  placed:
    vars: [pr, risk]
    steps:
      - id: review
        prompt: Review
        title: "Review {{.pr}}"
        placement:
          required: ["pr:acme#{{.pr}}"]
          prefer_for: 10m
          options:
            - {pool: refinery, resources: [flagship], weight: 100}
            - {pool: dev, resources: [workhorse], weight: 10, when: "{{.risk}} != 'high'"}
  graph:
    vars: [issue]
    steps:
      - {id: build, pool: dev, prompt: b, title: "Build #{{.issue}}"}
      - {id: review, pool: refinery, prompt: r, title: "Review #{{.issue}}", needs: [build]}
      - {id: device, pool: refinery, prompt: d, title: "Device #{{.issue}}", needs: [build]}
      - {id: rollout, pool: dev, prompt: o, title: "Rollout #{{.issue}}", needs: [review, device], when: {step: review, outcome: approve}}
      - {id: accept, pool: dev, prompt: a, title: "Accept #{{.issue}}", when: {step: review, outcome: approve}}
`), 0o600))
	cfg := fmt.Sprintf(`packs:
  - path: pk
    vars: {repo: /src, review_pool: refinery}
lease: %q
resources:
  pr: {keyed: true}
  emu: {max: 2}
  flagship: {max: 1}
  workhorse: {max: 4}
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
	go func() {
		done <- control.Run(ctx, root, withPass(dir, control.Options{Tick: time.Hour, Clock: clk}))
	}()
	t.Cleanup(func() {
		cancel()
		if err := <-done; !errors.Is(err, context.Canceled) {
			require.NoError(t, err)
		}
	})
	admin := dial(t, root.AdminSocket())
	if ch, ok := passChans.Load(dir); ok {
		passChans.Store(admin, ch)
	}
	return admin
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
	assert.JSONEq(t, `{"github":{"repo":"acme/widget","issue":42}}`, string(s0.Meta))
	assert.JSONEq(t, `{"github":{"repo":"acme/widget","issue":42}}`, string(r.Run.Meta))
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

// TestRunStepPoolVar routes one workflow's step to two pools by a run
// var, defaults it from pack vars and rejects an unknown pool (#384).
func TestRunStepPoolVar(t *testing.T) {
	admin := runRoot(t)
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "routed", Vars: map[string]string{"review_pool": "dev"}})
	assert.Equal(t, "dev", r.Steps[0].Pool)
	assert.Equal(t, "dev", showRun(t, admin, r.Run.ID).Steps[0].Pool)
	r = startRun(t, admin, rpc.RunStartParams{Workflow: "routed"})
	assert.Equal(t, "refinery", r.Steps[0].Pool, "pack var default")

	err := admin.Call(t.Context(), rpc.MethodRunStart, rpc.RunStartParams{Workflow: "routed", Vars: map[string]string{"review_pool": "nope"}}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown pool "nope"`)
	var runs []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodRunList, rpc.RunListParams{}, &runs))
	assert.Len(t, runs, 2, "a bad pool creates no ticket")
}

// TestRunStepResources: needs_resources is filled from run vars, and a
// keyed pr resource lets one step per PR be ready at a time (#400).
func TestRunStepResources(t *testing.T) {
	admin := runRoot(t)
	a := startRun(t, admin, rpc.RunStartParams{Workflow: "locked", Vars: map[string]string{"pr": "1"}})
	b := startRun(t, admin, rpc.RunStartParams{Workflow: "locked", Vars: map[string]string{"pr": "1"}})
	c := startRun(t, admin, rpc.RunStartParams{Workflow: "locked", Vars: map[string]string{"pr": "2"}})
	sf, err := store.DecodeFields[store.StepFields](a.Steps[0])
	require.NoError(t, err)
	assert.Equal(t, []string{"pr:acme#1", "emu"}, sf.Resources)

	require.NoError(t, admin.Call(t.Context(), rpc.MethodClaim, rpc.ClaimParams{ID: a.Steps[0].ID, Holder: "s1"}, nil))
	assert.ElementsMatch(t, []string{c.Steps[0].ID}, readyIDs(t, admin), "pr 1 is held; pr 2 is free")
	require.NoError(t, admin.Call(t.Context(), rpc.MethodClaim, rpc.ClaimParams{ID: c.Steps[0].ID, Holder: "s2"}, nil))
	assert.Empty(t, readyIDs(t, admin), "emu (max 2) is full too")
	// run show says what b waits on (#408).
	assert.Equal(t, []store.ResourceWait{{Ticket: b.Steps[0].ID, Key: "emu", Held: 2, Capacity: 2}}, showRun(t, admin, b.Run.ID).Waits)
	assert.Empty(t, showRun(t, admin, a.Run.ID).Waits, "a holds its slots")

	closeStep(t, admin, a.Steps[0].ID, "done")
	assert.Equal(t, []string{b.Steps[0].ID}, readyIDs(t, admin), "closing frees pr 1 and an emu slot")
}

// TestRunPlacement: a low-risk step falls back to its lower option only
// after prefer_for, and back up when the top option frees; a high-risk
// step never falls back (#409).
func TestRunPlacement(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC))
	admin := runRootLease(t, "", clk)
	hog := startRun(t, admin, rpc.RunStartParams{Workflow: "placed", Vars: map[string]string{"pr": "1", "risk": "low"}})
	low := startRun(t, admin, rpc.RunStartParams{Workflow: "placed", Vars: map[string]string{"pr": "2", "risk": "low"}})
	high := startRun(t, admin, rpc.RunStartParams{Workflow: "placed", Vars: map[string]string{"pr": "3", "risk": "high"}})
	placement := func(id string) (string, *store.StepPlacement) {
		r := showRun(t, admin, id)
		sf, err := store.DecodeFields[store.StepFields](r.Steps[0])
		require.NoError(t, err)
		return r.Steps[0].Pool, sf.Placement
	}
	pool, p := placement(high.Run.ID)
	assert.Equal(t, "refinery", pool)
	require.Len(t, p.Options, 1, "high risk drops the workhorse option")
	pool, p = placement(low.Run.ID)
	assert.Equal(t, "refinery", pool, "starts on the top option")
	assert.Equal(t, []string{"pr:acme#2", "flagship"}, p.Options[0].Resources)

	// hog takes the only flagship slot.
	require.NoError(t, admin.Call(t.Context(), rpc.MethodClaim, rpc.ClaimParams{ID: hog.Steps[0].ID, Holder: "s1"}, nil))
	pass(t, admin, 2)
	pool, _ = placement(low.Run.ID)
	assert.Equal(t, "refinery", pool, "holds out for flagship during prefer_for")

	clk.Advance(11 * time.Minute)
	passUntil(t, admin, func() bool { pool, _ = placement(low.Run.ID); return pool == "dev" }, "falls back after prefer_for")
	_, p = placement(low.Run.ID)
	assert.True(t, p.Fallback())
	assert.Equal(t, []string{"pr:acme#2", "workhorse"}, p.Options[p.Chosen].Resources)
	assert.Contains(t, readyIDs(t, admin), low.Steps[0].ID, "workhorse is free")
	pool, _ = placement(high.Run.ID)
	assert.Equal(t, "refinery", pool, "high risk never falls back")
	assert.NotContains(t, readyIDs(t, admin), high.Steps[0].ID)

	// Flagship frees before low is claimed: it moves back up.
	closeStep(t, admin, hog.Steps[0].ID, "done")
	passUntil(t, admin, func() bool { pool, _ = placement(low.Run.ID); return pool == "refinery" }, "back on flagship once free")
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
		{"empty var", rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": " "}, Key: "issue-"}, "var issue is empty"},
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

// TestRunStartAfter creates the outside blocker with the run, so the
// second run's first step is never ready before the first run is done
// (#149).
func TestRunStartAfter(t *testing.T) {
	admin := runRoot(t)
	a := startRun(t, admin, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "1"}})
	b := startRun(t, admin, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "2"}, After: []string{a.Run.ID}})
	assert.Equal(t, []string{a.Steps[0].ID}, readyIDs(t, admin), "b waits on run a")
	for _, s := range a.Steps {
		closeStep(t, admin, s.ID, "done")
	}
	assert.Equal(t, []string{b.Steps[0].ID}, readyIDs(t, admin))

	err := admin.Call(t.Context(), rpc.MethodRunStart, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "3"}, After: []string{"js-nope"}}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "after js-nope")
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

// TestSelfdevPRWork starts a pr-work shaped like packs/selfdev's as the
// planner does: three steps in dev, refinery, deploy-check, only
// implement ready, and a repeated start with the same key returns the
// same run.
func TestSelfdevPRWork(t *testing.T) {
	admin := selfdevRoot(t)
	p := rpc.RunStartParams{Workflow: "pr-work", Vars: map[string]string{"issue": "93"}, Key: "issue-93"}
	r := startRun(t, admin, p)
	require.Len(t, r.Steps, 3)
	var pools, titles []string
	for _, s := range r.Steps {
		pools = append(pools, s.Pool)
		titles = append(titles, s.Title)
		assert.JSONEq(t, `{"github":{"repo":"elee1766/jobshop","issue":93}}`, string(s.Meta))
	}
	assert.Equal(t, []string{"dev", "refinery", "deploy-check"}, pools)
	assert.Equal(t, "Implement #93", titles[0])
	assert.Contains(t, r.Steps[0].Body, "gh issue view 93 --comments")
	assert.Equal(t, []string{r.Steps[0].ID}, readyIDs(t, admin), "only implement is ready")

	again := startRun(t, admin, p)
	assert.True(t, again.Existing)
	assert.Equal(t, r.Run.ID, again.Run.ID)
}

// TestPostMergeOwnPool (#167): while one run's post-merge step is held
// waiting on a deploy, another run's review-merge is still claimable in
// the refinery pool.
func TestPostMergeOwnPool(t *testing.T) {
	admin := selfdevRoot(t)
	a := startRun(t, admin, rpc.RunStartParams{Workflow: "pr-work", Vars: map[string]string{"issue": "146"}, Key: "issue-146"})
	b := startRun(t, admin, rpc.RunStartParams{Workflow: "pr-work", Vars: map[string]string{"issue": "147"}, Key: "issue-147"})
	closeStep(t, admin, a.Steps[0].ID, "done")
	closeStep(t, admin, a.Steps[1].ID, "done")
	var held rpc.ClaimResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodClaim, rpc.ClaimParams{Project: "js", Pool: "deploy-check", Holder: "js-deploy"}, &held))
	require.Equal(t, a.Steps[2].ID, held.Ticket.ID, "post-merge is claimed in deploy-check")

	closeStep(t, admin, b.Steps[0].ID, "done")
	var cl rpc.ClaimResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodClaim, rpc.ClaimParams{Project: "js", Pool: "refinery", Holder: "js-refinery"}, &cl))
	assert.Equal(t, b.Steps[1].ID, cl.Ticket.ID, "review-merge is not queued behind the deploy wait")
}

// TestSelfdevDrill runs the escalation drill (#114): one failed close of
// its only step puts the run in needs_human and files the escalation.
func TestSelfdevDrill(t *testing.T) {
	admin := selfdevRoot(t)
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "drill", Vars: map[string]string{"issue": "114"}, Key: "drill-test"})
	require.Len(t, r.Steps, 1)
	assert.Equal(t, "dev", r.Steps[0].Pool)
	assert.Contains(t, r.Steps[0].Body, "jobshop close -r failed")
	closeStep(t, admin, r.Steps[0].ID, "failed")

	assert.Equal(t, store.RunNeedsHuman, runState(t, showRun(t, admin, r.Run.ID)))
	esc := escalations(t, admin, r.Run.ID)
	require.Len(t, esc, 1)
	assert.Equal(t, "plan", esc[0].Pool)
	assert.Contains(t, esc[0].Body, "Retry limit 1 reached")
}

// selfdevRoot starts a controller on a fixture pack shaped like
// packs/selfdev's pr-work and drill (JOB-2: not packs/, so pack edits do
// not break it) that starts no agents.
func selfdevRoot(t *testing.T) *rpc.Client {
	t.Helper()
	t.Parallel()
	dir, err := os.MkdirTemp("", "js") // short path for the unix socket
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	root := config.Root{Dir: dir}
	pack := filepath.Join(dir, "pack")
	require.NoError(t, os.MkdirAll(pack, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(pack, "pack.yaml"), []byte(`pools:
  dev: {agent: "{{.agent}}", repo: "{{.checkout}}", max_sessions: 0}
  plan: {agent: "{{.agent}}", repo: "{{.checkout}}", max_sessions: 0}
  refinery: {agent: "{{.agent}}", repo: "{{.checkout}}", max_sessions: 0}
  deploy-check: {agent: "{{.agent}}", repo: "{{.checkout}}", max_sessions: 0}
workflows:
  pr-work:
    vars: [issue]
    issue: issue
    repo: elee1766/jobshop
    escalate_pool: plan
    steps:
      - {id: implement, pool: dev, retries: 3, title: "Implement #{{.issue}}", prompt: "gh issue view {{.issue}} --comments"}
      - {id: review-merge, pool: refinery, retries: 2, title: "Merge #{{.issue}}", prompt: merge}
      - {id: post-merge, pool: deploy-check, retries: 2, title: "Deploy #{{.issue}}", prompt: deploy}
  drill:
    vars: [issue]
    issue: issue
    repo: elee1766/jobshop
    escalate_pool: plan
    steps:
      - {id: fail, pool: dev, retries: 1, title: "Drill #{{.issue}}", prompt: "jobshop close -r failed"}
`), 0o600))
	cfg := fmt.Sprintf(`projects:
  js: {}
agents:
  fake:
    command: %q
packs:
  - path: %s
    vars: {checkout: %q, agent: fake}
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
	return dial(t, root.AdminSocket())
}
