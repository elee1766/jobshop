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
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// The action hook answers from the event: push to main is denied, merge
// asks a human, comment is allowed by empty output, and "crash" exits 2.
// The claim hook refuses tickets titled "forbidden".
const hookYAML = `projects:
  js: {}
agents:
  fake:
    command: false
pools:
  dev:
    agent: fake
    dir: %[1]q
hooks:
  - on: action
    command: |
      ev=$(cat); echo "$ev" >> events.jsonl
      case "$ev" in
        *'"target":"main"'*) echo '{"decision":"deny","reason":"no pushes to main"}' ;;
        *'"action":"merge"'*) echo '{"decision":"ask_human","reason":"merges need a human"}' ;;
        *'"action":"crash"'*) exit 2 ;;
      esac
  - on: claim
    pools: [dev]
    command: |
      grep -q '"title":"forbidden"' && echo '{"decision":"deny","reason":"not this one"}' || true
`

func TestGuardHooks(t *testing.T) {
	t.Parallel()
	dir, err := os.MkdirTemp("", "js")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), fmt.Appendf(nil, hookYAML, dir), 0o600))
	root := config.Root{Dir: dir}
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- control.Run(ctx, root, control.Options{Tick: time.Hour}) }()
	t.Cleanup(func() { cancel(); require.NoError(t, <-done) })

	admin := dial(t, root.AdminSocket())
	var bad, good domain.Ticket
	p0, p1 := 0, 1
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "forbidden", Pool: "dev", Priority: &p0}, &bad))
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "fine", Pool: "dev", Priority: &p1}, &good))
	var sess rpc.SessionNewResult
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionNew, rpc.SessionNewParams{Agent: "jcode", Pool: "dev"}, &sess))
	s := dial(t, sess.Socket)

	// Claim: the hook refuses the first pick, which is deferred; the next
	// claim gets the other ticket.
	err = s.Call(ctx, rpc.MethodClaim, rpc.ClaimParams{}, nil)
	require.ErrorIs(t, err, rpc.ErrForbidden)
	assert.Contains(t, err.Error(), "not this one")
	err = s.Call(ctx, rpc.MethodClaim, rpc.ClaimParams{ID: bad.ID}, nil)
	require.ErrorIs(t, err, rpc.ErrForbidden, "claim by id runs the hook too")
	var cl rpc.ClaimResult
	require.NoError(t, s.Call(ctx, rpc.MethodClaim, rpc.ClaimParams{}, &cl))
	assert.Equal(t, good.ID, cl.Ticket.ID)
	require.NoError(t, s.Call(ctx, rpc.MethodClaim, rpc.ClaimParams{}, &cl), "resuming held work skips the hook")

	guard := func(action, target string) rpc.GuardResult {
		var r rpc.GuardResult
		require.NoError(t, s.Call(ctx, rpc.MethodGuard, rpc.GuardParams{Action: action, Target: target}, &r))
		return r
	}
	assert.Equal(t, domain.GuardAllow, guard("comment", "").Decision)
	r := guard("push", "main")
	assert.Equal(t, domain.GuardDeny, r.Decision)
	assert.Equal(t, "no pushes to main", r.Reason)
	r = guard("crash", "")
	assert.Equal(t, domain.GuardDeny, r.Decision, "fail closed")
	assert.Contains(t, r.Reason, "exited 2")

	// ask_human files one needs-human ticket and reads it back.
	r = guard("merge", "pr/1")
	require.Equal(t, domain.GuardAskHuman, r.Decision)
	require.NotEmpty(t, r.Ticket)
	again := guard("merge", "pr/1")
	assert.Equal(t, r.Ticket, again.Ticket, "one ticket per action and target")
	assert.Equal(t, domain.GuardAskHuman, again.Decision)
	var ask domain.Ticket
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketShow, rpc.IDParams{ID: r.Ticket}, &struct {
		Ticket *domain.Ticket `json:"ticket"`
	}{&ask}))
	assert.Equal(t, []string{"needs-human"}, ask.Labels)
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketClose, rpc.CloseParams{ID: r.Ticket}, nil))
	assert.Equal(t, domain.GuardAllow, guard("merge", "pr/1").Decision, "closed done approves")

	r = guard("merge", "pr/2")
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketClose, rpc.CloseParams{ID: r.Ticket, Resolution: "wontfix"}, nil))
	assert.Equal(t, domain.GuardDeny, guard("merge", "pr/2").Decision, "closed otherwise refuses")

	events, err := os.ReadFile(filepath.Join(dir, "events.jsonl"))
	require.NoError(t, err)
	assert.Contains(t, string(events), `"session":"`+sess.Ticket.ID+`"`)
	assert.Contains(t, string(events), `"ticket":{"id":"`+good.ID+`"`)
}

func TestGuardNoHooks(t *testing.T) {
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
	var r rpc.GuardResult
	require.NoError(t, admin.Call(ctx, rpc.MethodGuard, rpc.GuardParams{Action: "push"}, &r))
	assert.Equal(t, domain.GuardAllow, r.Decision)
}
