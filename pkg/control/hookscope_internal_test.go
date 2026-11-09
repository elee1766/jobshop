package control

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/automation"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/hooks"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite"
)

// A journaled plan applies only while its callback's references still
// resolve inside the scope its run accepted. Moving a pool to another pack
// instance, same name and project, revokes it for an unapplied plan.
func TestHookPreparedPlanRespectsRevokedInstanceScope(t *testing.T) {
	t.Parallel()
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
	fields, err := domain.DecodeFields[domain.RunFields](run)
	require.NoError(t, err)
	in := hooks.Invocation{
		Run: run.ID, RunInput: &hooks.Run{ID: run.ID, Workflow: "work", Project: "js", State: domain.RunRunning},
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

// A run2 run that lost its accepted scope runs no callback code and applies
// nothing, even a finish with no references; its pin also names its instance.
func TestHookRunWithoutAcceptedScopeFailsClosed(t *testing.T) {
	t.Parallel()
	root := config.Root{Dir: t.TempDir()}
	write := func(name, body string) {
		t.Helper()
		path := filepath.Join(root.Dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
		require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	}
	pack := "workflows:\n  work:\n    project: js\n    on_start: hooks.star:begin\n    on_step_completed: hooks.star:completed\n"
	write("one/pack.yaml", pack)
	write("one/hooks.star", "def begin(run):\n    return [finish_run(outcome=\"done\")]\ndef completed(run, step):\n    return []\n")
	load := func(work string) *liveConfig {
		write("jobshop.yaml", "projects: {js: {}}\nagents: {fake: {command: \"false\"}}\n"+work)
		cfg, err := config.Load(root)
		require.NoError(t, err)
		lc, err := newLiveConfig(cfg, root)
		require.NoError(t, err)
		return lc
	}
	// Same script, selectors and roots, but no instance bounds its references.
	hostLevel := load("workflows:\n  work:\n    project: js\n    on_start: one/hooks.star:begin\n    on_step_completed: one/hooks.star:completed\n")
	accepted := load("packs: [{path: one}]\n")
	require.NotEqual(t, accepted.cfg.HookRunPins["work"], hostLevel.cfg.HookRunPins["work"], "moving a workflow out of its instance changes its run pin")

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
	fields, err := domain.DecodeFields[domain.RunFields](run)
	require.NoError(t, err)
	require.True(t, hooks.ScopedRunPin(fields.Automation.Digest))
	in := hooks.Invocation{
		Run: run.ID, RunInput: &hooks.Run{ID: run.ID, Workflow: "work", Project: "js", State: domain.RunRunning},
		Entry: fields.Automation.OnStart, Digest: fields.Automation.Digest,
		Actions: []automation.Action{{Kind: "finish_run", Outcome: "done"}},
	}
	in.ReadSet = map[string]int64{run.ID: run.Revision}
	planned, err := c.planHookActions(t.Context(), accepted, "plan", in, accepted.hooks[in.Entry], nil)
	require.NoError(t, err, "the recorded scope admits a finish")
	in.Plan = &planned

	fields.Automation.Scope = nil
	data, err := json.Marshal(fields)
	require.NoError(t, err)
	run, err = db.Update(t.Context(), "test", run.ID, store.Update{Fields: data, IfRevision: run.Revision})
	require.NoError(t, err)
	in.ReadSet = map[string]int64{run.ID: run.Revision}

	_, err = hookHost{c}.AdmitHook(t.Context(), accepted.runtime, in)
	assert.ErrorContains(t, err, "accepted reference scope", "execute")
	_, err = c.planHookActions(t.Context(), accepted, "plan", in, accepted.hooks[in.Entry], nil)
	assert.ErrorContains(t, err, "accepted reference scope", "commit")
	assert.ErrorContains(t, c.applyPlannedHook(t.Context(), "plan", in), "accepted reference scope", "prepared apply")
	_, err = c.resumeHook(t.Context(), "test", run, fields)
	assert.ErrorContains(t, err, "accepted reference scope", "resume")
	after, err := db.Get(t.Context(), run.ID)
	require.NoError(t, err)
	assert.NotEqual(t, domain.StatusClosed, after.Status)
}

// A callback-started static run is checked too: a retry pool that moved to
// another project since planning keeps the prepared graph from applying.
func TestHookPreparedStaticChildRejectsRetryProjectDrift(t *testing.T) {
	t.Parallel()
	root := config.Root{Dir: t.TempDir()}
	write := func(name, body string) {
		t.Helper()
		path := filepath.Join(root.Dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
		require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	}
	write("jobshop.yaml", "projects: {js: {}, other: {}}\ndefault_project: js\nagents: {fake: {command: \"false\"}}\npacks: [{path: one}]\n")
	pack := func(retryProject string) string {
		return "pools:\n  dev: {project: js, agent: fake, max_sessions: 0, dir: .}\n  retry: {project: " + retryProject + ", agent: fake, max_sessions: 0, dir: .}\n" +
			"workflows:\n  work:\n    project: js\n    on_start: hooks.star:begin\n    on_step_completed: hooks.star:completed\n" +
			"  child:\n    steps: [{id: work, pool: dev, retry_pool: retry, title: work, prompt: work}]\n"
	}
	write("one/pack.yaml", pack("js"))
	write("one/hooks.star", "def begin(run):\n    return [start_run(workflow=\"child\",key=\"child\")]\ndef completed(run, step):\n    return []\n")
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
	fields, err := domain.DecodeFields[domain.RunFields](run)
	require.NoError(t, err)
	in := hooks.Invocation{
		Run: run.ID, RunInput: &hooks.Run{ID: run.ID, Workflow: "work", Project: "js", State: domain.RunRunning},
		ReadSet: map[string]int64{run.ID: run.Revision}, Entry: fields.Automation.OnStart, Digest: fields.Automation.Digest,
		Actions: []automation.Action{{Kind: "start_run", Workflow: "child", Key: "child"}},
	}
	plans, err := c.prepareHookRuns(t.Context(), accepted, in)
	require.NoError(t, err)
	planned, err := c.planHookActions(t.Context(), accepted, "plan", in, accepted.hooks[in.Entry], plans)
	require.NoError(t, err)
	in.Plan = &planned

	write("one/pack.yaml", pack("other"))
	moved, err := config.Load(root)
	require.NoError(t, err)
	current, err := newLiveConfig(moved, root)
	require.NoError(t, err)
	c.live.Store(current)
	assert.ErrorContains(t, c.applyPlannedHook(t.Context(), "plan", in), "planned pool retry changed")
	_, err = store.OpenRun(t.Context(), db, "child", "child")
	assert.ErrorIs(t, err, domain.ErrNotFound, "no child run may be written")
}
