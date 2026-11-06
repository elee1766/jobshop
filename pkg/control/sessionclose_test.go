package control_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// TestSessionClose closes a working session (JOB-111): the dry run lists
// its claim, a session socket without the grant is refused, and the close
// ends the session at once, hands the ticket back to its pool without a
// failure, keeps the journal readable and lets a new session take the work.
func TestSessionClose(t *testing.T) {
	root, start := poolRoot(t, "1m")
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "block"), nil, 0o600))
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	show := func(id string) rpc.ShowResult {
		var got rpc.ShowResult
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: id}, &got))
		return got
	}

	var task store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "work", Pool: "dev"}, &task))
	passEventually(t, root.Dir, func() bool {
		_, err := os.Stat(filepath.Join(root.Dir, "started"))
		return err == nil
	}, "turn started")
	sessID := show(task.ID).Ticket.Holder
	require.NotEmpty(t, sessID)

	var dry rpc.SessionCloseResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionClose, rpc.SessionCloseParams{ID: sessID, DryRun: true}, &dry))
	assert.True(t, dry.DryRun)
	require.Len(t, dry.Held, 1)
	assert.Equal(t, rpc.HeldTicket{ID: task.ID, Title: "work", Kind: store.KindTask, Status: store.StatusInProgress, Pool: "dev"}, dry.Held[0])
	assert.Equal(t, store.StatusOpen, show(sessID).Ticket.Status, "dry run leaves the session open")

	sess := dial(t, filepath.Join(root.SessionDir(sessID), "cli.sock"))
	assert.ErrorIs(t, sess.Call(t.Context(), rpc.MethodSessionClose, rpc.SessionCloseParams{ID: sessID}, nil), rpc.ErrForbidden)
	err := admin.Call(t.Context(), rpc.MethodSessionClose, rpc.SessionCloseParams{ID: sessID, Handoff: "nope"}, nil)
	assert.ErrorIs(t, err, store.ErrInvalid)

	var res rpc.SessionCloseResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionClose, rpc.SessionCloseParams{ID: sessID, Reason: "answered"}, &res))
	require.NotNil(t, res.Session)
	assert.Equal(t, store.StatusClosed, res.Session.Status)
	assert.Equal(t, "closed", res.Session.Resolution)
	require.Len(t, res.Held, 1)

	got := show(task.ID)
	assert.Equal(t, store.StatusOpen, got.Ticket.Status)
	assert.Empty(t, got.Ticket.Holder)
	assert.Equal(t, "dev", got.Ticket.Pool)
	require.NotEmpty(t, got.Comments)
	assert.Contains(t, got.Comments[len(got.Comments)-1].Body, "handed back to pool dev")
	sc := show(sessID)
	assert.True(t, strings.HasPrefix(sc.Comments[len(sc.Comments)-1].Body, "Session closed by admin: answered"), sc.Comments[len(sc.Comments)-1].Body)

	var j rpc.SessionJournalResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionJournal, rpc.SessionJournalParams{ID: sessID}, &j))
	assert.NotEmpty(t, j.Frames, "journal of a closed session stays readable")
	assert.ErrorIs(t, admin.Call(t.Context(), rpc.MethodSessionClose, rpc.SessionCloseParams{ID: sessID}, nil), store.ErrClosed)

	// The slot is free: a new session takes the handed-back ticket.
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))
	passEventually(t, root.Dir, func() bool { return show(task.ID).Ticket.Status == store.StatusClosed }, "a new session finishes the ticket")
	assert.NotEqual(t, sessID, show(task.ID).Ticket.Holder)
}
