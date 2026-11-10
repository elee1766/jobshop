package control_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// TestHandoff checks that a session hands off its own ticket in one
// write (#492): pool, release, fence and note land in a single revision,
// so no history entry shows the ticket held in the new pool.
func TestHandoff(t *testing.T) {
	admin := runRoot(t)
	ctx := t.Context()
	var a, b domain.Ticket
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "a", Pool: "dev"}, &a))
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "b", Pool: "dev"}, &b))

	var sess rpc.SessionNewResult
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionNew, rpc.SessionNewParams{Agent: "jcode", Pool: "dev"}, &sess))
	s := dial(t, sess.Socket)
	var cl rpc.ClaimResult
	require.NoError(t, s.Call(ctx, rpc.MethodClaim, rpc.ClaimParams{ID: a.ID}, &cl))
	other := claimAs(t, admin, b.ID, "js-other")

	// Another session's ticket is forbidden and left untouched.
	err := s.Call(ctx, rpc.MethodHandoff, rpc.HandoffParams{ID: b.ID, Pool: "refinery"}, nil)
	require.ErrorIs(t, err, rpc.ErrForbidden)
	got := getTicket(t, admin, b.ID)
	assert.Equal(t, "dev", got.Ticket.Pool)
	assert.Equal(t, "js-other", got.Ticket.Holder)

	err = s.Call(ctx, rpc.MethodHandoff, rpc.HandoffParams{Pool: "nope"}, nil)
	require.Error(t, err, "unknown pool")

	before := getTicket(t, admin, a.ID).Ticket
	require.NoError(t, s.Call(ctx, rpc.MethodHandoff, rpc.HandoffParams{Pool: "refinery", Note: "needs review"}, nil))
	got = getTicket(t, admin, a.ID)
	assert.Equal(t, "refinery", got.Ticket.Pool)
	assert.Equal(t, domain.StatusOpen, got.Ticket.Status)
	assert.Empty(t, got.Ticket.Holder)
	assert.Greater(t, got.Ticket.Fence, cl.Claim.Fence)
	assert.Equal(t, before.Revision+1, got.Ticket.Revision, "one write")
	require.NotEmpty(t, got.Comments)
	assert.Equal(t, "needs review", got.Comments[len(got.Comments)-1].Body)
	for _, h := range got.History {
		assert.NotEqual(t, "update", h.Op, "handoff wrote a separate update")
	}

	// The old claim is fenced out.
	require.Error(t, s.Call(ctx, rpc.MethodTicketClose, rpc.CloseParams{ID: a.ID}, nil))

	// Admin may hand off any ticket.
	require.NoError(t, admin.Call(ctx, rpc.MethodHandoff, rpc.HandoffParams{ID: b.ID, Pool: "refinery"}, nil))
	got = getTicket(t, admin, b.ID)
	assert.Equal(t, "refinery", got.Ticket.Pool)
	assert.Empty(t, got.Ticket.Holder)
	assert.Greater(t, got.Ticket.Fence, other.Claim.Fence)
}
