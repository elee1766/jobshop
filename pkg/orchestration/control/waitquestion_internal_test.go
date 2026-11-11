package control

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/orchestration/workflow"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/support/telemetry"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// questionStep claims a step for a new session and has it report waiting
// with a question. wait parks it on the question in the same call.
func questionStep(t *testing.T, c *Controller, wait bool) (step, sess domain.Ticket, q domain.Ticket) {
	t.Helper()
	ctx := t.Context()
	run := mkTicket(t, c, store.NewTicket{Kind: domain.KindRun, Title: "r", Fields: json.RawMessage(`{"workflow":"w","state":"running"}`)})
	step = mkTicket(t, c, store.NewTicket{
		Kind: domain.KindStep, Title: "await", Pool: "dev", ParentID: run.ID,
		Fields: json.RawMessage(`{"run":"` + run.ID + `","step":"await","index":0,"retries":3,"failures":0}`),
	})
	sess = mkSession(t, c)
	_, _, err := c.store.Claim(ctx, step.ID, sess.ID, time.Hour)
	require.NoError(t, err)
	out, err := c.report(ctx, sessScope(sess), rpc.ReportParams{
		State: rpc.ReportWaiting, Note: "sign-off", Question: "Approve the funds-risk change?",
		Context: "Reviewing the change.\nIt moves funds.", Wait: wait,
	})
	require.NoError(t, err)
	require.NotNil(t, out.(rpc.ReportResult).Question)
	return step, sess, *out.(rpc.ReportResult).Question
}

func parkOnQuestion(t *testing.T, c *Controller, step, sess domain.Ticket, q domain.Ticket, conds ...domain.WaitCond) {
	t.Helper()
	_, err := c.waitAdd(t.Context(), sessScope(sess), rpc.WaitAddParams{
		ID: step.ID, Conds: append([]domain.WaitCond{{Type: domain.WaitQuestion, Question: q.ID}}, conds...),
		Any: len(conds) > 0, Resume: true, Reason: "sign-off " + q.ID,
	})
	require.NoError(t, err)
}

// pumpQuestionWake runs what the wait pump does after a targeted wake.
func pumpQuestionWake(t *testing.T, c *Controller, prs *fakePRs, ths *fakeThreads) {
	t.Helper()
	w := c.waitWakes.Take()
	require.False(t, w.All)
	c.pollWaitsOn(t.Context(), prs, ths, w.Subjects)
}

// TestWaitQuestionAnswered (JOB-202): a step parks on its question with
// resume; the answer wakes the wait at once, clears it and hands the step
// back to the parking session. Nothing else changes first.
func TestWaitQuestionAnswered(t *testing.T) {
	t.Parallel()
	c, _, prs, threads := waitController(t)
	ctx := t.Context()
	step, sess, q := questionStep(t, c, false)
	parkOnQuestion(t, c, step, sess, q)
	got := mustGet(t, c, step.ID)
	require.NotNil(t, waitOf(got))
	assert.Equal(t, domain.StatusOpen, got.Status)
	assert.Empty(t, readyIDs(t, c), "parked")
	c.waitWakes.Take()

	// A full pass with the question open changes nothing.
	c.pollWaits(ctx, prs, threads)
	require.NotNil(t, waitOf(mustGet(t, c, step.ID)))

	_, err := c.answerQuestion(ctx, scope{admin: true, actor: "admin"}, rpc.AnswerParams{ID: q.ID, Text: "approved, go ahead"})
	require.NoError(t, err)
	w := c.waitWakes.Take()
	assert.True(t, w.Subjects[workflow.QuestionSubject(q.ID)], "closing the question wakes its waits")
	c.pollWaitsOn(ctx, prs, threads, w.Subjects)
	got = mustGet(t, c, step.ID)
	assert.Nil(t, waitOf(got))
	assert.Equal(t, sess.ID, got.Holder, "resume hands the step to the parking session")
	assert.Equal(t, []string{step.ID}, readyIDs(t, c))
	cs, err := store.Comments(ctx, c.store, step.ID)
	require.NoError(t, err)
	var bodies []string
	for _, cm := range cs {
		bodies = append(bodies, cm.Body)
	}
	assert.Contains(t, bodies[len(bodies)-1], "question "+q.ID+" closed (answered)")
	assert.Contains(t, bodies, "answer to question "+q.ID+" from admin: approved, go ahead", "the answer stays a comment on the step")
	sf, err := domain.DecodeFields[domain.StepFields](got)
	require.NoError(t, err)
	assert.Equal(t, 0, sf.Failures)
}

// TestWaitQuestionAny (JOB-202): with --any and a PR head change, the
// answer alone clears the wait.
func TestWaitQuestionAny(t *testing.T) {
	t.Parallel()
	c, _, prs, threads := waitController(t)
	ctx := t.Context()
	prs.set(recPR, `[[]]`, `[]`)
	step, sess, q := questionStep(t, c, false)
	parkOnQuestion(t, c, step, sess, q, domain.WaitCond{Type: domain.WaitPR, Repo: "acme/widget", PR: 5, State: "head_changed"})
	c.waitWakes.Take()
	c.pollWaits(ctx, prs, threads)
	w := waitOf(mustGet(t, c, step.ID))
	require.NotNil(t, w, "neither holds")
	assert.Equal(t, domain.WaitAny, w.Mode)

	_, err := c.answerQuestion(ctx, scope{admin: true, actor: "admin"}, rpc.AnswerParams{ID: q.ID, Text: "ok"})
	require.NoError(t, err)
	pumpQuestionWake(t, c, prs, threads)
	got := mustGet(t, c, step.ID)
	assert.Nil(t, waitOf(got))
	assert.Equal(t, sess.ID, got.Holder)
}

// TestWaitQuestionDismissed (JOB-202): a question closed without an
// answer (dismissed or cancelled through close, withdrawn by the sweep)
// also clears the wait.
func TestWaitQuestionDismissed(t *testing.T) {
	t.Parallel()
	c, _, prs, threads := waitController(t)
	ctx := t.Context()
	m, err := telemetry.NewMetrics()
	require.NoError(t, err)
	c.metrics = m
	admin := scope{admin: true, actor: "admin"}
	for _, resolution := range []string{"dismissed", "cancelled", "withdrawn"} {
		step, sess, q := questionStep(t, c, false)
		parkOnQuestion(t, c, step, sess, q)
		c.waitWakes.Take()
		if resolution == "withdrawn" {
			// The asked ticket going away withdraws the question.
			c.closeDerived(ctx, q, resolutionWithdrawn, "asked ticket closed", true)
		} else {
			_, err := c.closeTicket(ctx, admin, rpc.CloseParams{ID: q.ID, Resolution: resolution})
			require.NoError(t, err)
		}
		pumpQuestionWake(t, c, prs, threads)
		got := mustGet(t, c, step.ID)
		assert.Nil(t, waitOf(got), resolution)
		cs, err := store.Comments(ctx, c.store, step.ID)
		require.NoError(t, err)
		assert.Contains(t, cs[len(cs)-1].Body, "question "+q.ID+" closed ("+resolution+")")
	}
}

// TestWaitQuestionValidate (JOB-202): the id must be a question ticket in
// the caller's project; a question that is already closed is accepted and
// the wait clears on its first evaluation.
func TestWaitQuestionValidate(t *testing.T) {
	t.Parallel()
	c, _, prs, threads := waitController(t)
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	step, sess, q := questionStep(t, c, false)
	add := func(sc scope, id string) error {
		_, err := c.waitAdd(ctx, sc, rpc.WaitAddParams{ID: step.ID, Reason: "q", Conds: []domain.WaitCond{{Type: domain.WaitQuestion, Question: id}}})
		return err
	}

	require.ErrorIs(t, add(admin, ""), domain.ErrInvalid, "no id")
	require.ErrorIs(t, add(admin, "js-nope"), domain.ErrInvalid, "not a ticket")
	require.ErrorIs(t, add(admin, step.ID), domain.ErrInvalid, "a ticket, not a question")
	assert.ErrorContains(t, add(admin, step.ID), "is not a question")
	assert.Nil(t, waitOf(mustGet(t, c, step.ID)), "refused waits park nothing")

	other := mkTicket(t, c, store.NewTicket{
		Project: "other", Kind: domain.KindMessage, Title: "Question: x",
		Fields: json.RawMessage(`{"to":"` + questionTo + `"}`),
		Meta:   json.RawMessage(`{"question":{"ticket":"` + step.ID + `"}}`),
	})
	require.ErrorIs(t, add(sessScope(sess), other.ID), rpc.ErrForbidden, "another project's question")
	require.ErrorIs(t, add(admin, other.ID), rpc.ErrForbidden, "even for admin, the wait is for this ticket's project")
	assert.Nil(t, waitOf(mustGet(t, c, step.ID)))

	_, err := c.answerQuestion(ctx, admin, rpc.AnswerParams{ID: q.ID, Text: "done"})
	require.NoError(t, err)
	c.waitWakes.Take()
	require.NoError(t, add(sessScope(sess), q.ID))
	require.NotNil(t, waitOf(mustGet(t, c, step.ID)))
	c.pollWaits(ctx, prs, threads)
	assert.Nil(t, waitOf(mustGet(t, c, step.ID)), "already closed: clears on the first evaluation")
}

// TestReportWaitingWait (JOB-202): report waiting -q --wait files the
// question and parks the step on it in one call, released without a failed
// attempt; the answer resumes it.
func TestReportWaitingWait(t *testing.T) {
	t.Parallel()
	c, _, prs, threads := waitController(t)
	ctx := t.Context()
	step, sess, q := questionStep(t, c, true)
	got := mustGet(t, c, step.ID)
	w := waitOf(got)
	require.NotNil(t, w)
	assert.Equal(t, []domain.WaitCond{{Type: domain.WaitQuestion, Question: q.ID}}, w.Conds)
	assert.True(t, w.Resume)
	assert.Equal(t, sess.ID, w.Session)
	assert.Empty(t, got.Holder, "released")
	assert.Equal(t, domain.StatusOpen, got.Status)
	c.waitWakes.Take() // the park queued a full pass

	_, err := c.answerQuestion(ctx, scope{admin: true, actor: "admin"}, rpc.AnswerParams{ID: q.ID, Text: "yes"})
	require.NoError(t, err)
	pumpQuestionWake(t, c, prs, threads)
	got = mustGet(t, c, step.ID)
	assert.Nil(t, waitOf(got))
	assert.Equal(t, sess.ID, got.Holder)

	// Refusals leave no question behind.
	before, err := c.store.List(ctx, store.Filter{Kind: domain.KindMessage})
	require.NoError(t, err)
	sc := sessScope(sess)
	for _, p := range []rpc.ReportParams{
		{State: rpc.ReportProgress, Note: "x", Wait: true},
		{State: rpc.ReportWaiting, Note: "x", Wait: true},
		{State: rpc.ReportWaiting, Note: "x", Wait: true, Question: "q?", Context: "a\nb", Until: ptr(time.Now().Add(time.Hour))},
	} {
		_, err := c.report(ctx, sc, p)
		require.ErrorIs(t, err, domain.ErrInvalid, "%+v", p)
	}
	after, err := c.store.List(ctx, store.Filter{Kind: domain.KindMessage})
	require.NoError(t, err)
	assert.Len(t, after, len(before))
}

func ptr[T any](v T) *T { return &v }
