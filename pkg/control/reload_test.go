package control_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	host "github.com/elee1766/acphost"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/hostspec"
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
	t.Cleanup(func() {
		cancel()
		// The test can end before serve finishes startup, and Run then
		// returns the canceled context (#61).
		if err := <-done; !errors.Is(err, context.Canceled) {
			require.NoError(t, err)
		}
	})
	admin := dial(t, root.AdminSocket())

	var res rpc.ConfigShowResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodConfigShow, struct{}{}, &res))
	dev := res.Config["pools"].(map[string]any)["dev"].(map[string]any)
	assert.Equal(t, "from-pack", dev["prompt"])
	assert.InDelta(t, 0, dev["max_sessions"], 0)
	assert.Equal(t, []string{filepath.Join(pack, config.PackFileName), root.Config()}, res.Sources["pools.dev"])
	assert.Equal(t, []string{root.Config()}, res.Sources["agents.fake"])

	// A workflow added to the pack shows up, with its source, after reload.
	require.NoError(t, os.WriteFile(filepath.Join(pack, config.PackFileName),
		[]byte("pools:\n  dev:\n    prompt: from-pack\n    max_sessions: 3\nworkflows:\n  w:\n    vars: [issue]\n    steps:\n      - {id: s, pool: dev, title: \"Do #${issue}\", prompt: p}\n"), 0o600))
	require.NoError(t, admin.Call(t.Context(), rpc.MethodReload, struct{}{}, nil))
	res = rpc.ConfigShowResult{}
	require.NoError(t, admin.Call(t.Context(), rpc.MethodConfigShow, struct{}{}, &res))
	w := res.Config["workflows"].(map[string]any)["w"].(map[string]any)
	steps := w["steps"].([]any)
	require.Len(t, steps, 1)
	assert.Equal(t, "Do #${issue}", steps[0].(map[string]any)["title"])
	assert.Equal(t, []string{filepath.Join(pack, config.PackFileName)}, res.Sources["workflows.w"])
}

// TestEnvRefreshLive changes the agent env while a session is live (as a
// self-deploy changes PATH): the idle host is respawned with the new env
// between turns, loads the conversation and keeps working in the same
// session. A turn running at the change finishes first.
func TestEnvRefreshLive(t *testing.T) {
	root, start := poolRoot(t, "1m")
	writeEnv := func(v string) {
		cfg := fmt.Sprintf("projects:\n  js: {}\nagents:\n  fake:\n    command: %q\n    args: [fake-agent]\n    env: {DEPLOY: %q}\npools:\n  dev:\n    project: js\n    agent: fake\n    max_sessions: 1\n    idle_timeout: 1m\n    dir: %q\n",
			os.Args[0], v, root.Dir)
		require.NoError(t, os.WriteFile(root.Config(), []byte(cfg), 0o600))
	}
	writeEnv("one")
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())

	var a store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "a", Pool: "dev"}, &a))
	waitClosed(t, admin, a.ID)
	sess := openSessions(t, admin)
	require.Len(t, sess, 1)
	sdir := root.SessionDir(sess[0].ID)
	pid1, err := os.ReadFile(filepath.Join(sdir, host.PidFile))
	require.NoError(t, err)
	hasEnv := func(kv string) bool {
		s, err := hostspec.Read(sdir)
		return err == nil && slices.Contains(s.Env, kv)
	}
	require.True(t, hasEnv("DEPLOY=one"))

	// Idle session: the env changes, the host is respawned before the next turn.
	writeEnv("two")
	require.NoError(t, admin.Call(t.Context(), rpc.MethodReload, nil, nil))
	require.Eventually(t, func() bool {
		pid, err := os.ReadFile(filepath.Join(sdir, host.PidFile))
		return err == nil && string(pid) != string(pid1) && host.Alive(sdir) && hasEnv("DEPLOY=two")
	}, 20*time.Second, 50*time.Millisecond, "idle host respawned with the new env")
	var b store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "b", Pool: "dev"}, &b))
	waitClosed(t, admin, b.ID) // a prompt to an unloaded conversation would fail
	now := openSessions(t, admin)
	require.Len(t, now, 1)
	assert.Equal(t, sess[0].ID, now[0].ID, "same session, not crashed")

	// Busy session: the claim and turn survive; the refresh waits for it.
	pid2, err := os.ReadFile(filepath.Join(sdir, host.PidFile))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "block"), nil, 0o600))
	var c store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "c", Pool: "dev"}, &c))
	require.Eventually(t, func() bool {
		_, err := os.Stat(filepath.Join(root.Dir, "started"))
		return err == nil
	}, 20*time.Second, 50*time.Millisecond)
	writeEnv("three")
	require.NoError(t, admin.Call(t.Context(), rpc.MethodReload, nil, nil))
	time.Sleep(300 * time.Millisecond)
	pid, err := os.ReadFile(filepath.Join(sdir, host.PidFile))
	require.NoError(t, err)
	assert.Equal(t, string(pid2), string(pid), "busy host not stopped mid-turn")
	var held store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: c.ID}, &struct {
		Ticket *store.Ticket `json:"ticket"`
	}{&held}))
	assert.Equal(t, sess[0].ID, held.Holder, "claim kept")
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))
	waitClosed(t, admin, c.ID)
	require.Eventually(t, func() bool { return hasEnv("DEPLOY=three") && host.Alive(sdir) }, 20*time.Second, 50*time.Millisecond)
	assert.Equal(t, sess[0].ID, openSessions(t, admin)[0].ID)
}

// TestEnvRefreshStopsDaemon: an agent's detached helper (like jcode's
// `serve` daemon, which runs the agent's tools) keeps the env it started
// with. Refreshing the host must stop it, or tools keep the old PATH
// after a deploy (#111). A new controller adopting the live host with a
// different env refreshes it before the next turn.
func TestEnvRefreshStopsDaemon(t *testing.T) {
	root, start := poolRoot(t, "1m")
	writeEnv := func(v string) {
		cfg := fmt.Sprintf("projects:\n  js: {}\nagents:\n  fake:\n    command: %q\n    args: [fake-agent]\n    env: {DEPLOY: %q}\npools:\n  dev:\n    project: js\n    agent: fake\n    max_sessions: 1\n    idle_timeout: 1m\n    dir: %q\n",
			os.Args[0], v, root.Dir)
		require.NoError(t, os.WriteFile(root.Config(), []byte(cfg), 0o600))
	}
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "daemon"), nil, 0o600))
	daemons := func() map[string]int {
		b, _ := os.ReadFile(filepath.Join(root.Dir, "daemons"))
		out := map[string]int{}
		for _, l := range strings.Split(strings.TrimSpace(string(b)), "\n") {
			var pid int
			var env string
			if _, err := fmt.Sscan(l, &pid, &env); err == nil {
				out[env] = pid
			}
		}
		return out
	}
	alive := func(pid int) bool {
		// A killed child of an exited parent may linger as a zombie.
		b, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
		return err == nil && !strings.Contains(string(b), ") Z ")
	}
	t.Cleanup(func() {
		for _, p := range daemons() {
			_ = syscall.Kill(p, syscall.SIGKILL)
		}
	})

	writeEnv("one")
	stop := start()
	admin := dial(t, root.AdminSocket())
	var a store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "a", Pool: "dev"}, &a))
	waitClosed(t, admin, a.ID)
	require.Eventually(t, func() bool { return daemons()["one"] != 0 }, 10*time.Second, 50*time.Millisecond)
	old := daemons()["one"]
	sess := openSessions(t, admin)
	require.Len(t, sess, 1)
	stop()

	// A deploy: the next controller spawns hosts with another env.
	writeEnv("two")
	stop = start()
	defer stop()
	admin = dial(t, root.AdminSocket())
	require.Eventually(t, func() bool { return daemons()["two"] != 0 }, 20*time.Second, 50*time.Millisecond, "agent respawned with the new env")
	require.Eventually(t, func() bool { return !alive(old) }, 10*time.Second, 50*time.Millisecond, "old daemon stopped")
	var b store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "b", Pool: "dev"}, &b))
	waitClosed(t, admin, b.ID)
	assert.Equal(t, sess[0].ID, openSessions(t, admin)[0].ID, "same session")
	assert.True(t, alive(daemons()["two"]), "new daemon left running")
}
