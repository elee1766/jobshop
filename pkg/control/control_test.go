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
	t.Parallel()
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

	var follow store.Ticket
	require.NoError(t, s.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "follow-up"}, &follow))
	assert.Equal(t, "dev", follow.Pool, "a session's task with no pool goes to its own pool")

	require.NoError(t, admin.Call(ctx, rpc.MethodTicketClose, rpc.CloseParams{ID: sess.Ticket.ID}, nil))
	var got store.Ticket
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketShow, rpc.IDParams{ID: task.ID}, &struct {
		Ticket *store.Ticket `json:"ticket"`
	}{&got}))
	assert.Equal(t, store.StatusOpen, got.Status, "closing a session releases its claims")
	_, err = os.Stat(sess.Socket)
	assert.ErrorIs(t, err, os.ErrNotExist, "session socket is removed")
}

// dialWait bounds dial only on failure: dial returns as soon as the socket
// answers. With ~40 parallel controllers under -race -count=5, startup
// (migrate, config and pack load) took over 5s (#365).
const dialWait = 2 * time.Minute

func dial(t *testing.T, path string) *rpc.Client {
	t.Helper()
	var c *rpc.Client
	require.Eventually(t, func() bool {
		var err error
		c, err = rpc.Dial(t.Context(), path)
		return err == nil
	}, dialWait, 10*time.Millisecond)
	t.Cleanup(func() { _ = c.Close() })
	return c
}

// TestSessionListEnded filters closed sessions by resolution and close time,
// the one-command evidence the gasserver runbook uses.
func TestSessionListEnded(t *testing.T) {
	t.Parallel()
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
	newSess := func(resolution string) string {
		var s rpc.SessionNewResult
		require.NoError(t, admin.Call(ctx, rpc.MethodSessionNew, rpc.SessionNewParams{Agent: "jcode", Pool: "dev"}, &s))
		if resolution != "" {
			require.NoError(t, admin.Call(ctx, rpc.MethodTicketClose, rpc.CloseParams{ID: s.Ticket.ID, Resolution: resolution}, nil))
		}
		return s.Ticket.ID
	}
	crashed := newSess("crashed")
	newSess("done")
	newSess("")

	ids := func(p rpc.SessionListParams) []string {
		var infos []rpc.SessionInfo
		require.NoError(t, admin.Call(ctx, rpc.MethodSessionList, p, &infos))
		out := []string{}
		for _, i := range infos {
			out = append(out, i.ID)
		}
		return out
	}
	assert.Equal(t, []string{crashed}, ids(rpc.SessionListParams{State: "crashed"}))
	assert.Equal(t, []string{crashed}, ids(rpc.SessionListParams{State: "crashed", Since: time.Hour}))
	assert.Len(t, ids(rpc.SessionListParams{Since: time.Hour}), 2, "closed sessions only")
	assert.Empty(t, ids(rpc.SessionListParams{State: "crashed", Since: time.Nanosecond}))
	assert.Len(t, ids(rpc.SessionListParams{All: true}), 3)

	err = admin.Call(ctx, rpc.MethodSessionList, rpc.SessionListParams{Since: -time.Hour}, nil)
	assert.ErrorIs(t, err, store.ErrInvalid)
}

// TestMetaPacks checks that meta keys must be built-in or listed packs.
func TestMetaPacks(t *testing.T) {
	t.Parallel()
	dir, err := os.MkdirTemp("", "js")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), []byte("projects:\n  js: {}\nmeta_packs: [gx]\n"), 0o600))
	root := config.Root{Dir: dir}

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- control.Run(ctx, root, control.Options{Tick: time.Hour}) }()
	t.Cleanup(func() { cancel(); require.NoError(t, <-done) })

	admin := dial(t, root.AdminSocket())
	var tk store.Ticket
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "a", Meta: []byte(`{"github":{"n":1},"gx":{}}`)}, &tk))
	err = admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "b", Meta: []byte(`{"gihtub":{}}`)}, nil)
	require.ErrorContains(t, err, "unknown pack")
	err = admin.Call(ctx, rpc.MethodTicketUpdate, rpc.UpdateParams{ID: tk.ID, Meta: []byte(`{"nope":{}}`)}, nil)
	require.ErrorContains(t, err, "unknown pack")
}
