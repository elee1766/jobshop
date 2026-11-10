package control_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// writeSingleUseConfig is writePoolConfig with max_tickets set.
func writeSingleUseConfig(t *testing.T, root config.Root, maxSessions, maxTickets int) {
	t.Helper()
	cfg := fmt.Sprintf("projects:\n  js: {}\nagents:\n  fake:\n    command: %q\n    args: [fake-agent]\npools:\n  dev:\n    project: js\n    agent: fake\n    max_sessions: %d\n    max_tickets: %d\n    idle_timeout: 1m\n    dir: %q\n",
		os.Args[0], maxSessions, maxTickets, root.Dir)
	require.NoError(t, os.WriteFile(root.Config(), []byte(cfg), 0o600))
}

func sessionFields(t *testing.T, tk domain.Ticket) domain.SessionFields {
	t.Helper()
	f, err := domain.DecodeFields[domain.SessionFields](tk)
	require.NoError(t, err)
	return f
}

// TestMaxTicketsEndsSession (JOB-189): with max_tickets 1 a session works
// one ticket and ends ticket_limit; the next ticket gets a new session.
func TestMaxTicketsEndsSession(t *testing.T) {
	root, start := poolRoot(t, "1m")
	writeSingleUseConfig(t, root, 1, 1)
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())

	var a, b domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "a", Pool: "dev"}, &a))
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "b", Pool: "dev"}, &b))
	waitClosed(t, root, admin, a.ID)
	waitClosed(t, root, admin, b.ID)

	var sessions []domain.Ticket
	passEventually(t, root.Dir, func() bool {
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList,
			rpc.ListParams{Kind: domain.KindSession, Status: domain.StatusClosed}, &sessions))
		return len(sessions) >= 2
	}, "both single-use sessions end")
	for _, s := range sessions {
		assert.Equal(t, "ticket_limit", s.Resolution, "session %s", s.ID)
		assert.Equal(t, 1, sessionFields(t, s).Tickets)
	}
}

// TestMaxTicketsParkedTakesNoSlot (JOB-189): a session whose ticket waits
// with resume is suspended and does not count toward max_sessions, so the
// pool starts another session. The parked one resumes when a slot is free
// and the wait ended, then ends ticket_limit.
func TestMaxTicketsParkedTakesNoSlot(t *testing.T) {
	root, start := poolRoot(t, "1m")
	writeSingleUseConfig(t, root, 1, 1)
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	show := func(id string) rpc.ShowResult {
		var got rpc.ShowResult
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: id}, &got))
		return got
	}

	park, err := json.Marshal(rpc.WaitAddParams{
		Reason: "CI on PR #5", Resume: true,
		Conds: []domain.WaitCond{{Type: domain.WaitChecks, Repo: "acme/widget", PR: 5, State: "success"}},
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "park"), park, 0o600))
	var x domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "x", Pool: "dev"}, &x))
	var a string
	passEventually(t, root.Dir, func() bool {
		for _, s := range openSessions(t, admin) {
			if f := sessionFields(t, s); f.Suspended && f.Parked == x.ID {
				a = s.ID
				return true
			}
		}
		return false
	}, "A parks X and is suspended")

	// A is suspended and takes no slot: a new ticket gets its own session.
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "block"), nil, 0o600))
	var y domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "y", Pool: "dev"}, &y))
	passEventually(t, root.Dir, func() bool {
		got := show(y.ID).Ticket
		return got.Status == domain.StatusInProgress && got.Holder != "" && got.Holder != a
	}, "another session works Y while A is parked")

	var pools rpc.PoolStatusResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodPoolList, struct{}{}, &pools))
	require.Len(t, pools.Pools, 1)
	assert.Equal(t, 1, pools.Pools[0].Parked)
	assert.Equal(t, 1, pools.Pools[0].Open, "the parked session is not an open slot")

	// X's wait ends: it is pre-assigned to A, but the pool is full, so A
	// stays suspended until Y's session is gone.
	require.NoError(t, admin.Call(t.Context(), rpc.MethodWaitClear, rpc.WaitClearParams{ID: x.ID, Reason: "checks green"}, nil))
	assert.Equal(t, a, show(x.ID).Ticket.Holder)
	pass(t, root.Dir, 3)
	assert.True(t, sessionFields(t, show(a).Ticket).Suspended, "no free slot yet")

	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))
	waitClosed(t, root, admin, y.ID)
	waitClosed(t, root, admin, x.ID)
	passEventually(t, root.Dir, func() bool { return show(a).Ticket.Status == domain.StatusClosed }, "A ends")
	got := show(a).Ticket
	assert.Equal(t, "ticket_limit", got.Resolution)
	assert.Equal(t, 1, sessionFields(t, got).Tickets, "the parked ticket is counted once")
}

// TestMaxTicketsParkedGone (JOB-189): a parked session whose ticket was
// closed by someone else ends instead of staying suspended forever.
func TestMaxTicketsParkedGone(t *testing.T) {
	root, start := poolRoot(t, "1m")
	writeSingleUseConfig(t, root, 1, 2)
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	show := func(id string) rpc.ShowResult {
		var got rpc.ShowResult
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: id}, &got))
		return got
	}

	park, err := json.Marshal(rpc.WaitAddParams{
		Reason: "CI", Resume: true,
		Conds: []domain.WaitCond{{Type: domain.WaitChecks, Repo: "acme/widget", PR: 5, State: "success"}},
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "park"), park, 0o600))
	var x domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "x", Pool: "dev"}, &x))
	var a string
	passEventually(t, root.Dir, func() bool {
		for _, s := range openSessions(t, admin) {
			if f := sessionFields(t, s); f.Suspended && f.Parked == x.ID {
				a = s.ID
				return true
			}
		}
		return false
	}, "A parks X and is suspended")

	// Someone closes X by hand: A has nothing to resume for.
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketClose, rpc.CloseParams{ID: x.ID, Comment: "done by hand"}, nil))
	passEventually(t, root.Dir, func() bool { return show(a).Ticket.Status == domain.StatusClosed }, "A ends")
	assert.Equal(t, "parked_gone", show(a).Ticket.Resolution)
}
