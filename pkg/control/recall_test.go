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

// TestRecall: a comment added through the controller is found by the next
// recall call, and sessions only see their own project.
func TestRecall(t *testing.T) {
	t.Parallel()
	dir, err := os.MkdirTemp("", "js")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName),
		[]byte("default_project: js\nprojects:\n  js: {}\n  other: {}\n"), 0o600))
	root := config.Root{Dir: dir}
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- control.Run(ctx, root, control.Options{Tick: time.Hour}) }()
	t.Cleanup(func() { cancel(); require.NoError(t, <-done) })

	admin := dial(t, root.AdminSocket())
	var a, b store.Ticket
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Project: "js", Title: "a", Pool: "dev"}, &a))
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Project: "other", Title: "b", Pool: "dev"}, &b))
	require.NoError(t, admin.Call(ctx, rpc.MethodCommentAdd, rpc.CommentParams{ID: a.ID, Body: "vendorHash mismatch fixed by fakeHash"}, nil))
	require.NoError(t, admin.Call(ctx, rpc.MethodCommentAdd, rpc.CommentParams{ID: b.ID, Body: "vendorHash mismatch elsewhere"}, nil))
	// Learnings have their own index (#579); recall merges them in.
	require.NoError(t, admin.Call(ctx, rpc.MethodLearnAdd, rpc.LearnParams{Kind: store.LearningWorkaround, Summary: "vendorHash mismatch in other", Ticket: b.ID}, nil))
	require.NoError(t, admin.Call(ctx, rpc.MethodLearnAdd, rpc.LearnParams{Kind: store.LearningWorkaround, Summary: "vendorHash: paste the got value"}, nil))

	var hits []store.RecallHit
	require.NoError(t, admin.Call(ctx, rpc.MethodRecall, rpc.RecallParams{Query: "vendorHash mismatch"}, &hits))
	assert.Len(t, hits, 3, "admin with no project searches everything; all words match first")
	require.NoError(t, admin.Call(ctx, rpc.MethodRecall, rpc.RecallParams{Query: "vendorHash", Kind: store.RecallLearning}, &hits))
	assert.Len(t, hits, 2)

	var sess rpc.SessionNewResult
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionNew, rpc.SessionNewParams{Agent: "jcode", Pool: "dev", Project: "js"}, &sess))
	s := dial(t, sess.Socket)
	require.NoError(t, s.Call(ctx, rpc.MethodRecall, rpc.RecallParams{Query: "vendorHash"}, &hits))
	require.Len(t, hits, 2, "own comment and the global learning")
	assert.Equal(t, a.ID, hits[0].TicketID)
	assert.Equal(t, store.RecallComment, hits[0].Kind)
	assert.Equal(t, store.RecallLearning, hits[1].Kind)
	assert.ErrorIs(t, s.Call(ctx, rpc.MethodRecall, rpc.RecallParams{Query: "x", Project: "other"}, nil), rpc.ErrForbidden)
}
