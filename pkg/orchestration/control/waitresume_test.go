package control_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// TestWaitResumeBusySession (JOB-120): session A parks X with resume,
// then claims Y. X's wait clears while A is mid-turn on Y, so X goes
// back to its pool instead of queueing behind Y, and a second, idle
// session claims it.
func TestWaitResumeBusySession(t *testing.T) {
	root, start := poolRoot(t, "1m")
	writePoolConfig(t, root, 2, "")
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	show := func(id string) rpc.ShowResult {
		var got rpc.ShowResult
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: id}, &got))
		return got
	}
	waiting := func(tk domain.Ticket) bool {
		var f struct {
			Wait *domain.Wait `json:"wait"`
		}
		return json.Unmarshal(tk.Fields, &f) == nil && f.Wait != nil
	}

	park, err := json.Marshal(rpc.WaitAddParams{
		Reason: "CI on PR #5", Resume: true,
		Conds: []domain.WaitCond{{Type: domain.WaitChecks, Repo: "acme/widget", PR: 5, State: "success"}},
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "park"), park, 0o600))
	var x domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "x", Pool: "dev"}, &x))
	passEventually(t, root.Dir, func() bool {
		got := show(x.ID).Ticket
		return got.Status == domain.StatusOpen && waiting(got)
	}, "A parks X")
	var w struct {
		Wait domain.Wait `json:"wait"`
	}
	require.NoError(t, json.Unmarshal(show(x.ID).Ticket.Fields, &w))
	a := w.Wait.Session
	require.NotEmpty(t, a, "the wait names the parking session")

	// A claims Y and stays mid-turn on it.
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "block"), nil, 0o600))
	var y domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "y", Pool: "dev"}, &y))
	passEventually(t, root.Dir, func() bool {
		_, err := os.Stat(filepath.Join(root.Dir, "started"))
		got := show(y.ID).Ticket
		return err == nil && got.Status == domain.StatusInProgress && got.Holder == a
	}, "A works on Y")

	require.NoError(t, admin.Call(t.Context(), rpc.MethodWaitClear, rpc.WaitClearParams{ID: x.ID, Reason: "checks green"}, nil))
	got := show(x.ID)
	assert.False(t, waiting(got.Ticket))
	assert.Empty(t, got.Ticket.Holder, "not pre-assigned to busy A")
	require.NotEmpty(t, got.Comments)
	assert.Contains(t, got.Comments[len(got.Comments)-1].Body, "Session "+a+" is busy on "+y.ID)

	passEventually(t, root.Dir, func() bool {
		got := show(x.ID).Ticket
		return got.Status == domain.StatusInProgress && got.Holder != "" && got.Holder != a
	}, "an idle session claims X while A is busy on Y")
	assert.Equal(t, domain.StatusInProgress, show(y.ID).Ticket.Status, "A still on Y")

	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))
	waitClosed(t, root, admin, x.ID)
	waitClosed(t, root, admin, y.ID)
}
