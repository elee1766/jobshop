package control_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/clock"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

func getTicket(t *testing.T, admin *rpc.Client, id string) rpc.ShowResult {
	t.Helper()
	var r rpc.ShowResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: id}, &r))
	return r
}

func failures(t *testing.T, tk store.Ticket) int {
	t.Helper()
	f, err := store.DecodeFields[store.StepFields](tk)
	require.NoError(t, err)
	return f.Failures
}

// escalations lists escalation tickets for run.
func escalations(t *testing.T, admin *rpc.Client, run string) []store.Ticket {
	t.Helper()
	var all []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Label: "needs-human"}, &all))
	var out []store.Ticket
	for _, x := range all {
		if x.Title == "Escalate run "+run {
			out = append(out, x)
		}
	}
	return out
}

// claimAs claims ticket id for holder through the admin socket.
func claimAs(t *testing.T, admin *rpc.Client, id, holder string) rpc.ClaimResult {
	t.Helper()
	var cl rpc.ClaimResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodClaim, rpc.ClaimParams{ID: id, Holder: holder}, &cl))
	return cl
}

// TestStepFailureSources counts one failure for each source: close
// failed, CLI release, dead-session release, and lease expiry. Below the
// limit the step reopens with a comment naming the attempt.
func TestStepFailureSources(t *testing.T) {
	// A fake clock: step 3 advances it past the lease, so the lease never
	// expires mid-step under load (#314, #375).
	clk := clock.NewFake(time.Date(2026, 11, 1, 5, 0, 0, 0, time.UTC))
	admin := runRootLease(t, "2s", clk)
	// implement has retries: 3, so three failures reopen then exhaust.
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "5"}})
	step := r.Steps[0].ID

	// 1. Close failed.
	closeStep(t, admin, step, "failed")
	got := getTicket(t, admin, step)
	assert.Equal(t, store.StatusOpen, got.Ticket.Status, "below the limit a failed close reopens")
	assert.Equal(t, 1, failures(t, got.Ticket))
	require.NotEmpty(t, got.Comments)
	assert.Contains(t, got.Comments[len(got.Comments)-1].Body, "Attempt 1 of 3 failed (closed failed)")

	// 2. CLI release while holding.
	var sess rpc.SessionNewResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionNew, rpc.SessionNewParams{Agent: "jcode", Pool: "dev"}, &sess))
	s := dial(t, sess.Socket)
	var cl rpc.ClaimResult
	require.NoError(t, s.Call(t.Context(), rpc.MethodClaim, rpc.ClaimParams{}, &cl))
	require.Equal(t, step, cl.Ticket.ID)
	require.NoError(t, s.Call(t.Context(), rpc.MethodRelease, rpc.ReleaseParams{Reason: "stuck"}, nil))
	got = getTicket(t, admin, step)
	assert.Equal(t, 2, failures(t, got.Ticket))
	assert.Contains(t, got.Comments[len(got.Comments)-1].Body, "Attempt 2 of 3 failed (released: stuck)")

	// 3. Lease expiry, swept on the tick, exhausts the step.
	claimAs(t, admin, step, "js-gone")
	clk.Advance(3 * time.Second)
	passUntil(t, admin, func() bool {
		return getTicket(t, admin, step).Ticket.Status == store.StatusClosed
	}, "lease expiry closes the step")
	got = getTicket(t, admin, step)
	assert.Equal(t, "failed", got.Ticket.Resolution)
	assert.Equal(t, 3, failures(t, got.Ticket))
	assert.Contains(t, got.Comments[len(got.Comments)-1].Body, "lease_expired")
	assert.Contains(t, got.Comments[len(got.Comments)-1].Body, "Retry limit 3 reached")
	passUntil(t, admin, func() bool {
		return runState(t, showRun(t, admin, r.Run.ID)) == store.RunNeedsHuman
	}, "run needs a human")
}

// TestStepFailureDeadSession counts a release of a dead session's claim,
// and does not count a pre-assignment the session never claimed.
func TestStepFailureDeadSession(t *testing.T) {
	admin := runRoot(t)
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "5"}})
	var sess rpc.SessionNewResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionNew, rpc.SessionNewParams{Agent: "jcode", Pool: "dev"}, &sess))
	claimAs(t, admin, r.Steps[0].ID, sess.Ticket.ID)
	// Pre-assign step 2 to the same session: never claimed, so no failure.
	holder := sess.Ticket.ID
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketUpdate, rpc.UpdateParams{ID: r.Steps[1].ID, Holder: &holder}, nil))

	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketClose, rpc.CloseParams{ID: sess.Ticket.ID}, nil))
	got := getTicket(t, admin, r.Steps[0].ID)
	assert.Equal(t, store.StatusOpen, got.Ticket.Status)
	assert.Equal(t, 1, failures(t, got.Ticket))
	assert.Contains(t, got.Comments[len(got.Comments)-1].Body, "session closed")
	assert.Equal(t, 0, failures(t, getTicket(t, admin, r.Steps[1].ID).Ticket))
}

// TestStepEscalation closes the step failed at its limit, sets the run to
// needs_human and files exactly one escalation ticket in escalate_pool.
func TestStepEscalation(t *testing.T) {
	admin := runRoot(t)
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "77"}})
	closeStep(t, admin, r.Steps[0].ID, "done")
	review := r.Steps[1].ID
	require.NoError(t, admin.Call(t.Context(), rpc.MethodCommentAdd, rpc.CommentParams{ID: review, Body: "CI red on lint"}, nil))
	closeStep(t, admin, review, "failed")
	assert.Empty(t, escalations(t, admin, r.Run.ID), "below the limit no escalation")
	closeStep(t, admin, review, "failed")

	got := getTicket(t, admin, review)
	assert.Equal(t, store.StatusClosed, got.Ticket.Status)
	assert.Equal(t, "failed", got.Ticket.Resolution)
	run := showRun(t, admin, r.Run.ID)
	assert.Equal(t, store.RunNeedsHuman, runState(t, run))
	assert.Equal(t, store.StatusOpen, run.Run.Status, "the run stays open")
	assert.Empty(t, readyIDs(t, admin), "deploy stays blocked")

	esc := escalations(t, admin, r.Run.ID)
	require.Len(t, esc, 1)
	e := esc[0]
	assert.Equal(t, "dev", e.Pool, "escalate_pool")
	assert.Equal(t, store.KindTask, e.Kind)
	assert.JSONEq(t, `{"github":{"repo":"acme/widget","issue":77},"run":{"id":"`+r.Run.ID+`","workflow":"pr","escalation":"`+r.Run.ID+`/1"}}`, string(e.Meta))
	assert.Contains(t, e.Body, "Step 1 review")
	assert.Contains(t, e.Body, "failed, failures 2/2")
	assert.Contains(t, e.Body, "CI red on lint")
	assert.Contains(t, e.Body, "Retry limit 2 reached")
	assert.Equal(t, 3, strings.Count(e.Body[strings.Index(e.Body, "Step 1 review"):strings.Index(e.Body, "Step 2")], "\n- ["),
		"last 3 comments of the step")

	// Ticks repair but never duplicate.
	pass(t, admin, 3)
	assert.Len(t, escalations(t, admin, r.Run.ID), 1)
	// Closing the escalation does not file another for the same episode.
	closeStep(t, admin, e.ID, "done")
	pass(t, admin, 2)
	assert.Len(t, escalations(t, admin, r.Run.ID), 1)

	// A resumed run that exhausts again is a new episode: one more ticket.
	require.NoError(t, admin.Call(t.Context(), rpc.MethodRunResume, rpc.RunResumeParams{ID: r.Run.ID}, nil))
	closeStep(t, admin, review, "failed")
	closeStep(t, admin, review, "failed")
	assert.Len(t, escalations(t, admin, r.Run.ID), 2)
}
