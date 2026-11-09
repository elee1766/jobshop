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
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/orchestration/control"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// TestTicketUpsert: one open ticket per key, updated in place, listed in
// attention while labeled needs-human, and closed when the problem
// clears (#577). A later report opens a new one.
func TestTicketUpsert(t *testing.T) {
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

	up := func(p rpc.UpsertParams) rpc.UpsertResult {
		t.Helper()
		var r rpc.UpsertResult
		require.NoError(t, admin.Call(ctx, rpc.MethodTicketUpsert, p, &r))
		return r
	}
	report := rpc.UpsertParams{Key: "deploy:rollback", Title: "self-deploy rolled back abc", Body: "health check failed", Labels: []string{"needs-human"}}

	assert.Equal(t, "none", up(rpc.UpsertParams{Key: "deploy:rollback", Close: true}).Op)
	first := up(report)
	require.Equal(t, "created", first.Op)
	assert.Equal(t, []string{"needs-human"}, first.Ticket.Labels)
	assert.Equal(t, "unchanged", up(report).Op, "a repeated failure makes no second ticket")
	report.Title = "self-deploy rolled back def"
	again := up(report)
	assert.Equal(t, "updated", again.Op)
	assert.Equal(t, first.Ticket.ID, again.Ticket.ID)

	var att rpc.AttentionResult
	require.NoError(t, admin.Call(ctx, rpc.MethodAttention, struct{}{}, &att))
	require.Len(t, att.Items, 1)
	assert.Equal(t, first.Ticket.ID, att.Items[0].ID)
	assert.Equal(t, "self-deploy rolled back def", att.Items[0].Title)

	closed := up(rpc.UpsertParams{Key: "deploy:rollback", Close: true, Body: "deployed ghi"})
	assert.Equal(t, "closed", closed.Op)
	assert.Equal(t, domain.StatusClosed, closed.Ticket.Status)
	require.NoError(t, admin.Call(ctx, rpc.MethodAttention, struct{}{}, &att))
	assert.Empty(t, att.Items)

	next := up(report)
	assert.Equal(t, "created", next.Op)
	assert.NotEqual(t, first.Ticket.ID, next.Ticket.ID)

	err = admin.Call(ctx, rpc.MethodTicketUpsert, rpc.UpsertParams{Key: "k"}, nil)
	require.ErrorContains(t, err, "needs a title")
}
