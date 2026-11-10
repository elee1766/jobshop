package control_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

func commentBodies(t *testing.T, admin *rpc.Client, id string) []string {
	t.Helper()
	var out []string
	for _, c := range getTicket(t, admin, id).Comments {
		out = append(out, c.Body)
	}
	return out
}

func failReview(t *testing.T, admin *rpc.Client, review string) {
	t.Helper()
	closeStep(t, admin, review, "failed")
	closeStep(t, admin, review, "failed")
}

// TestDerivedEscalationsClose escalates a run twice and finishes it: the
// first escalation closes superseded when the run is resumed, the second
// when the run finishes (JOB-195).
func TestDerivedEscalationsClose(t *testing.T) {
	admin := runRoot(t)
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "78"}})
	closeStep(t, admin, r.Steps[0].ID, "done")
	review := r.Steps[1].ID
	failReview(t, admin, review)
	esc := escalations(t, admin, r.Run.ID)
	require.Len(t, esc, 1)
	assert.Equal(t, domain.StatusOpen, esc[0].Status, "open while the run needs a human")
	pass(t, admin, 2)
	assert.Equal(t, domain.StatusOpen, getTicket(t, admin, esc[0].ID).Ticket.Status, "a sweep leaves the current round")

	require.NoError(t, admin.Call(t.Context(), rpc.MethodRunResume, rpc.RunResumeParams{ID: r.Run.ID}, nil))
	first := getTicket(t, admin, esc[0].ID)
	assert.Equal(t, domain.StatusClosed, first.Ticket.Status)
	assert.Equal(t, "superseded", first.Ticket.Resolution)
	assert.Contains(t, commentBodies(t, admin, esc[0].ID), "run "+r.Run.ID+" resumed")

	failReview(t, admin, review)
	all := escalations(t, admin, r.Run.ID)
	require.Len(t, all, 2)
	var second domain.Ticket
	for _, e := range all {
		if e.ID != esc[0].ID {
			second = e
		}
	}
	assert.Equal(t, domain.StatusOpen, second.Status)

	require.NoError(t, admin.Call(t.Context(), rpc.MethodRunResume, rpc.RunResumeParams{ID: r.Run.ID}, nil))
	closeStep(t, admin, review, "done")
	closeStep(t, admin, r.Steps[2].ID, "done")
	pass(t, admin, 2)
	assert.Equal(t, "done", showRun(t, admin, r.Run.ID).Run.Resolution)
	for _, e := range []string{esc[0].ID, second.ID} {
		got := getTicket(t, admin, e)
		assert.Equal(t, domain.StatusClosed, got.Ticket.Status, e)
		assert.NotEmpty(t, commentBodies(t, admin, e), e)
	}
	assert.Equal(t, "superseded", getTicket(t, admin, esc[0].ID).Ticket.Resolution)
}

// TestDerivedEscalationHeld leaves an escalation a session holds open
// while the run lives, and closes it with a comment when the run is
// cancelled.
func TestDerivedEscalationHeld(t *testing.T) {
	admin := runRoot(t)
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "79"}})
	closeStep(t, admin, r.Steps[0].ID, "done")
	review := r.Steps[1].ID
	failReview(t, admin, review)
	esc := escalations(t, admin, r.Run.ID)
	require.Len(t, esc, 1)
	claimAs(t, admin, esc[0].ID, "human-1")

	require.NoError(t, admin.Call(t.Context(), rpc.MethodRunResume, rpc.RunResumeParams{ID: r.Run.ID}, nil))
	pass(t, admin, 2)
	got := getTicket(t, admin, esc[0].ID)
	assert.Equal(t, domain.StatusInProgress, got.Ticket.Status, "held: not closed under its holder")
	assert.Equal(t, "human-1", got.Ticket.Holder)
	assert.Contains(t, commentBodies(t, admin, esc[0].ID), "run "+r.Run.ID+" resumed. Left open because human-1 holds it; close it when done.")
	assert.Len(t, commentBodies(t, admin, esc[0].ID), 1, "one comment, not one per sweep")

	// A new round supersedes the held one: still open, with a comment naming it.
	failReview(t, admin, review)
	var newest string
	for _, e := range escalations(t, admin, r.Run.ID) {
		if e.ID != esc[0].ID {
			newest = e.ID
		}
	}
	require.NotEmpty(t, newest)
	assert.Equal(t, domain.StatusInProgress, getTicket(t, admin, esc[0].ID).Ticket.Status)
	assert.Contains(t, commentBodies(t, admin, esc[0].ID), "superseded by "+newest+". Left open because human-1 holds it; close it when done.")

	require.NoError(t, admin.Call(t.Context(), rpc.MethodRunCancel, rpc.IDParams{ID: r.Run.ID}, nil))
	got = getTicket(t, admin, esc[0].ID)
	assert.Equal(t, domain.StatusClosed, got.Ticket.Status)
	assert.Equal(t, domain.RunCancelled, got.Ticket.Resolution)
	assert.Contains(t, commentBodies(t, admin, esc[0].ID), "run "+r.Run.ID+" cancelled (closed while human-1 held it)")
}

// TestDerivedQuestionWithdrawn closes a question when the ticket it asks
// about closes without an answer.
func TestDerivedQuestionWithdrawn(t *testing.T) {
	admin := runRoot(t)
	var task domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Project: "js", Title: "work"}, &task))
	var rep rpc.ReportResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodReport, rpc.ReportParams{ID: task.ID, State: rpc.ReportWaiting, Note: "waiting", Question: "which key?"}, &rep))
	require.NotNil(t, rep.Question)
	q := rep.Question.ID
	pass(t, admin, 2)
	assert.Equal(t, domain.StatusOpen, getTicket(t, admin, q).Ticket.Status, "open while its ticket is")

	closeStep(t, admin, task.ID, "done")
	pass(t, admin, 2)
	got := getTicket(t, admin, q)
	assert.Equal(t, domain.StatusClosed, got.Ticket.Status)
	assert.Equal(t, "withdrawn", got.Ticket.Resolution)
	assert.Contains(t, commentBodies(t, admin, q), "asked ticket "+task.ID+" closed (done)")
}
