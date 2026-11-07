package control

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/automation"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/hooks"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite"
)

// A journaled plan applies only while its callback's references still
// resolve inside the scope its run accepted. Moving a pool to another pack
// instance, same name and project, revokes it for an unapplied plan.
func TestHookPreparedPlanRespectsRevokedInstanceScope(t *testing.T) {
	root := config.Root{Dir: t.TempDir()}
	write := func(name, body string) {
		t.Helper()
		path := filepath.Join(root.Dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
		require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	}
	write("jobshop.yaml", "projects: {js: {}}\nagents: {fake: {command: \"false\"}}\npacks: [{path: one}, {path: two}]\n")
	pool := "pools:\n  dev: {project: js, agent: fake, max_sessions: 0, dir: .}\n"
	workflow := "workflows:\n  work:\n    project: js\n    on_start: hooks.star:begin\n    on_step_completed: hooks.star:completed\n"
	write("one/pack.yaml", pool+workflow)
	write("two/pack.yaml", "pools: {}\n")
	write("one/hooks.star", "def begin(run):\n    return [start_step(id=\"work\",pool=\"dev\",prompt=\"work\")]\ndef completed(run, step):\n    return []\n")
	cfg, err := config.Load(root)
	require.NoError(t, err)
	accepted, err := newLiveConfig(cfg, root)
	require.NoError(t, err)
	db, err := sqlite.Open(t.Context(), root.DB(), sqlite.Options{})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	c := &Controller{store: db, root: root, live: new(atomic.Pointer[liveConfig])}
	c.live.Store(accepted)
	params := rpc.RunStartParams{Workflow: "work"}
	rp, err := c.planRun(t.Context(), accepted.cfg, scope{admin: true}, params)
	require.NoError(t, err)
	graph, err := c.buildRun(accepted.cfg, rp, params, nil, "")
	require.NoError(t, err)
	created, err := db.Apply(t.Context(), "test", graph)
	require.NoError(t, err)
	run := created[0]
	fields, err := store.DecodeFields[store.RunFields](run)
	require.NoError(t, err)
	in := hooks.Invocation{
		Run: run.ID, RunInput: &hooks.Run{ID: run.ID, Workflow: "work", Project: "js", State: store.RunRunning},
		ReadSet: map[string]int64{run.ID: run.Revision}, Entry: fields.Automation.OnStart, Digest: fields.Automation.Digest,
		Actions: []automation.Action{{Kind: "start_step", ID: "work", Pool: "dev", Prompt: "work"}},
	}
	planned, err := c.planHookActions(t.Context(), accepted, "plan", in, accepted.hooks[in.Entry], nil)
	require.NoError(t, err)
	in.Plan = &planned

	unscoped := *fields.Automation
	unscoped.Scope = nil
	assert.ErrorContains(t, checkPlannedTargets(accepted.cfg, in, &unscoped), "accepted reference scope", "a scoped pin without its scope fails closed")

	write("one/pack.yaml", workflow)
	write("two/pack.yaml", pool)
	moved, err := config.Load(root)
	require.NoError(t, err)
	current, err := newLiveConfig(moved, root)
	require.NoError(t, err)
	c.live.Store(current)
	require.Error(t, c.applyPlannedHook(t.Context(), "plan", in), "a prepared graph must not apply a now cross-instance step")
	children, err := store.Children(t.Context(), db, run.ID)
	require.NoError(t, err)
	assert.Empty(t, children)
}
