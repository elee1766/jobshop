package control

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/knowledge"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/support/telemetry"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

func learnings(t *testing.T, c *Controller, f store.LearningFilter) []domain.Learning {
	t.Helper()
	ls, err := knowledge.ListLearnings(t.Context(), c.store, f)
	require.NoError(t, err)
	return ls
}

// startWork starts a work run and claims its first step for a dev session.
func startWork(t *testing.T, c *Controller, key string) (domain.Ticket, domain.Ticket, scope) {
	t.Helper()
	ctx := t.Context()
	if c.metrics == nil {
		m, err := telemetry.NewMetrics()
		require.NoError(t, err)
		c.metrics = m
	}
	res, err := c.startRun(ctx, c.config(), scope{admin: true, actor: "admin"}, rpc.RunStartParams{Workflow: "work", Vars: map[string]string{"n": "1"}, Key: key}, nil)
	require.NoError(t, err)
	step := res.(rpc.RunResult).Steps[0]
	sess := mkTicket(t, c, store.NewTicket{Kind: domain.KindSession, Pool: "dev", Fields: json.RawMessage(`{"agent":"fake"}`)})
	_, _, err = c.store.Claim(ctx, step.ID, sess.ID, time.Hour)
	require.NoError(t, err)
	return step, sess, scope{actor: sess.ID, session: sess.ID, project: "js", pool: "dev"}
}

func TestLearnRPC(t *testing.T) {
	t.Parallel()
	c, _, _ := scanController(t)
	ctx := t.Context()
	step, sess, sc := startWork(t, c, "l1")

	_, err := c.learnAdd(ctx, sc, rpc.LearnParams{Kind: "oops", Summary: "x"})
	require.ErrorIs(t, err, domain.ErrInvalid)
	_, err = c.learnAdd(ctx, sc, rpc.LearnParams{Kind: domain.LearningCorrection, Summary: "  "})
	require.ErrorIs(t, err, domain.ErrInvalid)
	_, err = c.learnAdd(ctx, sc, rpc.LearnParams{Kind: domain.LearningCorrection, Summary: "x", Scope: "team:eng"})
	require.ErrorIs(t, err, domain.ErrInvalid)
	_, err = c.learnAdd(ctx, sc, rpc.LearnParams{Kind: domain.LearningCorrection, Summary: "x", Ticket: "js-nope"})
	require.ErrorIs(t, err, domain.ErrNotFound)

	// Defaults: the held ticket, its run's repo, the session, source agent.
	out, err := c.learnAdd(ctx, sc, rpc.LearnParams{Kind: domain.LearningCorrection, Summary: "use main as base", Fix: "git switch main"})
	require.NoError(t, err)
	l := out.(domain.Learning)
	assert.Equal(t, step.ID, l.Ticket)
	assert.Equal(t, sess.ID, l.Session)
	assert.Equal(t, "repo:acme/widget", l.Scope)
	assert.Equal(t, domain.LearningSourceAgent, l.Source)
	assert.Equal(t, domain.LearningOpen, l.Status)

	// An explicit scope wins; admin with no ticket is global.
	_, err = c.learnAdd(ctx, sc, rpc.LearnParams{Kind: domain.LearningWorkaround, Summary: "retry go mod download", Scope: "pool:dev"})
	require.NoError(t, err)
	out, err = c.learnAdd(ctx, scope{admin: true, actor: "admin"}, rpc.LearnParams{Kind: domain.LearningFailure, Summary: "host disk full"})
	require.NoError(t, err)
	assert.Equal(t, domain.ScopeGlobal, out.(domain.Learning).Scope)

	res, err := c.learnList(ctx, sc, rpc.LearnListParams{})
	require.NoError(t, err)
	assert.Len(t, res.([]domain.Learning), 3)
	res, err = c.learnList(ctx, sc, rpc.LearnListParams{Scope: "pool:dev"})
	require.NoError(t, err)
	require.Len(t, res.([]domain.Learning), 1)
	assert.Equal(t, "retry go mod download", res.([]domain.Learning)[0].Summary)
	res, err = c.learnList(ctx, sc, rpc.LearnListParams{Kind: domain.LearningFailure})
	require.NoError(t, err)
	assert.Len(t, res.([]domain.Learning), 1)
	_, err = c.learnList(ctx, sc, rpc.LearnListParams{Scope: "repo:bad"})
	require.ErrorIs(t, err, domain.ErrInvalid)
}

func TestLearnSince(t *testing.T) {
	t.Parallel()
	c, clk, _ := scanController(t)
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	_, err := c.learnAdd(ctx, admin, rpc.LearnParams{Kind: domain.LearningFailure, Summary: "old"})
	require.NoError(t, err)
	clk.advance(48 * time.Hour)
	_, err = c.learnAdd(ctx, admin, rpc.LearnParams{Kind: domain.LearningFailure, Summary: "new"})
	require.NoError(t, err)
	res, err := c.learnList(ctx, admin, rpc.LearnListParams{Since: 24 * time.Hour})
	require.NoError(t, err)
	require.Len(t, res.([]domain.Learning), 1)
	assert.Equal(t, "new", res.([]domain.Learning)[0].Summary)
}

func TestLearnStepFailure(t *testing.T) {
	t.Parallel()
	c, _, _ := scanController(t)
	ctx := t.Context()
	step, _, sc := startWork(t, c, "f1")
	_, err := c.report(ctx, sc, rpc.ReportParams{State: rpc.ReportFailed, Note: "tests flake on CI"})
	require.NoError(t, err)
	_, err = c.closeTicket(ctx, sc, rpc.CloseParams{Resolution: "failed"})
	require.NoError(t, err)
	ls := learnings(t, c, store.LearningFilter{Ticket: step.ID})
	require.Len(t, ls, 1)
	assert.Equal(t, domain.LearningFailure, ls[0].Kind)
	assert.Equal(t, domain.LearningSourceController, ls[0].Source)
	assert.Equal(t, "repo:acme/widget", ls[0].Scope)
	assert.Contains(t, ls[0].Summary, "closed failed")
	assert.Contains(t, ls[0].Summary, "tests flake on CI")
	assert.Contains(t, ls[0].Summary, "step a")
}

func TestLearnRelease(t *testing.T) {
	t.Parallel()
	c, clk, _ := scanController(t)
	ctx := t.Context()
	step, _, sc := startWork(t, c, "r1")
	until := clk.now().Add(time.Hour) // release checks the controller clock
	_, err := c.release(ctx, sc, rpc.ReleaseParams{Reason: "CI later", Until: &until, TimeWait: true})
	require.NoError(t, err)
	assert.Empty(t, learnings(t, c, store.LearningFilter{Ticket: step.ID}), "a deferred release is not a failure")

	step, _, sc = startWork(t, c, "r2")
	_, err = c.release(ctx, sc, rpc.ReleaseParams{Reason: "gave up"})
	require.NoError(t, err)
	ls := learnings(t, c, store.LearningFilter{Ticket: step.ID})
	require.Len(t, ls, 1)
	assert.Contains(t, ls[0].Summary, "released (failure 1/1): gave up")
}

func TestLearnReviewFinding(t *testing.T) {
	t.Parallel()
	c, _, _ := scanController(t)
	ctx := t.Context()
	step, _, sc := startWork(t, c, "v1")
	_, err := c.closeTicket(ctx, sc, rpc.CloseParams{})
	require.NoError(t, err)

	// The reviewer holds its own ticket and comments the finding there.
	rev := mkTicket(t, c, store.NewTicket{Title: "review", Pool: "plan"})
	rs := mkTicket(t, c, store.NewTicket{Kind: domain.KindSession, Pool: "plan", Fields: json.RawMessage(`{"agent":"fake"}`)})
	_, _, err = c.store.Claim(ctx, rev.ID, rs.ID, time.Hour)
	require.NoError(t, err)
	_, err = store.AddComment(ctx, c.store, rs.ID, rev.ID, "missing tests for the error path")
	require.NoError(t, err)
	rsc := scope{actor: rs.ID, session: rs.ID, project: "js", pool: "plan"}
	_, err = c.resumeRun(ctx, rsc, rpc.RunResumeParams{ID: step.ParentID, Step: "a"})
	require.NoError(t, err)

	ls := learnings(t, c, store.LearningFilter{Ticket: step.ID})
	require.Len(t, ls, 1)
	assert.Equal(t, domain.LearningCorrection, ls[0].Kind)
	assert.Contains(t, ls[0].Summary, "review resumed step a")
	assert.Contains(t, ls[0].Summary, "missing tests for the error path")
}

func TestLearnEscalation(t *testing.T) {
	t.Parallel()
	c, _, _ := scanController(t)
	ctx := t.Context()
	step, _, sc := startWork(t, c, "e1")
	_, err := c.closeTicket(ctx, sc, rpc.CloseParams{Resolution: "failed"}) // retries: 1
	require.NoError(t, err)
	ls := learnings(t, c, store.LearningFilter{Ticket: step.ParentID})
	require.Len(t, ls, 1, "the escalation is one learning on the run")
	assert.Contains(t, ls[0].Summary, "escalated to a human")
	assert.Equal(t, "repo:acme/widget", ls[0].Scope)
	assert.Len(t, learnings(t, c, store.LearningFilter{Ticket: step.ID}), 1, "plus the step failure")
}

func TestLearnAnswer(t *testing.T) {
	t.Parallel()
	c, _, _ := scanController(t)
	ctx := t.Context()
	step, _, sc := startWork(t, c, "q1")
	out, err := c.report(ctx, sc, rpc.ReportParams{State: rpc.ReportWaiting, Question: "Which base branch?"})
	require.NoError(t, err)
	q := out.(rpc.ReportResult).Question
	_, err = c.answerQuestion(ctx, scope{admin: true, actor: "admin"}, rpc.AnswerParams{ID: q.ID, Text: "use main", From: "@op"})
	require.NoError(t, err)
	ls := learnings(t, c, store.LearningFilter{Ticket: step.ID})
	require.Len(t, ls, 1)
	assert.Equal(t, domain.LearningCorrection, ls[0].Kind)
	assert.Contains(t, ls[0].Summary, "Which base branch?")
	assert.Equal(t, "answer from @op: use main", ls[0].Fix)
}

func TestLearnFinding(t *testing.T) {
	t.Parallel()
	c, _, dir := scanController(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "issues"), []byte("issue-7 no run\n"), 0o600))
	c.runTriggers(t.Context(), c.config())
	ls := learnings(t, c, store.LearningFilter{Kind: domain.LearningFailure})
	require.Len(t, ls, 1)
	assert.Contains(t, ls[0].Summary, "finding issue_no_run on issue-7: no run")
	assert.Equal(t, domain.LearningSourceController, ls[0].Source)
}

// TestLearnUpdate: open -> pr pending -> applied or rejected (#273).
func TestLearnUpdate(t *testing.T) {
	t.Parallel()
	c, _, _ := scanController(t)
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	add := func() int64 {
		out, err := c.learnAdd(ctx, admin, rpc.LearnParams{Kind: domain.LearningFailure, Summary: "x"})
		require.NoError(t, err)
		return out.(domain.Learning).ID
	}
	a, b := add(), add()
	pr := "https://github.com/elee1766/jobshop/pull/9"

	_, err := c.learnUpdate(ctx, admin, rpc.LearnUpdateParams{})
	require.ErrorIs(t, err, domain.ErrInvalid)
	_, err = c.learnUpdate(ctx, admin, rpc.LearnUpdateParams{IDs: []int64{a}})
	require.ErrorIs(t, err, domain.ErrInvalid, "needs status or pr")
	_, err = c.learnUpdate(ctx, admin, rpc.LearnUpdateParams{IDs: []int64{a}, Status: domain.LearningRejected})
	require.ErrorIs(t, err, domain.ErrInvalid, "rejected needs a pr")
	_, err = c.learnUpdate(ctx, admin, rpc.LearnUpdateParams{IDs: []int64{a}, Status: "done"})
	require.ErrorIs(t, err, domain.ErrInvalid)
	_, err = c.learnUpdate(ctx, admin, rpc.LearnUpdateParams{IDs: []int64{999}, PR: pr})
	require.ErrorIs(t, err, domain.ErrNotFound)

	out, err := c.learnUpdate(ctx, admin, rpc.LearnUpdateParams{IDs: []int64{a, b}, PR: pr})
	require.NoError(t, err)
	for _, l := range out.([]domain.Learning) {
		assert.Equal(t, domain.LearningOpen, l.Status, "pr pending keeps it open")
		assert.Equal(t, pr, l.PR)
	}
	out, err = c.learnUpdate(ctx, admin, rpc.LearnUpdateParams{IDs: []int64{a}, Status: domain.LearningApplied})
	require.NoError(t, err)
	assert.Equal(t, pr, out.([]domain.Learning)[0].PR, "pr kept")
	_, err = c.learnUpdate(ctx, admin, rpc.LearnUpdateParams{IDs: []int64{b}, Status: domain.LearningRejected})
	require.NoError(t, err, "pr already set")

	rej := learnings(t, c, store.LearningFilter{Status: domain.LearningRejected})
	require.Len(t, rej, 1)
	assert.Equal(t, b, rej[0].ID)
	assert.Equal(t, pr, rej[0].PR)
	n, err := knowledge.CountLearnings(ctx, c.store, time.Time{}, domain.LearningCount{Status: domain.LearningOpen})
	require.NoError(t, err)
	assert.Zero(t, n)
}
