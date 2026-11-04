package control_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// stepByName maps workflow step ids to their tickets in r.
func stepByName(t *testing.T, r rpc.RunResult) map[string]store.Ticket {
	t.Helper()
	out := map[string]store.Ticket{}
	for _, s := range r.Steps {
		f, err := store.DecodeFields[store.StepFields](s)
		require.NoError(t, err)
		out[f.Step] = s
	}
	return out
}

func closeOutcome(t *testing.T, admin *rpc.Client, id, outcome string) {
	t.Helper()
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketClose, rpc.CloseParams{ID: id, Outcome: outcome}, nil))
}

// TestRunGraphApprove: review and device become ready together after
// build (needs:), the approve outcome opens rollout and accept, and the
// run closes done.
func TestRunGraphApprove(t *testing.T) {
	admin := runRoot(t)
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "graph", Vars: map[string]string{"issue": "7"}})
	s := stepByName(t, r)
	assert.Equal(t, []string{s["build"].ID}, readyIDs(t, admin), "only build is ready")

	closeStep(t, admin, s["build"].ID, "done")
	assert.ElementsMatch(t, []string{s["review"].ID, s["device"].ID}, readyIDs(t, admin), "parallel steps ready together")

	closeOutcome(t, admin, s["review"].ID, "approve")
	assert.Equal(t, []string{s["device"].ID}, readyIDs(t, admin), "rollout still waits on device")
	closeStep(t, admin, s["device"].ID, "done")
	assert.Equal(t, []string{s["rollout"].ID}, readyIDs(t, admin), "approve opens rollout")

	closeStep(t, admin, s["rollout"].ID, "done")
	assert.Equal(t, []string{s["accept"].ID}, readyIDs(t, admin))
	closeStep(t, admin, s["accept"].ID, "done")

	got := showRun(t, admin, r.Run.ID)
	assert.Equal(t, store.RunDone, runState(t, got))
	rf, err := store.DecodeFields[store.StepFields](stepByName(t, got)["review"])
	require.NoError(t, err)
	assert.Equal(t, "approve", rf.Outcome)
}

// TestRunGraphReject: a reject outcome skips both gated steps (the second
// only after the first is skipped) and the run closes done.
func TestRunGraphReject(t *testing.T) {
	admin := runRoot(t)
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "graph", Vars: map[string]string{"issue": "8"}})
	s := stepByName(t, r)
	closeStep(t, admin, s["build"].ID, "done")
	closeStep(t, admin, s["device"].ID, "done")
	closeOutcome(t, admin, s["review"].ID, "reject")

	got := showRun(t, admin, r.Run.ID)
	gs := stepByName(t, got)
	for _, name := range []string{"rollout", "accept"} {
		assert.Equal(t, store.StatusClosed, gs[name].Status, name)
		assert.Equal(t, store.StepSkipped, gs[name].Resolution, name)
	}
	assert.Empty(t, readyIDs(t, admin))
	assert.Equal(t, store.RunDone, runState(t, got), "skipped steps finish the run")
}

// TestRunGraphOutcomeErrors: --outcome is for steps and must be a word.
func TestRunGraphOutcomeErrors(t *testing.T) {
	admin := runRoot(t)
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "graph", Vars: map[string]string{"issue": "9"}})
	err := admin.Call(t.Context(), rpc.MethodTicketClose, rpc.CloseParams{ID: r.Steps[0].ID, Outcome: "Not A Word"}, nil)
	require.ErrorIs(t, err, store.ErrInvalid)
	err = admin.Call(t.Context(), rpc.MethodTicketClose, rpc.CloseParams{ID: r.Run.ID, Outcome: "ok"}, nil)
	require.ErrorIs(t, err, store.ErrInvalid)
	assert.Equal(t, store.StatusOpen, showRun(t, admin, r.Run.ID).Steps[0].Status, "failed close leaves the step open")
}
