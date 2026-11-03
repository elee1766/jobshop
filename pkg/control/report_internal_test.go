package control

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

func TestReportQuestionAnswer(t *testing.T) {
	c, clk, _ := stallController(t)
	ctx := t.Context()
	res, err := c.startRun(ctx, scope{admin: true, actor: "admin"}, rpc.RunStartParams{Workflow: "work", Vars: map[string]string{"n": "1"}, Key: "w1"}, nil)
	require.NoError(t, err)
	step := res.(rpc.RunResult).Steps[0]
	sess := mkTicket(t, c, store.NewTicket{Kind: store.KindSession, Pool: "dev", Fields: json.RawMessage(`{"agent":"fake"}`)})
	_, _, err = c.store.Claim(ctx, step.ID, sess.ID, time.Hour)
	require.NoError(t, err)
	sc := scope{actor: sess.ID, session: sess.ID, project: "js", pool: "dev"}
	admin := scope{admin: true, actor: "admin"}

	_, err = c.report(ctx, sc, rpc.ReportParams{State: "stuck"})
	require.ErrorIs(t, err, store.ErrInvalid)

	// blocked labels the ticket; any other state clears the label.
	_, err = c.report(ctx, sc, rpc.ReportParams{State: rpc.ReportBlocked, Note: "needs #12"})
	require.NoError(t, err)
	got, err := c.store.Get(ctx, step.ID)
	require.NoError(t, err)
	assert.Contains(t, got.Labels, blockedLabel)
	rm, ok := reportOf(got)
	require.True(t, ok)
	assert.Equal(t, "needs #12", rm.Note)
	assert.Contains(t, string(got.Meta), `"github"`, "other packs are kept")
	ls, err := c.list(ctx, admin, rpc.ListParams{Report: rpc.ReportBlocked})
	require.NoError(t, err)
	assert.Len(t, ls.([]store.Ticket), 1)

	// A waiting report with a question files a needs-human question.
	out, err := c.report(ctx, sc, rpc.ReportParams{State: rpc.ReportWaiting, Note: "base?", Question: "Which base branch?"})
	require.NoError(t, err)
	q := out.(rpc.ReportResult).Question
	require.NotNil(t, q)
	assert.Equal(t, store.KindMessage, q.Kind)
	assert.Contains(t, q.Labels, needsHumanLabel)
	qm, ok := questionOf(*q)
	require.True(t, ok)
	assert.Equal(t, questionMeta{Ticket: step.ID, Session: sess.ID}, qm)
	got, _ = c.store.Get(ctx, step.ID)
	assert.NotContains(t, got.Labels, blockedLabel)
	assert.Empty(t, runnerInbox(t, c, sess.ID), "the question is not delivered to any session")

	// Waiting on an open question is not a stalled step or silent session;
	// the question shows up as needs_human instead.
	clk.advance(3 * time.Hour)
	keys := stallKeys(t, c)
	assert.Contains(t, keys, "stall:needs_human:"+q.ID)
	assert.NotContains(t, keys, "stall:step_stalled:"+step.ID)
	assert.NotContains(t, keys, "stall:session_silent:"+sess.ID)

	// The open question is an attention item of kind question (#185).
	items, err := c.attentionItems(ctx)
	require.NoError(t, err)
	var qi *rpc.AttentionItem
	for i := range items {
		if items[i].ID == q.ID {
			qi = &items[i]
		}
	}
	require.NotNil(t, qi, "question listed in attention")
	assert.Equal(t, rpc.AttentionQuestion, qi.Kind)

	_, err = c.answerQuestion(ctx, sc, rpc.AnswerParams{ID: q.ID, Text: "main"})
	require.Error(t, err, "sessions cannot answer")
	ans, err := c.answerQuestion(ctx, admin, rpc.AnswerParams{ID: q.ID, Text: "use main", From: "@op"})
	require.NoError(t, err)
	ar := ans.(rpc.AnswerResult)
	assert.Equal(t, resolutionAnswered, ar.Question.Resolution)
	require.NotNil(t, ar.Message)
	mf, err := store.DecodeFields[store.MessageFields](*ar.Message)
	require.NoError(t, err)
	assert.Equal(t, store.MessageFields{To: sess.ID, From: "@op"}, mf)
	assert.Equal(t, "use main", ar.Message.Body)
	inbox := runnerInbox(t, c, sess.ID)
	require.Len(t, inbox, 1)
	assert.Equal(t, ar.Message.ID, inbox[0].ID)
	got, _ = c.store.Get(ctx, step.ID)
	rm, _ = reportOf(got)
	assert.Equal(t, rpc.ReportProgress, rm.State)
	cs, err := c.store.Comments(ctx, step.ID)
	require.NoError(t, err)
	assert.Contains(t, cs[len(cs)-1].Body, "use main")
	_, err = c.answerQuestion(ctx, admin, rpc.AnswerParams{ID: q.ID, Text: "again"})
	require.ErrorIs(t, err, store.ErrInvalid)

	// waiting --until hands the ticket back deferred.
	until := clk.now().Add(2 * time.Hour)
	_, err = c.report(ctx, sc, rpc.ReportParams{State: rpc.ReportReady, Until: &until})
	require.ErrorIs(t, err, store.ErrInvalid, "until is only for waiting")
	_, err = c.report(ctx, sc, rpc.ReportParams{State: rpc.ReportWaiting, Note: "CI", Until: &until})
	require.NoError(t, err)
	got, _ = c.store.Get(ctx, step.ID)
	assert.Equal(t, store.StatusOpen, got.Status)
	assert.True(t, got.DeferUntil.Equal(until))
}

// runnerInbox is the inbox a session runner for id would read.
func runnerInbox(t *testing.T, c *Controller, id string) []store.Ticket {
	t.Helper()
	r := &runner{c: c, id: id, pool: "dev"}
	return r.inbox(t.Context(), false)
}
