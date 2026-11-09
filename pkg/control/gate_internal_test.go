package control

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/knowledge"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// TestCuratorGate: applying a curator PR records a memory edit with a
// baseline, the edits trigger condition waits for runs or age, and
// evaluate writes the result on the edit and its learnings.
func TestCuratorGate(t *testing.T) {
	t.Parallel()
	c, clk, _ := scanController(t)
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	w := config.EditsWhen{Runs: 2, Age: "72h"}

	why, err := c.editsHold(ctx, w, clk.now())
	require.NoError(t, err)
	assert.Equal(t, "no pending memory edits", why)

	// Before the merge: a dev step failing twice, and the learning cited.
	step := mkTicket(t, c, store.NewTicket{Kind: domain.KindTask, Title: "x", Pool: "dev"})
	l1, err := c.learnAdd(ctx, admin, rpc.LearnParams{Kind: domain.LearningFailure, Summary: "go vet failed on shadowed err variable", Ticket: step.ID})
	require.NoError(t, err)
	id1 := l1.(domain.Learning).ID
	clk.advance(time.Hour)

	const pr = "https://github.com/elee1766/jobshop/pull/900"
	_, err = c.learnUpdate(ctx, admin, rpc.LearnUpdateParams{IDs: []int64{id1}, PR: pr})
	require.NoError(t, err)
	edits, err := knowledge.ListEdits(ctx, c.store, "", 0)
	require.NoError(t, err)
	assert.Empty(t, edits, "a pending PR is not an edit yet")

	merged := clk.now()
	_, err = c.learnUpdate(ctx, admin, rpc.LearnUpdateParams{IDs: []int64{id1}, Status: domain.LearningApplied, MergedAt: merged})
	require.NoError(t, err)
	edits, err = knowledge.ListEdits(ctx, c.store, domain.EditPending, 0)
	require.NoError(t, err)
	require.Len(t, edits, 1)
	e := edits[0]
	assert.Equal(t, pr, e.PR)
	assert.Equal(t, []int64{id1}, e.Learnings)
	assert.Equal(t, []string{"dev"}, e.Pools)
	assert.Equal(t, 1, e.Baseline.Recurrence)
	assert.True(t, e.MergedAt.Equal(merged))

	// Applying again does not record a second edit.
	_, err = c.learnUpdate(ctx, admin, rpc.LearnUpdateParams{IDs: []int64{id1}, Status: domain.LearningApplied})
	require.NoError(t, err)
	edits, err = knowledge.ListEdits(ctx, c.store, "", 0)
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
			_, err = c.learnAdd(ctx, admin, rpc.LearnParams{Kind: domain.LearningFailure, Summary: "go vet failed again on shadowed err variable", Ticket: s.ID})
			require.NoError(t, err)
		}
	}
	clk.advance(time.Minute)
	why, err = c.editsHold(ctx, w, clk.now())
	require.NoError(t, err)
	assert.Empty(t, why, "due after 2 runs")

	out, err = c.learnEvaluate(ctx, admin, rpc.LearnEvaluateParams{Runs: 2, Age: 72 * time.Hour})
	require.NoError(t, err)
	res := out.([]domain.MemoryEdit)
	require.Len(t, res, 1)
	assert.Equal(t, domain.EditNotHelped, res[0].State)
	assert.Equal(t, 2, res[0].After.Runs)
	assert.Equal(t, 1, res[0].After.Recurrence)
	assert.Contains(t, res[0].Reason, "still recurs")

	l, err := knowledge.GetLearning(ctx, c.store, id1)
	require.NoError(t, err)
	assert.Contains(t, l.Gate, "not_helped: ")
	assert.Contains(t, l.Gate, pr)

	why, err = c.editsHold(ctx, w, clk.now())
	require.NoError(t, err)
	assert.Equal(t, "no pending memory edits", why, "evaluated edits leave the trigger")
	_, err = c.learnEvaluate(ctx, admin, rpc.LearnEvaluateParams{ID: e.ID})
	require.ErrorIs(t, err, domain.ErrNotFound)
}

// TestCuratorGateAge: an edit with no runs is due after age, and with no
// steps since the merge it counts as helped.
func TestCuratorGateAge(t *testing.T) {
	t.Parallel()
	c, clk, _ := scanController(t)
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	l, err := c.learnAdd(ctx, admin, rpc.LearnParams{Kind: domain.LearningWorkaround, Summary: "retry gh on 502", Scope: "pool:dev"})
	require.NoError(t, err)
	id := l.(domain.Learning).ID
	_, err = c.learnUpdate(ctx, admin, rpc.LearnUpdateParams{IDs: []int64{id}, Status: domain.LearningApplied, PR: "https://github.com/elee1766/jobshop/pull/901"})
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
	res := out.([]domain.MemoryEdit)
	require.Len(t, res, 1)
	assert.Equal(t, domain.EditHelped, res[0].State)
	got, err := knowledge.GetLearning(ctx, c.store, id)
	require.NoError(t, err)
	assert.Contains(t, got.Gate, "helped: no steps")
}
