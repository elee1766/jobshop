package control_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

func startBatch(t *testing.T, admin *rpc.Client, p rpc.BatchStartParams) rpc.BatchResult {
	t.Helper()
	var r rpc.BatchResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodBatchStart, p, &r))
	return r
}

func claimID(t *testing.T, admin *rpc.Client, id string) error {
	t.Helper()
	return admin.Call(t.Context(), rpc.MethodClaim, rpc.ClaimParams{ID: id, Holder: "tester"}, nil)
}

// TestBatchStartOrder: a>b>c edges exist from creation, so b's implement
// cannot be claimed until a's last step closes.
func TestBatchStartOrder(t *testing.T) {
	admin := runRoot(t)
	r := startBatch(t, admin, rpc.BatchStartParams{Key: "track", Workflow: "pr", Issues: []string{"1", "2", "3"}, Order: []string{"1>2>3"}})
	require.Len(t, r.Runs, 3)
	assert.Equal(t, store.KindBatch, r.Batch.Kind)
	assert.Empty(t, r.Batch.Pool)
	f, err := store.DecodeFields[store.BatchFields](r.Batch)
	require.NoError(t, err)
	assert.Equal(t, []string{"1", "2", "3"}, f.Issues)
	assert.Equal(t, [][2]string{{"1", "2"}, {"2", "3"}}, f.Order)
	assert.Equal(t, store.BatchOpen, f.State)
	for i, run := range r.Runs {
		assert.Equal(t, r.Batch.ID, run.Run.ParentID)
		rf, err := store.DecodeFields[store.RunFields](run.Run)
		require.NoError(t, err)
		assert.Equal(t, "issue-"+f.Issues[i], rf.Key)
	}
	a, b, c := r.Runs[0], r.Runs[1], r.Runs[2]
	assert.Equal(t, []string{a.Steps[0].ID}, readyIDs(t, admin))

	require.ErrorContains(t, claimID(t, admin, b.Steps[0].ID), "blocked")
	require.NoError(t, claimID(t, admin, a.Steps[0].ID))
	closeStep(t, admin, a.Steps[0].ID, "done")
	closeStep(t, admin, a.Steps[1].ID, "done")
	require.ErrorContains(t, claimID(t, admin, b.Steps[0].ID), "blocked", "review merged is not enough: wait for the last step")
	closeStep(t, admin, a.Steps[2].ID, "done")
	require.NoError(t, claimID(t, admin, b.Steps[0].ID))
	require.ErrorContains(t, claimID(t, admin, c.Steps[0].ID), "blocked")
}

// TestBatchStartSameKey returns the open batch without new runs.
func TestBatchStartSameKey(t *testing.T) {
	admin := runRoot(t)
	r1 := startBatch(t, admin, rpc.BatchStartParams{Key: "k", Workflow: "pr", Issues: []string{"1", "2"}})
	r2 := startBatch(t, admin, rpc.BatchStartParams{Key: "k", Workflow: "pr", Issues: []string{"1", "2", "9"}})
	assert.True(t, r2.Existing)
	assert.Equal(t, r1.Batch.ID, r2.Batch.ID)
	require.Len(t, r2.Runs, 2)
	assert.Equal(t, r1.Runs[0].Run.ID, r2.Runs[0].Run.ID)
}

// TestBatchStartAdopts re-parents an open issue-<n> run instead of
// starting another, and wires the order edge to it.
func TestBatchStartAdopts(t *testing.T) {
	admin := runRoot(t)
	old := startRun(t, admin, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "2"}, Key: "issue-2"})
	r := startBatch(t, admin, rpc.BatchStartParams{Key: "k", Workflow: "pr", Issues: []string{"1", "2"}, Order: []string{"1>2"}})
	require.Len(t, r.Runs, 2)
	assert.Equal(t, old.Run.ID, r.Runs[1].Run.ID)
	assert.Equal(t, []string{old.Run.ID}, r.Adopted)
	assert.Equal(t, r.Batch.ID, r.Runs[1].Run.ParentID)
	assert.Equal(t, []string{r.Runs[0].Steps[0].ID}, readyIDs(t, admin), "adopted run's first step now waits")
}

// TestBatchStartAdoptClaimed refuses to order an adopted run whose first
// step is already claimed, and writes nothing.
func TestBatchStartAdoptClaimed(t *testing.T) {
	admin := runRoot(t)
	old := startRun(t, admin, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "2"}, Key: "issue-2"})
	require.NoError(t, claimID(t, admin, old.Steps[0].ID))
	err := admin.Call(t.Context(), rpc.MethodBatchStart, rpc.BatchStartParams{Key: "k", Workflow: "pr", Issues: []string{"1", "2"}, Order: []string{"1>2"}}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), old.Steps[0].ID)
	var runs []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodRunList, rpc.RunListParams{}, &runs))
	assert.Len(t, runs, 1, "no run for issue 1 was created")
	assert.Empty(t, showRun(t, admin, old.Run.ID).Run.ParentID)

	// Without an order on it, the claimed run is adopted.
	r := startBatch(t, admin, rpc.BatchStartParams{Key: "k", Workflow: "pr", Issues: []string{"1", "2"}, Order: []string{"2>1"}})
	assert.Equal(t, []string{old.Run.ID}, r.Adopted)
}

func TestBatchStartInvalid(t *testing.T) {
	admin := runRoot(t)
	for _, p := range []rpc.BatchStartParams{
		{Workflow: "pr", Issues: []string{"1"}},
		{Key: "k", Workflow: "pr"},
		{Key: "k", Workflow: "pr", Issues: []string{"1", "1"}},
		{Key: "k", Workflow: "pr", Issues: []string{"1", "2"}, Order: []string{"1>3"}},
		{Key: "k", Workflow: "pr", Issues: []string{"1", "2"}, Order: []string{"1"}},
		{Key: "k", Workflow: "pr", Issues: []string{"1", "2"}, Order: []string{"1>2", "2>1"}},
		{Key: "k", Workflow: "nope", Issues: []string{"1"}},
	} {
		assert.Error(t, admin.Call(t.Context(), rpc.MethodBatchStart, p, nil), "%+v", p)
	}
	var runs []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodRunList, rpc.RunListParams{}, &runs))
	assert.Empty(t, runs)
}
