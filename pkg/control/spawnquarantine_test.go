package control_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// TestSpawnQuarantine points the pool at a missing repo, as on gasserver
// 2026-11-02 01:32-05:40 (#146): the first spawn quarantines the pool,
// files one escalation, and starts no second session until reload.
func TestSpawnQuarantine(t *testing.T) {
	root, start := poolRoot(t, "1m")
	missing := filepath.Join(root.Dir, "no-such-repo")
	cfg := fmt.Sprintf("projects:\n  js: {}\nagents:\n  fake:\n    command: %q\n    args: [fake-agent]\npools:\n  dev:\n    project: js\n    agent: fake\n    max_sessions: 1\n    idle_timeout: 1m\n    repo: %q\n", os.Args[0], missing)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.FileName), []byte(cfg), 0o600))
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	list := func(p rpc.ListParams) []store.Ticket {
		var out []store.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, p, &out))
		return out
	}
	status := func() rpc.PoolStatusResult {
		var st rpc.PoolStatusResult
		require.NoError(t, admin.Call(t.Context(), rpc.MethodPoolList, nil, &st))
		return st
	}

	var task store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))
	require.Eventually(t, func() bool { return status().Quarantined["dev"] != "" }, 20*time.Second, 50*time.Millisecond)
	assert.Contains(t, status().Quarantined["dev"], "no-such-repo")
	pass(t, root.Dir, 3)
	sessions := list(rpc.ListParams{Kind: store.KindSession})
	require.Len(t, sessions, 1, "one spawn, no retries")
	assert.Equal(t, "crashed", sessions[0].Resolution)

	esc := list(rpc.ListParams{Label: "needs-human"})
	require.Len(t, esc, 1)
	assert.True(t, strings.HasPrefix(esc[0].Title, "Pool dev quarantined: spawn: pool dev repo "), esc[0].Title)
	assert.Contains(t, string(esc[0].Meta), `"pool":{"name":"dev"`)

	// A second quarantine while the first escalation is open files none.
	require.NoError(t, admin.Call(t.Context(), rpc.MethodPoolUnquarantine, rpc.PoolParams{Pool: "dev"}, nil))
	require.Eventually(t, func() bool { return len(list(rpc.ListParams{Kind: store.KindSession})) == 2 }, 20*time.Second, 50*time.Millisecond)
	require.Eventually(t, func() bool { return status().Quarantined["dev"] != "" }, 20*time.Second, 50*time.Millisecond)
	assert.Len(t, list(rpc.ListParams{Label: "needs-human"}), 1, "escalation deduped per pool")

	// Reload clears it; the fixed config starts a session that does work.
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.FileName),
		fmt.Appendf(nil, "projects:\n  js: {}\nagents:\n  fake:\n    command: %q\n    args: [fake-agent]\npools:\n  dev:\n    project: js\n    agent: fake\n    max_sessions: 1\n    idle_timeout: 1m\n    dir: %q\n", os.Args[0], root.Dir), 0o600))
	require.NoError(t, admin.Call(t.Context(), rpc.MethodReload, struct{}{}, nil))
	assert.Empty(t, status().Quarantined)
	require.Eventually(t, func() bool {
		var got struct {
			Ticket store.Ticket `json:"ticket"`
		}
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: task.ID}, &got))
		return got.Ticket.Status == store.StatusClosed
	}, 20*time.Second, 50*time.Millisecond)
}

// TestSpawnQuarantineAgentMissing: an agent command that is not there is
// permanent too.
func TestSpawnQuarantineAgentMissing(t *testing.T) {
	root, start := poolRoot(t, "1m")
	cfg := fmt.Sprintf("projects:\n  js: {}\nagents:\n  fake:\n    command: no-such-agent-146\npools:\n  dev:\n    project: js\n    agent: fake\n    max_sessions: 1\n    idle_timeout: 1m\n    dir: %q\n", root.Dir)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.FileName), []byte(cfg), 0o600))
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, nil))
	var st rpc.PoolStatusResult
	require.Eventually(t, func() bool {
		require.NoError(t, admin.Call(t.Context(), rpc.MethodPoolList, nil, &st))
		return st.Quarantined["dev"] != ""
	}, 20*time.Second, 50*time.Millisecond)
	assert.Contains(t, st.Quarantined["dev"], "no-such-agent-146")
}
