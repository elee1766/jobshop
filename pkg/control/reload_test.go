package control_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// writePoolConfig rewrites the root's config with one fake-agent pool.
func writePoolConfig(t *testing.T, root config.Root, maxSessions int, prompt string) {
	t.Helper()
	cfg := fmt.Sprintf("projects:\n  js: {}\nagents:\n  fake:\n    command: %q\n    args: [fake-agent]\npools:\n  dev:\n    project: js\n    agent: fake\n    max_sessions: %d\n    idle_timeout: 1m\n    dir: %q\n    prompt: %q\n",
		os.Args[0], maxSessions, root.Dir, prompt)
	require.NoError(t, os.WriteFile(root.Config(), []byte(cfg), 0o600))
}

func waitClosed(t *testing.T, admin *rpc.Client, id string) {
	t.Helper()
	require.Eventually(t, func() bool {
		var got store.Ticket
		err := admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: id}, &struct {
			Ticket *store.Ticket `json:"ticket"`
		}{&got})
		return err == nil && got.Status == store.StatusClosed
	}, 20*time.Second, 50*time.Millisecond)
}

func openSessions(t *testing.T, admin *rpc.Client) []store.Ticket {
	t.Helper()
	var out []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList,
		rpc.ListParams{Kind: store.KindSession, Status: store.StatusOpen}, &out))
	return out
}

// TestReloadPrompt changes the pool prompt and reloads: the running
// session's next turn uses the new prompt.
func TestReloadPrompt(t *testing.T) {
	root, start := poolRoot(t, "1m")
	writePoolConfig(t, root, 1, "OLD-PROMPT")
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())

	var a, b store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "a", Pool: "dev"}, &a))
	waitClosed(t, admin, a.ID)
	sess := openSessions(t, admin)
	require.Len(t, sess, 1)

	writePoolConfig(t, root, 1, "NEW-PROMPT")
	var res rpc.ReloadResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodReload, struct{}{}, &res))
	assert.Equal(t, 1, res.Pools)

	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "b", Pool: "dev"}, &b))
	waitClosed(t, admin, b.ID)
	prompts, err := os.ReadFile(filepath.Join(root.Dir, "prompts"))
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(prompts)), "\n")
	require.GreaterOrEqual(t, len(lines), 2)
	assert.Contains(t, lines[0], "OLD-PROMPT")
	last := lines[len(lines)-1]
	assert.Contains(t, last, "NEW-PROMPT")
	assert.Contains(t, last, b.ID)
	after := openSessions(t, admin)
	require.Len(t, after, 1)
	assert.Equal(t, sess[0].ID, after[0].ID, "same session, not restarted")
}

// TestReloadMaxSessions lowers max_sessions while two sessions work: no
// new session starts and the running ones keep going.
func TestReloadMaxSessions(t *testing.T) {
	root, start := poolRoot(t, "1m")
	writePoolConfig(t, root, 3, "")
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "block"), nil, 0o600))
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())

	var tickets [3]store.Ticket
	for i := range 2 {
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &tickets[i]))
	}
	require.Eventually(t, func() bool { return len(openSessions(t, admin)) == 2 }, 20*time.Second, 50*time.Millisecond)

	writePoolConfig(t, root, 1, "")
	require.NoError(t, admin.Call(t.Context(), rpc.MethodReload, struct{}{}, nil))
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &tickets[2]))
	time.Sleep(400 * time.Millisecond) // several supervisor ticks
	assert.Len(t, openSessions(t, admin), 2, "no new session over the lowered cap; running ones kept")

	// The running sessions finish their turns and drain the queue.
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))
	for _, tk := range tickets {
		waitClosed(t, admin, tk.ID)
	}
}

// TestReloadInvalid rejects a broken config, over the admin method and
// over the reload channel the serve command feeds from SIGHUP, and keeps
// the old config serving.
func TestReloadInvalid(t *testing.T) {
	reload := make(chan struct{})
	root, start := poolRootOpt(t, "1m", control.Options{Reload: reload})
	writePoolConfig(t, root, 1, "OLD-PROMPT")
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())

	require.NoError(t, os.WriteFile(root.Config(), []byte("pools:\n  dev:\n    max_sessions: [\n"), 0o600))
	err := admin.Call(t.Context(), rpc.MethodReload, struct{}{}, nil)
	require.Error(t, err)
	require.NoError(t, os.WriteFile(root.Config(), []byte("bogus_key: 1\n"), 0o600))
	reload <- struct{}{} // like SIGHUP; the error is only logged

	var a store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "a", Pool: "dev"}, &a))
	waitClosed(t, admin, a.ID)
	prompts, err := os.ReadFile(filepath.Join(root.Dir, "prompts"))
	require.NoError(t, err)
	assert.Contains(t, string(prompts), "OLD-PROMPT", "old config still active")

	// A valid file sent through the channel applies.
	writePoolConfig(t, root, 1, "NEW-PROMPT")
	reload <- struct{}{}
	// The loop takes the next send only after the first reload finished.
	reload <- struct{}{}
	var b store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "b", Pool: "dev"}, &b))
	waitClosed(t, admin, b.ID)
	prompts, err = os.ReadFile(filepath.Join(root.Dir, "prompts"))
	require.NoError(t, err)
	assert.Contains(t, string(prompts), "NEW-PROMPT")
}

// TestConfigShow reports the merged pool and which files set it.
func TestConfigShow(t *testing.T) {
	t.Parallel()
	dir, err := os.MkdirTemp("", "js") // short path for the unix socket
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	root := config.Root{Dir: dir}
	pack := filepath.Join(root.Dir, "pk")
	require.NoError(t, os.MkdirAll(pack, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(pack, config.PackFileName),
		[]byte("pools:\n  dev:\n    prompt: from-pack\n    max_sessions: 3\n"), 0o600))
	cfg := fmt.Sprintf("packs:\n  - path: pk\nprojects:\n  js: {}\nagents:\n  fake:\n    command: %q\npools:\n  dev:\n    project: js\n    agent: fake\n    max_sessions: 0\n    dir: %q\n",
		os.Args[0], root.Dir)
	require.NoError(t, os.WriteFile(root.Config(), []byte(cfg), 0o600))
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- control.Run(ctx, root, control.Options{Tick: time.Hour}) }()
	t.Cleanup(func() { cancel(); require.NoError(t, <-done) })
	admin := dial(t, root.AdminSocket())

	var res rpc.ConfigShowResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodConfigShow, struct{}{}, &res))
	dev := res.Config["pools"].(map[string]any)["dev"].(map[string]any)
	assert.Equal(t, "from-pack", dev["prompt"])
	assert.InDelta(t, 0, dev["max_sessions"], 0)
	assert.Equal(t, []string{filepath.Join(pack, config.PackFileName), root.Config()}, res.Sources["pools.dev"])
	assert.Equal(t, []string{root.Config()}, res.Sources["agents.fake"])
}
