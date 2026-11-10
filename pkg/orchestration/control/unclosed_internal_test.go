package control

import (
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// TestUnclosedTurnFirstReleaseNoFailure (#241): the first turn that ends
// holding a step releases it deferred without counting a failure; a
// second unclosed release of the same step counts one.
func TestUnclosedTurnFirstReleaseNoFailure(t *testing.T) {
	t.Parallel()
	c, clk, _ := subController(t)
	ctx := t.Context()
	sess := mkSession(t, c)
	step := mkTicket(t, c, store.NewTicket{Kind: domain.KindStep, Title: "implement", Pool: "dev", Fields: json.RawMessage(`{"run":"x","step":"implement","retries":3}`)})
	r := &runner{c: c, id: sess.ID, log: slog.New(slog.DiscardHandler)}
	failures := func() int {
		got, err := c.store.Get(ctx, step.ID)
		require.NoError(t, err)
		f, err := domain.DecodeFields[domain.StepFields](got)
		require.NoError(t, err)
		return f.Failures
	}
	turn := func() domain.Ticket {
		_, err := c.store.Update(ctx, "controller", step.ID, store.Update{DeferUntil: &time.Time{}})
		require.NoError(t, err)
		_, _, err = c.store.Claim(ctx, step.ID, sess.ID, time.Hour)
		require.NoError(t, err)
		// ticketID "" skips the reminder turn (no agent connection here).
		r.afterTurn(ctx, nil, "")
		got, err := c.store.Get(ctx, step.ID)
		require.NoError(t, err)
		return got
	}

	got := turn()
	assert.Equal(t, domain.StatusOpen, got.Status)
	assert.True(t, got.DeferUntil.After(clk.now()), "deferred")
	assert.Equal(t, 0, failures(), "first unclosed turn is not a failure")

	got = turn()
	assert.Equal(t, domain.StatusOpen, got.Status)
	assert.True(t, got.DeferUntil.After(clk.now()), "deferred")
	assert.Equal(t, 1, failures(), "a repeat counts")
}
