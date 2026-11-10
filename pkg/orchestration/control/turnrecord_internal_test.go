package control

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/orchestration/triggers"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// TestTurnRowsAreNotActivity (JOB-184): turn rows leave the ticket's revision
// and its last activity alone, so a held ticket whose session only turns
// still ages for session_silent and the stale signals.
func TestTurnRowsAreNotActivity(t *testing.T) {
	t.Parallel()
	c, clk, _, _ := waitController(t)
	ctx := t.Context()
	sess := mkSession(t, c)
	task := mkTicket(t, c, store.NewTicket{Title: "work", Pool: "dev"})
	claimed, _, err := c.store.Claim(ctx, task.ID, sess.ID, time.Hour)
	require.NoError(t, err)
	before, err := triggers.LastActivity(ctx, c.store, claimed)
	require.NoError(t, err)

	clk.advance(10 * time.Minute)
	require.NoError(t, c.store.Record(ctx, sess.ID, task.ID, domain.OpTurnStart, struct{}{}))
	clk.advance(time.Minute)
	require.NoError(t, c.store.Record(ctx, sess.ID, task.ID, domain.OpTurnEnd, domain.TurnDiff{Stop: "end_turn"}))

	after, err := triggers.LastActivity(ctx, c.store, claimed)
	require.NoError(t, err)
	assert.Equal(t, before, after, "a turn is not activity on the ticket")
	got, err := c.store.Get(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, claimed.Revision, got.Revision, "a turn row does not bump the revision")
	hist, err := store.History(ctx, c.store, task.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.OpTurnEnd, hist[len(hist)-1].Op)
}
