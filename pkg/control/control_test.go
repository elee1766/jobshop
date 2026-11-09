package control_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/control/controltest"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
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
	var task domain.Ticket
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

	var follow domain.Ticket
	require.NoError(t, s.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "follow-up"}, &follow))
	assert.Equal(t, "dev", follow.Pool, "a session's task with no pool goes to its own pool")

	require.NoError(t, admin.Call(ctx, rpc.MethodTicketClose, rpc.CloseParams{ID: sess.Ticket.ID}, nil))
	var got domain.Ticket
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketShow, rpc.IDParams{ID: task.ID}, &struct {
		Ticket *domain.Ticket `json:"ticket"`
	}{&got}))
	assert.Equal(t, domain.StatusOpen, got.Status, "closing a session releases its claims")
	_, err = os.Stat(sess.Socket)
	assert.ErrorIs(t, err, os.ErrNotExist, "session socket is removed")
}

// dialWait bounds dial and the other Eventually waits in these tests only
// on failure: they return as soon as the condition holds. With ~40 parallel
// controllers under -race -count=5, startup (migrate, config and pack load)
// and trigger passes took over 5s (#365).
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
	assert.ErrorIs(t, err, domain.ErrInvalid)
}

// Session lists have no cursor, so old history must not hide newer chats.
func TestSessionListBeyondDefaultPage(t *testing.T) {
	t.Parallel()
	root := controltest.NewRoot(t, "projects:\n  js: {}\n")
	// A page of closed sessions, written in one transaction: one
	// session.new and close per session took 11s under -race (JOB-117).
	// Their ended state is saved, as for a session that closed normally,
	// so startup has none to backfill.
	seedStore(t, root, func(st store.Store) {
		g := store.Graph{Close: map[string]string{}}
		for i := range store.DefaultLimit {
			g.Tickets = append(g.Tickets, store.NewTicket{ID: fmt.Sprintf("js-s-old%d", i), Project: "js", Kind: domain.KindSession, Title: "dev jcode session", Pool: "dev", Fields: []byte(`{"agent":"jcode","ended":{}}`)})
			g.Close[g.Tickets[i].ID] = "done"
		}
		_, err := st.Apply(t.Context(), "test", g)
		require.NoError(t, err)
	})
	c := controltest.Start(t, root, control.Options{})
	var newest rpc.SessionNewResult
	c.Call(rpc.MethodSessionNew, rpc.SessionNewParams{Agent: "jcode", Pool: "dev"}, &newest)
	var sessions []rpc.SessionInfo
	c.Call(rpc.MethodSessionList, rpc.SessionListParams{All: true}, &sessions)
	require.Len(t, sessions, store.DefaultLimit+1)
	ids := make([]string, 0, len(sessions))
	for _, session := range sessions {
		ids = append(ids, session.ID)
	}
	assert.Contains(t, ids, newest.Ticket.ID)
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
	var tk domain.Ticket
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "a", Meta: []byte(`{"github":{"n":1},"gx":{}}`)}, &tk))
	err = admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "b", Meta: []byte(`{"gihtub":{}}`)}, nil)
	require.ErrorContains(t, err, "unknown pack")
	err = admin.Call(ctx, rpc.MethodTicketUpdate, rpc.UpdateParams{ID: tk.ID, Meta: []byte(`{"nope":{}}`)}, nil)
	require.ErrorContains(t, err, "unknown pack")
}
