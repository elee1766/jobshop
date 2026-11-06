package control

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/learning"
	"github.com/elee1766/jobshop/pkg/memedit"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

var gateT0 = time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)

func fxStep(id, run, pool string, at time.Time, failures int, failed bool) store.Ticket {
	t := store.Ticket{
		ID: id, Kind: store.KindStep, Pool: pool, ParentID: run, CreatedAt: at,
		Fields: json.RawMessage(fmt.Sprintf(`{"run":%q,"step":"a","failures":%d}`, run, failures)),
	}
	if failed {
		t.Status, t.Resolution = store.StatusClosed, "failed"
	}
	return t
}

// TestGateMetrics: metrics over fixture steps and learnings count only
// the window and the scoped pools; recurrence matches the cited
// learnings by kind and wording.
func TestGateMetrics(t *testing.T) {
	t.Parallel()
	h := time.Hour
	steps := []store.Ticket{
		fxStep("s1", "r1", "dev", gateT0.Add(h), 2, false),
		fxStep("s2", "r1", "dev", gateT0.Add(2*h), 0, true),
		fxStep("s3", "r2", "dev", gateT0.Add(3*h), 1, false),
		fxStep("s4", "r3", "refinery", gateT0.Add(3*h), 5, true), // other pool
		fxStep("s5", "r4", "dev", gateT0.Add(30*h), 3, false),    // outside window
	}
	cited := []store.Learning{{ID: 1, Kind: store.LearningFailure, Summary: "golangci-lint failed: unused parameter ctx in handler", CreatedAt: gateT0.Add(-h)}}
	ls := []store.Learning{
		cited[0], // before the window
		{ID: 2, Kind: store.LearningFailure, Summary: "golangci-lint failed again: unused parameter ctx in handler", CreatedAt: gateT0.Add(h)},
		{ID: 3, Kind: store.LearningCorrection, Summary: "golangci-lint failed: unused parameter ctx in handler", CreatedAt: gateT0.Add(h)}, // other kind
		{ID: 4, Kind: store.LearningFailure, Summary: "nix build hash mismatch", CreatedAt: gateT0.Add(h)},
		{ID: 5, Kind: store.LearningCorrection, Summary: "review resumed step a (s3): missing test", Ticket: "s3", CreatedAt: gateT0.Add(4 * h)},
		{ID: 6, Kind: store.LearningCorrection, Summary: "review resumed step a (s4): wrong pool", Ticket: "s4", CreatedAt: gateT0.Add(4 * h)},
	}
	m := gateMetrics(steps, ls, cited, []string{"dev"}, gateT0, gateT0.Add(24*h))
	assert.Equal(t, 2, m.Runs)
	assert.Equal(t, 3, m.Steps)
	assert.Equal(t, 3, m.Retries)
	assert.Equal(t, 1, m.Escalations)
	assert.Equal(t, 1, m.Findings, "only findings on steps of the scoped pools")
	assert.Equal(t, 1, m.Recurrence, "same kind and similar summary")

	all := gateMetrics(steps, ls, cited, nil, gateT0, gateT0.Add(24*h))
	assert.Equal(t, 4, all.Steps, "no pools: every pool")
	assert.Equal(t, 2, all.Findings)
}

func TestGateDecide(t *testing.T) {
	t.Parallel()
	base := store.PoolMetrics{Steps: 20, Retries: 4, Escalations: 1, Findings: 1, Recurrence: 3}
	for _, tc := range []struct {
		name  string
		after store.PoolMetrics
		want  string
	}{
		{"no recurrence, same failures", store.PoolMetrics{Steps: 20, Retries: 4, Escalations: 1, Findings: 1}, store.EditHelped},
		{"less recurrence", store.PoolMetrics{Steps: 20, Retries: 3, Recurrence: 1}, store.EditHelped},
		{"no steps since merge", store.PoolMetrics{}, store.EditHelped},
		{"recurs as often", store.PoolMetrics{Steps: 20, Retries: 4, Recurrence: 3}, store.EditNotHelped},
		{"recurs more per step", store.PoolMetrics{Steps: 10, Recurrence: 2}, store.EditNotHelped},
		{"failures rose", store.PoolMetrics{Steps: 20, Retries: 9, Escalations: 2, Findings: 2}, store.EditNotHelped},
		{"failures within noise", store.PoolMetrics{Steps: 20, Retries: 6, Escalations: 1, Findings: 0}, store.EditHelped},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, reason := gateDecide(base, tc.after)
			assert.Equal(t, tc.want, got, reason)
			assert.NotEmpty(t, reason)
		})
	}
	got, _ := gateDecide(store.PoolMetrics{}, store.PoolMetrics{Steps: 5, Recurrence: 1})
	assert.Equal(t, store.EditNotHelped, got, "recurs after none in the baseline")
}

// TestCuratorGate: applying a curator PR records a memory edit with a
// baseline, the edits trigger condition waits for runs or age, and
// evaluate writes the result on the edit and its learnings.
func TestCuratorGate(t *testing.T) {
	t.Parallel()
	c, clk, _ := stallController(t)
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	w := config.EditsWhen{Runs: 2, Age: "72h"}

	why, err := c.editsHold(ctx, w, clk.now())
	require.NoError(t, err)
	assert.Equal(t, "no pending memory edits", why)

	// Before the merge: a dev step failing twice, and the learning cited.
	step := mkTicket(t, c, store.NewTicket{Kind: store.KindTask, Title: "x", Pool: "dev"})
	l1, err := c.learnAdd(ctx, admin, rpc.LearnParams{Kind: store.LearningFailure, Summary: "go vet failed on shadowed err variable", Ticket: step.ID})
	require.NoError(t, err)
	id1 := l1.(store.Learning).ID
	clk.advance(time.Hour)

	const pr = "https://github.com/elee1766/jobshop/pull/900"
	_, err = c.learnUpdate(ctx, admin, rpc.LearnUpdateParams{IDs: []int64{id1}, PR: pr})
	require.NoError(t, err)
	edits, err := memedit.List(ctx, c.store, "", 0)
	require.NoError(t, err)
	assert.Empty(t, edits, "a pending PR is not an edit yet")

	merged := clk.now()
	_, err = c.learnUpdate(ctx, admin, rpc.LearnUpdateParams{IDs: []int64{id1}, Status: store.LearningApplied, MergedAt: merged})
	require.NoError(t, err)
	edits, err = memedit.List(ctx, c.store, store.EditPending, 0)
	require.NoError(t, err)
	require.Len(t, edits, 1)
	e := edits[0]
	assert.Equal(t, pr, e.PR)
	assert.Equal(t, []int64{id1}, e.Learnings)
	assert.Equal(t, []string{"dev"}, e.Pools)
	assert.Equal(t, 1, e.Baseline.Recurrence)
	assert.True(t, e.MergedAt.Equal(merged))

	// Applying again does not record a second edit.
	_, err = c.learnUpdate(ctx, admin, rpc.LearnUpdateParams{IDs: []int64{id1}, Status: store.LearningApplied})
	require.NoError(t, err)
	edits, err = memedit.List(ctx, c.store, "", 0)
	require.NoError(t, err)
	assert.Len(t, edits, 1)

	clk.advance(time.Hour)
	why, err = c.editsHold(ctx, w, clk.now())
	require.NoError(t, err)
	assert.Equal(t, "1 pending memory edits, none due", why)
	out, err := c.learnEvaluate(ctx, admin, rpc.LearnEvaluateParams{Runs: 2, Age: 72 * time.Hour})
	require.NoError(t, err)
	assert.Empty(t, out, "not due")

	// Two dev runs since the merge, and the same failure recurs.
	for i := range 2 {
		res, err := c.startRun(ctx, c.config(), admin, rpc.RunStartParams{Workflow: "work", Vars: map[string]string{"n": fmt.Sprint(i)}, Key: fmt.Sprint("g", i)}, nil)
		require.NoError(t, err)
		if i == 0 {
			s := res.(rpc.RunResult).Steps[0]
			_, err = c.learnAdd(ctx, admin, rpc.LearnParams{Kind: store.LearningFailure, Summary: "go vet failed again on shadowed err variable", Ticket: s.ID})
			require.NoError(t, err)
		}
	}
	clk.advance(time.Minute)
	why, err = c.editsHold(ctx, w, clk.now())
	require.NoError(t, err)
	assert.Empty(t, why, "due after 2 runs")

	out, err = c.learnEvaluate(ctx, admin, rpc.LearnEvaluateParams{Runs: 2, Age: 72 * time.Hour})
	require.NoError(t, err)
	res := out.([]store.MemoryEdit)
	require.Len(t, res, 1)
	assert.Equal(t, store.EditNotHelped, res[0].State)
	assert.Equal(t, 2, res[0].After.Runs)
	assert.Equal(t, 1, res[0].After.Recurrence)
	assert.Contains(t, res[0].Reason, "still recurs")

	l, err := learning.Get(ctx, c.store, id1)
	require.NoError(t, err)
	assert.Contains(t, l.Gate, "not_helped: ")
	assert.Contains(t, l.Gate, pr)

	why, err = c.editsHold(ctx, w, clk.now())
	require.NoError(t, err)
	assert.Equal(t, "no pending memory edits", why, "evaluated edits leave the trigger")
	_, err = c.learnEvaluate(ctx, admin, rpc.LearnEvaluateParams{ID: e.ID})
	require.ErrorIs(t, err, store.ErrNotFound)
}

// TestCuratorGateAge: an edit with no runs is due after age, and with no
// steps since the merge it counts as helped.
func TestCuratorGateAge(t *testing.T) {
	t.Parallel()
	c, clk, _ := stallController(t)
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	l, err := c.learnAdd(ctx, admin, rpc.LearnParams{Kind: store.LearningWorkaround, Summary: "retry gh on 502", Scope: "pool:dev"})
	require.NoError(t, err)
	id := l.(store.Learning).ID
	_, err = c.learnUpdate(ctx, admin, rpc.LearnUpdateParams{IDs: []int64{id}, Status: store.LearningApplied, PR: "https://github.com/elee1766/jobshop/pull/901"})
	require.NoError(t, err)

	w := config.EditsWhen{Age: "72h"}
	clk.advance(71 * time.Hour)
	why, err := c.editsHold(ctx, w, clk.now())
	require.NoError(t, err)
	assert.NotEmpty(t, why)
	clk.advance(time.Hour)
	why, err = c.editsHold(ctx, w, clk.now())
	require.NoError(t, err)
	assert.Empty(t, why)

	out, err := c.learnEvaluate(ctx, admin, rpc.LearnEvaluateParams{})
	require.NoError(t, err)
	res := out.([]store.MemoryEdit)
	require.Len(t, res, 1)
	assert.Equal(t, store.EditHelped, res[0].State)
	got, err := learning.Get(ctx, c.store, id)
	require.NoError(t, err)
	assert.Contains(t, got.Gate, "helped: no steps")
}
