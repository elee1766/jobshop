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
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

const deskExecYAML = `projects:
  js: {}
desk_commands:
  echo:
    command: cat
  text:
    command: echo hello
  fail:
    command: echo broken >&2; exit 3
`

// desk.exec runs a desk_commands entry with JSON stdin in the root, and
// records a history row with the actor on the named ticket (#542).
func TestDeskExec(t *testing.T) {
	t.Parallel()
	dir, err := os.MkdirTemp("", "js")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), fmt.Append(nil, deskExecYAML), 0o600))
	root := config.Root{Dir: dir}
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- control.Run(ctx, root, control.Options{Tick: time.Hour}) }()
	t.Cleanup(func() { cancel(); require.NoError(t, <-done) })
	admin := dial(t, root.AdminSocket())

	var tk domain.Ticket
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "t"}, &tk))

	var res rpc.DeskExecResult
	require.NoError(t, admin.Call(ctx, rpc.MethodDeskExec, rpc.DeskExecParams{Name: "echo", Input: []byte(`{"a":1}`), Ticket: tk.ID}, &res))
	assert.JSONEq(t, `{"a":1}`, string(res.Output))

	res = rpc.DeskExecResult{}
	require.NoError(t, admin.Call(ctx, rpc.MethodDeskExec, rpc.DeskExecParams{Name: "text"}, &res))
	assert.Equal(t, "hello\n", res.Text)

	err = admin.Call(ctx, rpc.MethodDeskExec, rpc.DeskExecParams{Name: "fail", Ticket: tk.ID}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exited 3: broken")

	err = admin.Call(ctx, rpc.MethodDeskExec, rpc.DeskExecParams{Name: "nope"}, nil)
	require.ErrorIs(t, err, domain.ErrNotFound)

	var show rpc.ShowResult
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketShow, rpc.IDParams{ID: tk.ID}, &show))
	var execs []store.HistoryEntry
	for _, h := range show.History {
		if h.Op == "desk_exec" {
			execs = append(execs, h)
		}
	}
	require.Len(t, execs, 2)
	assert.JSONEq(t, `{"name":"echo","exit":0}`, string(execs[0].Diff))
	assert.Contains(t, string(execs[1].Diff), `"exit":3`)
	assert.NotEmpty(t, execs[0].Actor)
}
