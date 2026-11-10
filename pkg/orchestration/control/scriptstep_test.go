package control_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	packpkg "github.com/gfx-labs/jobshop/pkg/pack"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/support/clock"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

const scriptStepHooks = `def begin(run):
    return start_step(id = "check", pool = "dev", script = check, worktree = True, timeout = "1m",
                      retries = 0, inputs = {"issue": run.inputs["issue"]})

def check(run, step):
    w = shell(["sh", "-c", "echo $0 > note; pwd; test ! -e ../../../jobshop.db || echo db-visible", step.inputs["issue"]])
    if w.exit_code != 0:
        fail("write failed: " + w.stderr)
    r = shell(["cat", "note"])
    return result(outcome = "pass", output = {"note": r.stdout.strip(), "cwd": w.stdout.strip()})

def completed(run, step):
    return finish_run(outcome = step.outcome, output = step.output)
`

// TestScriptStepSandboxWorktree (JOB-124): a callback starts a Starlark
// step function; the controller runs it in-process, its shell calls run
// in the pool sandbox inside the run worktree, its outcome and output
// route the run, and each call is in the automation journal.
func TestScriptStepSandboxWorktree(t *testing.T) {
	bwrapWorks(t)
	t.Parallel()
	root, repo := worktreeRoot(t, true)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "pk", "hooks.star"), []byte(scriptStepHooks), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "pk", packpkg.FileName), []byte(`workflows:
  scripted:
    project: js
    vars: [issue]
    on_start: hooks.star:begin
    on_step_completed: hooks.star:completed
`), 0o600))
	cfg, err := os.ReadFile(root.Config())
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(root.Config(), append(cfg, []byte("automation: {shell: true}\n")...), 0o600))

	admin, _ := startAuto(t, root, clock.NewFake(time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)))
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "scripted", Vars: map[string]string{"issue": "42"}})
	passUntilWall(t, root, func() bool {
		return showRun(t, admin, r.Run.ID).Run.Status == domain.StatusClosed
	}, "script step did not finish the run")

	res := showRun(t, admin, r.Run.ID)
	require.Len(t, res.Steps, 1)
	step := res.Steps[0]
	sf := stepFields(t, step)
	assert.Empty(t, step.Pool, "no session may claim a script step")
	assert.Equal(t, domain.StepScript, sf.Type)
	assert.Equal(t, "hooks.star:check", sf.Script)
	require.Equal(t, "ok", sf.Result, sf.Output)
	assert.Equal(t, "pass", sf.Outcome)
	assert.Equal(t, "42", sf.StructuredOutput["note"], "second shell call sees the first one's file")
	assert.Equal(t, filepath.Join(root.Dir, "runs", r.Run.ID, "work"), sf.StructuredOutput["cwd"])
	_, err = os.Stat(filepath.Join(repo, "note"))
	assert.True(t, os.IsNotExist(err), "main checkout untouched")

	rf, err := domain.DecodeFields[domain.RunFields](res.Run)
	require.NoError(t, err)
	assert.Equal(t, domain.RunDone, rf.State)
	assert.Equal(t, "42", rf.Automation.Output["note"])

	var audit []store.Record
	require.NoError(t, admin.Call(t.Context(), rpc.MethodAutomationList, rpc.AutomationListParams{Run: r.Run.ID}, &audit))
	var stepKey string
	for _, rec := range audit {
		if rec.State == "step_done" {
			stepKey = rec.Key
		}
	}
	require.NotEmpty(t, stepKey, "step attempt not journaled: %v", audit)
	var trace rpc.AutomationResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodAutomationShow, rpc.IDParams{ID: stepKey}, &trace))
	require.Len(t, trace.Calls, 2)
	var call map[string]any
	require.NoError(t, json.Unmarshal(trace.Calls[0].Value, &call))
	observation, _ := call["observation"].(map[string]any)
	assert.Equal(t, "shell", observation["capability"])
	assert.Equal(t, "completed", trace.Calls[0].State)
	assert.False(t, strings.Contains(fmt.Sprint(sf.StructuredOutput), "db-visible"), "jobshop.db hidden by the sandbox")
}
