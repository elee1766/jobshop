package control_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control/controltest"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// A run whose last step closed done went needs_human and its escalation
// said "a step failed at its retry limit" (#1246). These tests cover the
// cause (a pack var added after the run started), the escalation text for
// a callback failure, and resume re-running the completion callback.

// lastStepHooks starts step a; completing a starts b from a prompt file.
const lastStepHooks = `def begin(run):
    return [start_step(id="a",pool="dev",prompt="first")]
def completed(run, step):
    if step.id == "a":
        return [start_step(id="b",pool="dev",prompt_file="next.md",needs=["a"])]
    return [finish_run(outcome="done")]
def handle(event, args):
    return []
`

// addPackDefault gives the test pack a var default, as a pack deploy that
// adds a var does (pool_prefix in #1239), and reloads.
func addPackDefault(t *testing.T, c *controltest.Controller, dir, name, value string) {
	t.Helper()
	path := filepath.Join(dir, "pack.yaml")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, []byte("defaults: {"+name+": "+value+"}\n"+string(data)), 0o600))
	c.Call(rpc.MethodReload, struct{}{}, nil)
}

// A prompt file that uses a pack var added after the run started renders
// with the live value: the run keeps going instead of escalating with
// `var "x" is not set` once its last step closes done.
func TestHookPromptUsesPackVarAddedAfterStart(t *testing.T) {
	t.Parallel()
	root := hookTestRoot(t, lastStepHooks)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "next.md"), []byte("pool {{.added}}dev"), 0o600))
	c := controltest.Start(t, root, control.Options{})
	id, first := startOpenRun(t, c)
	_, had := runFieldsOf(t, c, id).Inputs["added"]
	require.False(t, had, "the run must start without the var")

	addPackDefault(t, c, root.Dir, "added", "ops-")
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: first.ID}, nil)
	c.PassUntil(func() bool {
		return len(hookRun(t, c, id).Steps) == 2 || runFieldsOf(t, c, id).State != domain.RunRunning
	}, "completion callback did not run")
	require.Equal(t, domain.RunRunning, runFieldsOf(t, c, id).State, "comments: %v", commentsOf(t, c, id))
	assert.Equal(t, "pool ops-dev", hookRun(t, c, id).Steps[1].Body)
	assert.Empty(t, escalations(t, c.Admin, id))
}

// A completion callback that fails after every step closed done files an
// escalation naming the callback error, not a step at its retry limit.
// Resume after the fix runs the callback again and the run finishes.
func TestHookFailureEscalationNamesError(t *testing.T) {
	t.Parallel()
	const fixed = `def begin(run):
    return [start_step(id="a",pool="dev",prompt="first")]
def completed(run, step):
    return [finish_run(outcome="done")]
def handle(event, args):
    return []
`
	broken := strings.Replace(fixed, `    return [finish_run(outcome="done")]`, `    fail("post-merge setup broke")`, 1)
	root := hookTestRoot(t, broken)
	c := controltest.Start(t, root, control.Options{})
	id, first := startOpenRun(t, c)
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: first.ID}, nil)
	c.PassUntil(func() bool { return len(escalations(t, c.Admin, id)) == 1 }, "failed callback did not escalate")

	f := runFieldsOf(t, c, id)
	assert.Equal(t, domain.RunNeedsHuman, f.State)
	assert.Contains(t, f.HookError, "post-merge setup broke")
	esc := escalations(t, c.Admin, id)[0]
	assert.Contains(t, esc.Body, "needs a human: its workflow callback failed:")
	assert.Contains(t, esc.Body, "post-merge setup broke")
	assert.NotContains(t, esc.Body, "retry limit")

	// Resuming before the fix fails again: a new escalation round, with
	// the same error, replaces the first.
	c.Call(rpc.MethodRunResume, rpc.RunResumeParams{ID: id}, nil)
	c.PassUntil(func() bool { return len(escalations(t, c.Admin, id)) == 2 }, "second failure did not escalate again")
	assert.Equal(t, 2, runFieldsOf(t, c, id).Escalations)

	deploy(t, c, root, fixed)
	c.Call(rpc.MethodRunResume, rpc.RunResumeParams{ID: id}, nil)
	c.PassUntil(func() bool { return hookRun(t, c, id).Run.Status == domain.StatusClosed }, "resumed run did not finish")
	assert.Empty(t, runFieldsOf(t, c, id).HookError, "a resumed run drops the hook error")
	for _, e := range escalations(t, c.Admin, id) {
		assert.Equal(t, domain.StatusClosed, e.Status, "escalation %s settles with the run", e.ID)
	}
}

// A needs_human callback run whose steps all finished and that has no
// failed invocation to retry still resumes: the last step's completion
// callback runs again.
func TestHookResumeRerunsCompletionWhenStepsDone(t *testing.T) {
	t.Parallel()
	const waiting = `def begin(run):
    return [start_step(id="a",pool="dev",prompt="first")]
def completed(run, step):
    return []
def handle(event, args):
    return []
`
	root := hookTestRoot(t, waiting)
	c := controltest.Start(t, root, control.Options{})
	id, first := startOpenRun(t, c)
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: first.ID}, nil)
	c.PassUntil(func() bool { _, ok := runFieldsOf(t, c, id).Automation.Completed[first.ID]; return ok }, "completion not recorded")

	// Held for a human with nothing failed, as a lost invocation record
	// leaves it.
	run := hookRun(t, c, id).Run
	f := runFieldsOf(t, c, id)
	f.State = domain.RunNeedsHuman
	b, err := json.Marshal(f)
	require.NoError(t, err)
	c.Call(rpc.MethodTicketUpdate, rpc.UpdateParams{ID: id, Fields: b, IfRevision: run.Revision}, nil)

	deploy(t, c, root, strings.Replace(waiting, "def completed(run, step):\n    return []", "def completed(run, step):\n    return [finish_run(outcome=\"done\")]", 1))
	c.Call(rpc.MethodRunResume, rpc.RunResumeParams{ID: id}, nil)
	c.PassUntil(func() bool { return hookRun(t, c, id).Run.Status == domain.StatusClosed }, "resume did not re-run the completion callback")
	assert.Equal(t, 1, countContaining(commentsOf(t, c, id), "on_step_completed runs again for "+first.ID))
}
