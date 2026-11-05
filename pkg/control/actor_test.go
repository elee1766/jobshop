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

// TestMetaActor: an admin call may name its actor (Desk's proxy user, #515)
// and history records it; a session socket cannot.
func TestMetaActor(t *testing.T) {
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
	actx := rpc.WithActor(ctx, "desk:alice")
	var task store.Ticket
	require.NoError(t, admin.Call(actx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))
	assert.Equal(t, "desk:alice", task.CreatedBy)
	require.NoError(t, admin.Call(actx, rpc.MethodCommentAdd, rpc.CommentParams{ID: task.ID, Body: "hi"}, nil))
	var plain store.Ticket
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "u", Pool: "dev"}, &plain))
	assert.Equal(t, "admin", plain.CreatedBy, "no actor: unchanged")

	var show rpc.ShowResult
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketShow, rpc.IDParams{ID: task.ID}, &show))
	require.NotEmpty(t, show.History)
	assert.Equal(t, "desk:alice", show.History[0].Actor)
	require.Len(t, show.Comments, 1)
	assert.Equal(t, "desk:alice", show.Comments[0].Author)

	err = admin.Call(rpc.WithActor(ctx, "bad\x00name"), rpc.MethodWhoami, struct{}{}, nil)
	require.ErrorIs(t, err, store.ErrInvalid)

	var sess rpc.SessionNewResult
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionNew, rpc.SessionNewParams{Agent: "jcode", Pool: "dev"}, &sess))
	s := dial(t, sess.Socket)
	err = s.Call(rpc.WithActor(ctx, "desk:alice"), rpc.MethodWhoami, struct{}{}, nil)
	require.ErrorIs(t, err, rpc.ErrForbidden, "a session socket cannot set the actor")
	var who rpc.Whoami
	require.NoError(t, s.Call(ctx, rpc.MethodWhoami, struct{}{}, &who))
	assert.Equal(t, sess.Ticket.ID, who.Actor)
}
