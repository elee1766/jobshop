package control_test

import (
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

// setPoolConfig rewrites the dev pool's model lines and reloads if asked.
func setPoolConfig(t *testing.T, root config.Root, lines string) {
	t.Helper()
	path := filepath.Join(root.Dir, config.FileName)
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	s := string(b)
	if i := strings.Index(s, "    # model\n"); i >= 0 {
		s = s[:i]
	}
	s += "    # model\n" + lines
	require.NoError(t, os.WriteFile(path, []byte(s), 0o600))
}

func configSets(t *testing.T, root config.Root) string {
	t.Helper()
	b, _ := os.ReadFile(filepath.Join(root.Dir, "configsets"))
	return string(b)
}

// TestSessionConfig (#304): options are set before the first turn, an
// unadvertised id is skipped, a reload applies between turns and never
// during one, and `session config` overrides and resets one session.
func TestSessionConfig(t *testing.T) {
	root, start := poolRootReal(t, "1m")
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "configopts"), nil, 0o600))
	setPoolConfig(t, root, "    model: m1\n    config_options: {other_id: x}\n")
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	show := func(id string) store.Ticket {
		var got store.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: id}, &struct {
			Ticket *store.Ticket `json:"ticket"`
		}{&got}))
		return got
	}
	info := func() rpc.SessionInfo {
		var infos []rpc.SessionInfo
		require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionList, rpc.SessionListParams{}, &infos))
		require.Len(t, infos, 1)
		return infos[0]
	}

	var task store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))
	passEventually(t, root.Dir, func() bool { return show(task.ID).Status == store.StatusClosed }, "condition not met after supervisor passes")
	assert.Equal(t, "model=m1 busy=false\n", configSets(t, root), "set once before the turn; other_id skipped")
	in := info()
	assert.Equal(t, "m1", in.Model)
	assert.Equal(t, "medium", in.Effort)
	sess, pid := in.ID, in.HostPid

	// A reload during a turn waits for it to end.
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "block"), nil, 0o600))
	var busy store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "busy", Pool: "dev"}, &busy))
	passEventually(t, root.Dir, func() bool {
		_, err := os.Stat(filepath.Join(root.Dir, "started"))
		return err == nil
	}, "condition not met after supervisor passes")
	setPoolConfig(t, root, "    model: m2\n    effort: high\n")
	require.NoError(t, admin.Call(t.Context(), rpc.MethodReload, struct{}{}, nil))
	pass(t, root.Dir, 3)
	assert.NotContains(t, configSets(t, root), "m2", "not applied mid-turn")
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))
	passEventually(t, root.Dir, func() bool { return show(busy.ID).Status == store.StatusClosed }, "condition not met after supervisor passes")
	// Model and effort are applied one after the other; wait for both.
	passEventually(t, root.Dir, func() bool { in := info(); return in.Model == "m2" && in.Effort == "high" }, "condition not met after supervisor passes")
	sets := configSets(t, root)
	assert.Contains(t, sets, "model=m2 busy=false")
	assert.Contains(t, sets, "reasoning_effort=high busy=false")
	assert.NotContains(t, sets, "busy=true")
	in = info()
	assert.Equal(t, "high", in.Effort)
	assert.Equal(t, pid, in.HostPid, "no respawn")

	// Override one session, then reset back to the pool's value.
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionConfig, rpc.SessionConfigParams{ID: sess, Options: map[string]string{"model": "openai-oauth:m3"}}, nil))
	passEventually(t, root.Dir, func() bool { return info().Model == "openai-oauth:m3" }, "condition not met after supervisor passes")
	assert.Equal(t, map[string]string{"model": "openai-oauth:m3"}, info().Override)
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionConfig, rpc.SessionConfigParams{ID: sess, Reset: true}, nil))
	passEventually(t, root.Dir, func() bool { return info().Model == "m2" }, "condition not met after supervisor passes")
	assert.Empty(t, info().Override)
	assert.Equal(t, store.StatusOpen, show(sess).Status)

	assert.Error(t, admin.Call(t.Context(), rpc.MethodSessionConfig, rpc.SessionConfigParams{ID: sess}, nil), "options or reset required")
	sc := dial(t, filepath.Join(root.SessionDir(sess), "cli.sock"))
	assert.ErrorIs(t, sc.Call(t.Context(), rpc.MethodSessionConfig, rpc.SessionConfigParams{ID: sess, Reset: true}, nil), rpc.ErrForbidden)
}

// TestSessionConfigRejected: an option the agent refuses ends the session
// config_failed before any turn runs.
func TestSessionConfigRejected(t *testing.T) {
	root, start := poolRoot(t, "1m")
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "configopts"), nil, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "configreject"), []byte("bad-route"), 0o600))
	setPoolConfig(t, root, "    model: bad-route\n")
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	// Finish the start-up tick before the ticket exists: under load it
	// could otherwise spawn the session, and the pass below a second one
	// once the first ended.
	pass(t, root.Dir, 1)

	var task store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))
	// One pass spawns the session; the hourly tick never comes (#550).
	// No more passes: one would spawn another session after this one ends.
	pass(t, root.Dir, 1)
	var sess store.Ticket
	require.Eventually(t, func() bool {
		var all []store.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession, Status: store.StatusClosed}, &all))
		if len(all) == 0 {
			return false
		}
		sess = all[0]
		return true
	}, dialWait, 50*time.Millisecond)
	assert.Equal(t, "config_failed", sess.Resolution)
	assert.Equal(t, 2, strings.Count(configSets(t, root), "model=bad-route"), "one retry")
	_, err := os.Stat(filepath.Join(root.Dir, "prompts"))
	assert.ErrorIs(t, err, os.ErrNotExist, "no turn on the wrong model")
}
