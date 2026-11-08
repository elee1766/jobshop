package control

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// TestWaitStaleSnapshot (JOB-144): a decision made from a wait snapshot
// is applied at the snapshot's revision. If the wait was cleared by hand
// and parked again on another condition while the pass was reading
// GitHub, the old decision fails with a conflict and the new wait stays.
func TestWaitStaleSnapshot(t *testing.T) {
	t.Parallel()
	c, _, prs, threads := waitController(t)
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	tk := mkTicket(t, c, store.NewTicket{Title: "ci", Pool: "dev"})
	prs.set(recPR, `[[]]`, `[{"name":"go","bucket":"pass"}]`)
	park := func(reason string) {
		_, err := c.waitAdd(ctx, admin, rpc.WaitAddParams{ID: tk.ID, Reason: reason, Conds: []store.WaitCond{{Type: store.WaitChecks, Repo: "acme/widget", PR: 5, State: "success"}}})
		require.NoError(t, err)
	}
	park("first")
	stale := mustGet(t, c, tk.ID)

	// Another actor clears the wait and parks a new one after the pass
	// listed the ticket.
	_, err := c.waitClear(ctx, admin, rpc.WaitClearParams{ID: tk.ID, Reason: "by hand"})
	require.NoError(t, err)
	park("second")

	// The pass read green checks for the stale snapshot: clearing the new
	// wait on that observation would be wrong.
	pass := c.newWaitPass(prs, threads)
	err = c.evalWait(ctx, pass, stale)
	require.ErrorIs(t, err, store.ErrConflict)
	w := waitOf(mustGet(t, c, tk.ID))
	require.NotNil(t, w, "the newer wait is untouched")
	assert.Equal(t, "second", w.Reason)

	// The next pass evaluates the current wait.
	c.pollWaits(ctx, prs, threads)
	assert.Nil(t, waitOf(mustGet(t, c, tk.ID)))
}
