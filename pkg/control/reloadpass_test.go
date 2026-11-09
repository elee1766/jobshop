package control_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
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

// holdLog blocks the controller goroutine that logs msg until release is
// closed, after telling held. A supervisor pass logs under tickMu, so
// blocking one of its records holds the pass where it is. Since JOB-103
// nothing in a pass runs a command or network call to hold it in.
type holdLog struct {
	slog.Handler
	msg     string
	held    chan<- struct{}
	release <-chan struct{}
}

func (h holdLog) Handle(ctx context.Context, r slog.Record) error {
	if r.Message == h.msg {
		select {
		case h.held <- struct{}{}:
			<-h.release
		default:
		}
	}
	return h.Handler.Handle(ctx, r)
}

func (h holdLog) WithAttrs(a []slog.Attr) slog.Handler {
	return holdLog{h.Handler.WithAttrs(a), h.msg, h.held, h.release}
}

func (h holdLog) WithGroup(g string) slog.Handler {
	return holdLog{h.Handler.WithGroup(g), h.msg, h.held, h.release}
}

// TestReloadDuringPass (JOB-104): a reload while a supervisor pass is
// held open returns at once, the held pass finishes on the config it
// started with, and the next pass uses the new one. The pass is held in
// the log record of a-hold firing; a ticket labeled go lets the triggers
// fire.
func TestReloadDuringPass(t *testing.T) {
	held := make(chan struct{})
	release := make(chan struct{})
	log := slog.New(holdLog{testLog(t).Handler(), "trigger fired", held, release})
	root, start := poolRootOpt(t, "1m", control.Options{Log: log})
	write := func(triggers string) {
		t.Helper()
		cfg := fmt.Sprintf(reloadPassYAML, os.Args[0], root.Dir, triggers)
		require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.FileName), []byte(cfg), 0o600))
	}
	// Triggers run in name order: a-hold fires first and its log record
	// holds the pass, then b-old files its ticket in the same pass.
	write(`  a-hold:
    when: {label: go, open_above: 0}
    ticket: {title: hold, pool: dev}
  b-old:
    when: {label: go, open_above: 0}
    ticket: {title: from-old, pool: dev}
`)
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	titles := func() []string {
		t.Helper()
		var ts []domain.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: domain.KindTask}, &ts))
		var out []string
		for _, tk := range ts {
			if tk.Title != "gate" && tk.Title != "hold" {
				out = append(out, tk.Title)
			}
		}
		slices.Sort(out)
		return out
	}
	pass(t, root.Dir, 1)
	require.Empty(t, titles(), "no gate: nothing fires")

	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "gate", Labels: []string{"go"}}, nil))
	v, ok := passChans.Load(root.Dir)
	require.True(t, ok)
	done := make(chan struct{})
	select {
	case v.(chan chan struct{}) <- done:
	case <-time.After(dialWait):
		t.Fatal("controller took no pass")
	}
	select {
	case <-held:
	case <-time.After(dialWait):
		t.Fatal("the pass never fired a-hold")
	}

	write(`  b-new:
    when: {label: go, open_above: 0}
    ticket: {title: from-new, pool: dev}
`)
	// Before JOB-104 reload waited for the pass: this call would time out.
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	require.NoError(t, admin.Call(ctx, rpc.MethodReload, struct{}{}, nil), "reload returns while the pass is held")
	require.Empty(t, titles(), "the held pass has not reached b-old yet")

	close(release)
	select {
	case <-done:
	case <-time.After(dialWait):
		t.Fatal("held pass did not finish")
	}
	require.Equal(t, []string{"from-old"}, titles(), "the held pass ran on the config it started with")

	pass(t, root.Dir, 1)
	require.Equal(t, []string{"from-new", "from-old"}, titles(), "the next pass uses the new config")
}
