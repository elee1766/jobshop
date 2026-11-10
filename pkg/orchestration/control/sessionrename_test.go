package control_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// TestSessionRename (#626): a session renames itself, names need not be
// unique, empty restores the default, and renaming another session needs
// a grant.
func TestSessionRename(t *testing.T) {
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
	var a, b rpc.SessionNewResult
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionNew, rpc.SessionNewParams{Agent: "jcode", Pool: "dev", DisplayName: "fix ci"}, &a))
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionNew, rpc.SessionNewParams{Agent: "jcode", Pool: "dev", Title: "dev jcode"}, &b))

	list := func() map[string]string {
		var infos []rpc.SessionInfo
		require.NoError(t, admin.Call(ctx, rpc.MethodSessionList, rpc.SessionListParams{}, &infos))
		out := map[string]string{}
		for _, in := range infos {
			out[in.ID] = in.DisplayName
		}
		return out
	}
	assert.Equal(t, map[string]string{a.Ticket.ID: "fix ci", b.Ticket.ID: "dev jcode"}, list(), "set at spawn; default is the session title")

	sb := dial(t, b.Socket)
	var info rpc.SessionInfo
	require.NoError(t, sb.Call(ctx, rpc.MethodSessionRename, rpc.SessionRenameParams{Name: "  fix   ci "}, &info))
	assert.Equal(t, "fix ci", info.DisplayName, "whitespace collapsed")
	assert.Equal(t, map[string]string{a.Ticket.ID: "fix ci", b.Ticket.ID: "fix ci"}, list(), "not unique")

	assert.ErrorIs(t, sb.Call(ctx, rpc.MethodSessionRename, rpc.SessionRenameParams{ID: a.Ticket.ID, Name: "x"}, nil), rpc.ErrForbidden)
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionRename, rpc.SessionRenameParams{ID: a.Ticket.ID, Name: "other"}, nil))
	require.NoError(t, sb.Call(ctx, rpc.MethodSessionRename, rpc.SessionRenameParams{Name: ""}, nil))
	assert.Equal(t, map[string]string{a.Ticket.ID: "other", b.Ticket.ID: "dev jcode"}, list())
	assert.Error(t, admin.Call(ctx, rpc.MethodSessionRename, rpc.SessionRenameParams{Name: "x"}, nil), "admin needs an id")
}
