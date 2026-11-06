package control_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// reloadPassYAML has a manual-only pool and the given triggers. The
// tickets the triggers file wait in it; nothing claims them.
const reloadPassYAML = `projects:
  js: {}
agents:
  fake:
    command: %q
    args: [fake-agent]
pools:
  dev:
    project: js
    agent: fake
    max_sessions: 0
    dir: %q
triggers:
%s`

// TestReloadDuringPass (JOB-104): a reload while a supervisor pass is
// held open in a trigger check returns at once, the held pass finishes on
// the config it started with, and the next pass uses the new one.
func TestReloadDuringPass(t *testing.T) {
	root, start := poolRoot(t, "1m")
	gate := filepath.Join(root.Dir, "gate")
	fifo := filepath.Join(root.Dir, "hold")
	require.NoError(t, syscall.Mkfifo(fifo, 0o600))
	write := func(triggers string) {
		t.Helper()
		cfg := fmt.Sprintf(reloadPassYAML, os.Args[0], root.Dir, triggers)
		require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.FileName), []byte(cfg), 0o600))
	}
	// Triggers run in name order: a-hold blocks on the fifo once the gate
	// exists, then b-old files its ticket in the same pass.
	write(fmt.Sprintf(`  a-hold:
    when: {check: "test -e %[1]s || exit 1; read x < %[2]s; exit 1"}
    ticket: {title: never, pool: dev}
  b-old:
    when: {check: "test -e %[1]s"}
    ticket: {title: from-old, pool: dev}
`, gate, fifo))
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	titles := func() []string {
		t.Helper()
		var ts []store.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindTask}, &ts))
		var out []string
		for _, tk := range ts {
			out = append(out, tk.Title)
		}
		slices.Sort(out)
		return out
	}
	pass(t, root.Dir, 1)
	require.Empty(t, titles(), "no gate: nothing fires")

	require.NoError(t, os.WriteFile(gate, nil, 0o600))
	v, ok := passChans.Load(root.Dir)
	require.True(t, ok)
	done := make(chan struct{})
	select {
	case v.(chan chan struct{}) <- done:
	case <-time.After(dialWait):
		t.Fatal("controller took no pass")
	}
	// Opening the fifo for writing returns once a-hold's check opened it
	// for reading: the pass is held there.
	opened := make(chan *os.File, 1)
	go func() {
		f, err := os.OpenFile(fifo, os.O_WRONLY, 0)
		if err != nil {
			f = nil
		}
		opened <- f
	}()
	var w *os.File
	select {
	case w = <-opened:
		require.NotNil(t, w, "open fifo")
	case <-time.After(dialWait):
		t.Fatal("the pass never reached the held check")
	}

	write(fmt.Sprintf(`  b-new:
    when: {check: "test -e %s"}
    ticket: {title: from-new, pool: dev}
`, gate))
	// Before JOB-104 reload waited for the pass: this call would time out.
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	require.NoError(t, admin.Call(ctx, rpc.MethodReload, struct{}{}, nil), "reload returns while the pass is held")
	require.Empty(t, titles(), "the held pass has not reached b-old yet")

	_, err := w.WriteString("go\n")
	require.NoError(t, err)
	require.NoError(t, w.Close())
	select {
	case <-done:
	case <-time.After(dialWait):
		t.Fatal("held pass did not finish")
	}
	require.Equal(t, []string{"from-old"}, titles(), "the held pass ran on the config it started with")

	pass(t, root.Dir, 1)
	require.Equal(t, []string{"from-new", "from-old"}, titles(), "the next pass uses the new config")
}
