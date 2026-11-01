package control_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// TestSessionScope drives a real controller over its sockets: a session can
// only act as itself, and closing the session frees its claims.
func TestSessionScope(t *testing.T) {
	// Short path: unix socket paths are limited to 108 bytes.
	dir, err := os.MkdirTemp("", "js")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), []byte("projects:\n  js: {}\n"), 0o600))
	root := config.Root{Dir: dir}

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- control.Run(ctx, root, control.Options{Tick: time.Hour}) }()
	t.Cleanup(func() { cancel(); require.NoError(t, <-done) })

	admin := dial(t, root.AdminSocket())
	var task store.Ticket
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))
	var sess rpc.SessionNewResult
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionNew, rpc.SessionNewParams{Agent: "jcode", Pool: "dev"}, &sess))

	s := dial(t, sess.Socket)
	err = s.Call(ctx, rpc.MethodClaim, rpc.ClaimParams{Holder: "someone-else"}, nil)
	assert.ErrorIs(t, err, rpc.ErrForbidden)
	err = s.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "x", Holder: sess.Ticket.ID}, nil)
	assert.ErrorIs(t, err, rpc.ErrForbidden, "sessions cannot assign")

	var cl rpc.ClaimResult
	require.NoError(t, s.Call(ctx, rpc.MethodClaim, rpc.ClaimParams{}, &cl))
	assert.Equal(t, task.ID, cl.Ticket.ID)
	assert.Equal(t, sess.Ticket.ID, cl.Claim.Holder)
	err = s.Call(ctx, rpc.MethodClaim, rpc.ClaimParams{}, &cl)
	require.NoError(t, err, "claim --next resumes held work")

	require.NoError(t, admin.Call(ctx, rpc.MethodTicketClose, rpc.CloseParams{ID: sess.Ticket.ID}, nil))
	var got store.Ticket
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketShow, rpc.IDParams{ID: task.ID}, &struct {
		Ticket *store.Ticket `json:"ticket"`
	}{&got}))
	assert.Equal(t, store.StatusOpen, got.Status, "closing a session releases its claims")
	_, err = os.Stat(sess.Socket)
	assert.ErrorIs(t, err, os.ErrNotExist, "session socket is removed")
}

func dial(t *testing.T, path string) *rpc.Client {
	t.Helper()
	var c *rpc.Client
	require.Eventually(t, func() bool {
		var err error
		c, err = rpc.Dial(t.Context(), path)
		return err == nil
	}, 5*time.Second, 10*time.Millisecond)
	t.Cleanup(func() { _ = c.Close() })
	return c
}
