package control_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// TestPoolInterrupt (JOB-206): with interrupt: queue an interrupt message
// does not cancel a running turn and is delivered after it; with
// interrupt: refuse interrupt sends to the pool fail and plain ones pass.
func TestPoolInterrupt(t *testing.T) {
	root, start := poolRootOpt(t, "1h", control.Options{MessagePoll: 50 * time.Millisecond})
	cfg, err := os.ReadFile(root.Config())
	require.NoError(t, err)
	// The dev pool is the config's last block.
	setMode := func(mode string) {
		require.NoError(t, os.WriteFile(root.Config(), append(append([]byte{}, cfg...), "    interrupt: "+mode+"\n"...), 0o600))
	}
	setMode(config.InterruptQueue)
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	show := func(id string) domain.Ticket {
		var got rpc.ShowResult
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: id}, &got))
		return got.Ticket
	}
	send := func(to string) (domain.Ticket, error) {
		f, err := json.Marshal(domain.MessageFields{To: to, Interrupt: true})
		require.NoError(t, err)
		var m domain.Ticket
		err = admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Kind: domain.KindMessage, Title: "stop that", Fields: f}, &m)
		return m, err
	}

	// A blocked turn keeps running with an interrupt message waiting.
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "block"), nil, 0o600))
	var work domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "b", Pool: "dev"}, &work))
	passEventually(t, root.Dir, func() bool {
		_, err := os.Stat(filepath.Join(root.Dir, "started"))
		return err == nil
	}, "turn started")
	var sessions []domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: domain.KindSession}, &sessions))
	require.Len(t, sessions, 1)
	sess := sessions[0].ID
	urgent, err := send(sess)
	require.NoError(t, err)
	// Many message polls pass in fake time; none cancels the turn.
	for range 20 {
		pass(t, root.Dir, 1)
		advance(t, root.Dir, 50*time.Millisecond)
	}
	assert.Equal(t, domain.StatusOpen, show(urgent.ID).Status, "queued until the turn ends")
	assert.Equal(t, domain.StatusInProgress, show(work.ID).Status)

	// The turn finishes the ticket, then the message is delivered.
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))
	passEventually(t, root.Dir, func() bool {
		return show(work.ID).Status == domain.StatusClosed && show(urgent.ID).Status == domain.StatusClosed
	}, "ticket closed, then message delivered")
	assert.Equal(t, "delivered", show(urgent.ID).Resolution)

	// refuse: interrupt sends fail by session id, key and pool; a plain
	// message still goes through.
	setMode(config.InterruptRefuse)
	require.NoError(t, admin.Call(t.Context(), rpc.MethodReload, struct{}{}, nil))
	_, err = send("pool:dev")
	require.ErrorIs(t, err, rpc.ErrForbidden)
	_, err = send(sess)
	require.ErrorIs(t, err, rpc.ErrForbidden)
	err = admin.Call(t.Context(), rpc.MethodSessionSend, rpc.SessionSendParams{ID: sess, Text: "stop", Interrupt: true}, nil)
	require.ErrorIs(t, err, rpc.ErrForbidden)
	err = admin.Call(t.Context(), rpc.MethodSessionSend, rpc.SessionSendParams{Pool: "dev", Key: "k", Text: "stop", Interrupt: true}, nil)
	require.ErrorIs(t, err, rpc.ErrForbidden)
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionSend, rpc.SessionSendParams{ID: sess, Text: "later"}, nil))
	// Another pool's recipient is not affected.
	_, err = send("pool:other")
	require.NoError(t, err)
}
