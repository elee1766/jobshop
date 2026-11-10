package control_test

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	host "github.com/gfx-labs/acphost/pkg/acphost"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// Runner state across a controller restart (JOB-196). Each test stops the
// controller while the runner holds state that is only worth anything if
// the next controller reads it back, and checks that behavior.

// sessionFieldsOf reads a session's stored fields.
func sessionFieldsOf(t *testing.T, admin *rpc.Client, id string) domain.SessionFields {
	t.Helper()
	var got rpc.ShowResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: id}, &got))
	f, err := domain.DecodeFields[domain.SessionFields](got.Ticket)
	require.NoError(t, err)
	return f
}

// promptLog is the prompts the fake agent received, oldest first.
func promptLog(t *testing.T, dir string) []string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "prompts"))
	if err != nil {
		return nil
	}
	var out []string
	for line := range strings.Lines(string(b)) {
		var s string
		require.NoError(t, json.Unmarshal([]byte(line), &s))
		out = append(out, s)
	}
	return out
}

// TestRestartKeepsBackoff stops the controller while a session waits out a
// provider rate limit: the next controller does not prompt until the stored
// backoff has passed, then retries the same ticket.
func TestRestartKeepsBackoff(t *testing.T) {
	backoffs := new(atomic.Int32)
	root, start := poolRootOpt(t, "1m", control.Options{
		Hosts: realHosts{}, Log: slog.New(backoffLog{testLog(t).Handler(), backoffs}),
	})
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "ratelimit"), []byte("1"), 0o600))
	stop := start()
	defer func() { stop() }()
	admin := dial(t, root.AdminSocket())

	var task domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))
	show := func(id string) domain.Ticket {
		var got rpc.ShowResult
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: id}, &got))
		return got.Ticket
	}
	// Passes only, no clock steps: the backoff must not run out here.
	var id string
	passUntil(t, root.Dir, func() bool {
		var list []domain.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: domain.KindSession}, &list))
		if len(list) != 1 {
			return false
		}
		id = list[0].ID
		return backoffs.Load() == 1 && sessionFieldsOf(t, admin, id).BackoffUntil != nil
	}, "backoff stored")
	until := *sessionFieldsOf(t, admin, id).BackoffUntil
	require.Len(t, promptLog(t, root.Dir), 1, "the rate limited prompt")

	stop()
	stop = start()
	admin = dial(t, root.AdminSocket())
	// The restarted runner reads the stored backoff and shows it.
	passUntil(t, root.Dir, func() bool {
		var list []rpc.SessionInfo
		require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionList, rpc.SessionListParams{}, &list))
		return len(list) == 1 && list[0].BackoffUntil != nil
	}, "backoff read after restart")
	var list []rpc.SessionInfo
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionList, rpc.SessionListParams{}, &list))
	assert.True(t, list[0].BackoffUntil.Equal(until), "same backoff end: %s vs %s", list[0].BackoffUntil, until)
	pass(t, root.Dir, 5)
	assert.Len(t, promptLog(t, root.Dir), 1, "no prompt while the backoff runs")
	assert.Equal(t, domain.StatusInProgress, show(task.ID).Status, "claim kept")

	// Past the backoff the ticket is retried and closed.
	passEventually(t, root.Dir, func() bool { return show(task.ID).Status == domain.StatusClosed }, "ticket not retried")
	assert.Len(t, promptLog(t, root.Dir), 2)
	assert.Nil(t, sessionFieldsOf(t, admin, id).BackoffUntil, "backoff cleared once it passed")
}

// TestRestartKeepsChatPoolPrompt restarts the controller between two chat
// messages: the second turn does not carry the pool prompt again.
func TestRestartKeepsChatPoolPrompt(t *testing.T) {
	root, start := poolRoot(t, "1m")
	f, err := os.OpenFile(filepath.Join(root.Dir, config.FileName), os.O_APPEND|os.O_WRONLY, 0o600)
	require.NoError(t, err)
	_, err = fmt.Fprintf(f, "  chat:\n    project: js\n    agent: fake\n    type: chat\n    prompt: CHAT-POOL-PROMPT\n    dir: %q\nprompt_rules:\n  - {name: chat-always, text: CHAT-ALWAYS, at: before, inject: always, match: {pools: [chat]}}\n", root.Dir)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	stop := start()
	defer func() { stop() }()
	admin := dial(t, root.AdminSocket())
	send := func(text string) rpc.SessionSendResult {
		t.Helper()
		var res rpc.SessionSendResult
		require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionSend, rpc.SessionSendParams{Pool: "chat", Key: "matrix:r:1", Text: text}, &res))
		return res
	}
	first := send("hello")
	passEventually(t, root.Dir, func() bool { return len(promptLog(t, root.Dir)) == 1 }, "first message delivered")
	passEventually(t, root.Dir, func() bool { return sessionFieldsOf(t, admin, first.Session).Delivered != nil }, "delivery stored")
	assert.Contains(t, promptLog(t, root.Dir)[0], "CHAT-POOL-PROMPT")
	assert.Contains(t, promptLog(t, root.Dir)[0], "CHAT-ALWAYS")
	stored := sessionFieldsOf(t, admin, first.Session)
	assert.Contains(t, stored.Injected, "@pool")

	stop()
	stop = start()
	admin = dial(t, root.AdminSocket())
	again := send("after the restart")
	assert.Equal(t, first.Session, again.Session)
	passEventually(t, root.Dir, func() bool { return len(promptLog(t, root.Dir)) == 2 }, "second message delivered")
	p := promptLog(t, root.Dir)[1]
	assert.Contains(t, p, "after the restart")
	assert.NotContains(t, p, "CHAT-POOL-PROMPT", "pool prompt not resent after a controller restart")
	assert.NotContains(t, p, "CHAT-ALWAYS", "pool prompt turn rules not resent: the chat is past its first message turn")
	assert.NotNil(t, sessionFieldsOf(t, admin, first.Session).Delivered)
}

// TestRestartKeepsRecoveryAndSummary respawns a host whose conversation
// cannot be reloaded, and restarts the controller before the next prompt:
// that prompt still opens with the recovery note and the journal summary.
func TestRestartKeepsRecoveryAndSummary(t *testing.T) {
	root, start := poolRootReal(t, "1m")
	stop := start()
	defer func() { stop() }()
	admin := dial(t, root.AdminSocket())
	show := func(id string) domain.Ticket {
		var got rpc.ShowResult
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: id}, &got))
		return got.Ticket
	}
	var first domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "first-ticket-title", Pool: "dev"}, &first))
	passEventually(t, root.Dir, func() bool { return show(first.ID).Status == domain.StatusClosed }, "first ticket not closed")
	var sessions []domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: domain.KindSession}, &sessions))
	require.Len(t, sessions, 1)
	id := sessions[0].ID
	sdir := root.SessionDir(id)
	pid := hostPidOf(t, sdir)

	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "loadfail"), nil, 0o600))
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionRestart, rpc.SessionRestartParams{ID: id}, nil))
	passEventually(t, root.Dir, func() bool {
		p := hostPidOf(t, sdir)
		f := sessionFieldsOf(t, admin, id)
		return p != 0 && p != pid && hostRunning(sdir) && f.SeedDue && f.RecoveryDue
	}, "host respawned with a new conversation")
	f := sessionFieldsOf(t, admin, id)
	assert.NotEmpty(t, f.ACPSession)
	pid = hostPidOf(t, sdir)

	stop()
	require.True(t, hostRunning(sdir), "host outlives the controller")
	stop = start()
	admin = dial(t, root.AdminSocket())
	var second domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "second-ticket-title", Pool: "dev"}, &second))
	passEventually(t, root.Dir, func() bool { return show(second.ID).Status == domain.StatusClosed }, "second ticket not closed")
	assert.Equal(t, pid, hostPidOf(t, sdir), "the new controller adopted the respawned host")

	prompts := promptLog(t, root.Dir)
	var p string
	for _, s := range prompts {
		if strings.Contains(s, "second-ticket-title") {
			p = s
		}
	}
	require.NotEmpty(t, p)
	before, after, ok := strings.Cut(p, host.SummaryEnd)
	require.True(t, ok, "prompt seeded with a summary: %q", p)
	assert.Contains(t, before, "first-ticket-title", "summary rebuilt from the journal")
	assert.Contains(t, after, "this session was resumed", "recovery note kept")
	assert.Equal(t, 1, strings.Count(strings.Join(prompts, "\n"), "this session was resumed"), "note once")
	got := sessionFieldsOf(t, admin, id)
	assert.False(t, got.SeedDue || got.RecoveryDue, "marks cleared once the prompt carried them")
}
