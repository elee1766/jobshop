package control_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control/controltest"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// wakeRoot is hookTestRoot with a dev pool that may start one fake-agent
// session, so an agent step a callback creates gets a session.
func wakeRoot(t *testing.T, script string) config.Root {
	t.Helper()
	root := controltest.NewRoot(t, `pack: pack.yaml
projects: {js: {}}
agents:
  fake:
    command: `+os.Args[0]+`
    args: [fake-agent]
`)
	pack := `pools:
  dev: {project: js, agent: fake, max_sessions: 1, dir: "$ROOT"}
workflows:
  dynamic:
    project: js
    on_start: hooks.star:begin
    on_step_completed: hooks.star:completed
`
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "pack.yaml"), []byte(strings.ReplaceAll(pack, "$ROOT", root.Dir)), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "hooks.star"), []byte(script), 0o600))
	_, err := config.Load(root)
	require.NoError(t, err)
	return root
}

// TestCallbackStepsStartWithoutTicker (JOB-193): the ticker never fires
// and no test pass runs after the controller's first. A callback creates
// a command step, which runs and closes; the next callback creates an
// agent step in a pool with no session, which gets one and claims the
// step. Each step used to wait for the next 30s tick.
func TestCallbackStepsStartWithoutTicker(t *testing.T) {
	t.Parallel()
	root := wakeRoot(t, `def begin(run):
    return [start_step(id="gate", type="command", pool="dev", run="printf 'outcome: ready\\n'")]

def completed(run, step):
    if step.id == "gate":
        return [start_step(id="review", pool="dev", prompt="Review it", needs=["gate"])]
    return [finish_run(outcome="done")]
`)
	// The fake clock never advances, so no ticker or poll timer fires.
	c := controltest.Start(t, root, control.Options{})
	var started rpc.RunResult
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic"}, &started)

	require.Eventually(t, func() bool {
		return hookRun(t, c, started.Run.ID).Run.Status == domain.StatusClosed
	}, controltest.DialWait, 10*time.Millisecond, "callback steps did not run without a ticker")

	run := hookRun(t, c, started.Run.ID)
	assert.Equal(t, "done", run.Run.Resolution)
	require.Len(t, run.Steps, 2)
	for _, step := range run.Steps {
		assert.Equal(t, domain.StatusClosed, step.Status)
		assert.Equal(t, "done", step.Resolution)
	}

	// Loop check: with nothing due, passes apply no invocation, so no
	// pass asks for another.
	audit := func() int {
		var records []store.Record
		c.Call(rpc.MethodAutomationList, rpc.AutomationListParams{Run: started.Run.ID}, &records)
		return len(records)
	}
	assert.Equal(t, 3, audit(), "begin, gate completed, review completed")
	c.Pass(5)
	assert.Equal(t, 3, audit(), "passes with nothing due apply no invocation")
}

// TestIdleRunAppliesNoInvocationPerPass (JOB-193): a run waiting on an
// open step has nothing due, so passes (which kick the callback worker)
// apply no new invocation and so wake no further pass.
func TestIdleRunAppliesNoInvocationPerPass(t *testing.T) {
	t.Parallel()
	root := hookTestRoot(t, routingHooks)
	c := controltest.Start(t, root, control.Options{})
	var started rpc.RunResult
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic", Key: "idle", Inputs: map[string]any{"task": "x"}}, &started)
	c.PassUntil(func() bool { return len(hookRun(t, c, started.Run.ID).Steps) == 1 }, "start hook did not create work")
	records := func() []store.Record {
		var out []store.Record
		c.Call(rpc.MethodAutomationList, rpc.AutomationListParams{Run: started.Run.ID}, &out)
		return out
	}
	before := records()
	require.Len(t, before, 1)
	c.Pass(10)
	assert.Equal(t, before, records(), "no invocation is created or changed by passes")
}
