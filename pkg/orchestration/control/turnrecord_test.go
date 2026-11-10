package control_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// TestTurnRows (JOB-184): a session that claims a ticket and runs two turns
// (the prompt, then the unclosed reminder) leaves a turn_start and turn_end
// pair for each on the ticket. The rows do not move the ticket's revision,
// and ticket.timeline lists the two turns on the working span.
func TestTurnRows(t *testing.T) {
	root, start := poolRoot(t, "1m")
	// The first turn blocks until "go", so the fake clock can move inside the
	// claim: a claim with no length is not drawn.
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "noclose"), nil, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "block"), nil, 0o600))
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())

	var task domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "work", Pool: "dev"}, &task))
	passEventually(t, root.Dir, func() bool {
		_, err := os.Stat(filepath.Join(root.Dir, "started"))
		return err == nil
	}, "first turn started")
	advance(t, root.Dir, time.Minute)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))
	var got rpc.ShowResult
	ends := func() int {
		n := 0
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: task.ID}, &got))
		for _, h := range got.History {
			if h.Op == domain.OpTurnEnd {
				n++
			}
		}
		return n
	}
	passEventually(t, root.Dir, func() bool { return ends() >= 2 }, "two turns recorded")

	var ops []string
	var claimRev int64
	for _, h := range got.History {
		switch h.Op {
		case "claim":
			claimRev = h.Revision
		case domain.OpTurnStart, domain.OpTurnEnd:
			ops = append(ops, h.Op)
			assert.Equal(t, claimRev, h.Revision, "a turn row does not bump the ticket's revision")
			assert.NotEmpty(t, h.Actor)
			if h.Op == domain.OpTurnEnd {
				assert.JSONEq(t, `{"stop":"end_turn"}`, string(h.Diff))
			}
		}
	}
	assert.Equal(t, []string{domain.OpTurnStart, domain.OpTurnEnd, domain.OpTurnStart, domain.OpTurnEnd}, ops)

	var tl rpc.TimelineResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketTimeline, rpc.GraphParams{ID: task.ID}, &tl))
	require.NotEmpty(t, tl.Rows)
	var turns []rpc.TimelineTurn
	for _, s := range tl.Rows[0].Segments {
		if s.State == rpc.TimelineWorking || s.State == rpc.TimelineRetry {
			turns = append(turns, s.Turns...)
		}
	}
	require.Len(t, turns, 2)
	for _, turn := range turns {
		assert.Equal(t, "end_turn", turn.Stop)
		assert.False(t, turn.End.IsZero())
		assert.False(t, turn.End.Before(turn.Start))
	}
	assert.False(t, turns[1].Start.Before(turns[0].End), "turns are in order")
}
