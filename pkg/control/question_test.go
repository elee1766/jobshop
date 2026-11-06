package control_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// TestQuestionOptions (JOB-101): a session asks with context and two
// options; attention returns both with the controller's links; answering
// with option 2 delivers the chosen option, its description and the note
// to the asking session.
func TestQuestionOptions(t *testing.T) {
	root, start := poolRoot(t, "1m")
	// A chat pool session claims nothing on its own, so the asked ticket
	// stays held while the question is open.
	f, err := os.OpenFile(filepath.Join(root.Dir, config.FileName), os.O_APPEND|os.O_WRONLY, 0o600)
	require.NoError(t, err)
	_, err = fmt.Fprintf(f, "  ask:\n    project: js\n    agent: fake\n    type: chat\n    max_sessions: 2\n    dir: %q\n", root.Dir)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	stop := start()
	defer stop()
	ctx := t.Context()
	admin := dial(t, root.AdminSocket())

	var task store.Ticket
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{
		Title: "Pick a base branch", Meta: json.RawMessage(`{"github":{"repo":"acme/widget","issue":7}}`),
	}, &task))
	require.NoError(t, admin.Call(ctx, rpc.MethodCommentAdd, rpc.CommentParams{ID: task.ID, Body: "tests pass on both branches"}, nil))
	var sess rpc.SessionNewResult
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionNew, rpc.SessionNewParams{Pool: "ask"}, &sess))
	w := dial(t, sess.Socket)
	require.NoError(t, w.Call(ctx, rpc.MethodClaim, rpc.ClaimParams{ID: task.ID}, nil))

	var rep rpc.ReportResult
	require.NoError(t, w.Call(ctx, rpc.MethodReport, rpc.ReportParams{
		State: rpc.ReportWaiting, Note: "base?", Question: "Which base branch?",
		Context: "I rebased the fix onto both branches.\nmain has the new API, release does not; the fix needs it there too.",
		Options: []rpc.QuestionOption{
			{Title: "main", Description: "Ship in the next minor.", Recommended: true},
			{Title: "release", Description: "Backport the API with the fix."},
		},
	}, &rep))
	require.NotNil(t, rep.Question)
	q := rep.Question

	var att rpc.AttentionResult
	require.NoError(t, admin.Call(ctx, rpc.MethodAttention, struct{}{}, &att))
	var it *rpc.AttentionItem
	for i := range att.Items {
		if att.Items[i].ID == q.ID {
			it = &att.Items[i]
		}
	}
	require.NotNil(t, it, "the question is listed")
	assert.Equal(t, rpc.AttentionQuestion, it.Kind)
	assert.Equal(t, []rpc.QuestionOption{
		{ID: 1, Title: "main", Description: "Ship in the next minor.", Recommended: true},
		{ID: 2, Title: "release", Description: "Backport the API with the fix."},
	}, it.Options)
	assert.True(t, strings.HasPrefix(it.Context, "I rebased the fix onto both branches."), it.Context)
	for _, want := range []string{
		"- Ticket: " + task.ID + " (Pick a base branch)",
		"- Issue: https://github.com/acme/widget/issues/7",
		"- Asked by session: " + sess.Ticket.ID,
		"tests pass on both branches",
	} {
		assert.Contains(t, it.Context, want)
	}

	// An option the question does not have is refused; option 2 answers.
	err = admin.Call(ctx, rpc.MethodQuestionAnswer, rpc.AnswerParams{ID: q.ID, Option: 3}, nil)
	require.ErrorIs(t, err, store.ErrInvalid)
	var ans rpc.AnswerResult
	require.NoError(t, admin.Call(ctx, rpc.MethodQuestionAnswer, rpc.AnswerParams{ID: q.ID, Option: 2, Text: "after the freeze", From: "@op"}, &ans))
	require.NotNil(t, ans.Message)
	want := "Chose option 2: release\nBackport the API with the fix.\n\nNote: after the freeze"
	assert.Equal(t, want, ans.Message.Body)
	assert.Equal(t, "answered", ans.Question.Resolution)
	passEventually(t, root.Dir, func() bool {
		return strings.Contains(promptsFile(t, root), "Chose option 2: release")
	}, "the asking session gets the chosen option")
}
