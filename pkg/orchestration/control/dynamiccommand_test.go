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

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control/controltest"
	packpkg "github.com/gfx-labs/jobshop/pkg/pack"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// Command steps a callback starts (start_step(type="command")): the step
// records its argv and settings, so these runs have no workflow step to
// read them from.

// TestDynamicCommandStepRetryAndRestart: exit 75 defers the step without a
// failure, and a restarted controller runs the recorded command again.
// Arguments reach the command as written, JOBSHOP_VAR_* comes from the
// run's inputs, and both commands hold the one emu slot in turn.
func TestDynamicCommandStepRetryAndRestart(t *testing.T) {
	root := fieldsRoot(t, `def begin(run):
    return [start_step(id="gate", type="command", pool="dev", needs_resources=["emu"], run=["sh", "-c", "if [ ! -f attempted ]; then touch attempted; exit 75; fi\nprintf '%s\\n' \"$1\" \"$2\" \"$3\" \"$JOBSHOP_VAR_TASK\"\nprintf 'outcome: ready\\n'", "gate", "with spaces", "", "{{literal}}"], timeout="1m", retry_after="2m", on_restart="retry", retries=1)]
def completed(run, step):
    if step.id == "gate":
        if step.outcome != "ready" or step.pool != "dev":
            fail("command result or pool lost")
        return [start_step(id="finish", type="command", pool="dev", needs_resources=["emu"], run="printf 'outcome: done\\n'", needs=["gate"])]
    return [finish_run(outcome=step.outcome)]
`)
	c := controltest.Start(t, root, control.Options{})
	var started rpc.RunResult
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic", Inputs: map[string]any{"task": "accepted input"}}, &started)
	var first domain.StepFields
	c.PassUntil(func() bool {
		run := hookRun(t, c, started.Run.ID)
		if len(run.Steps) != 1 {
			return false
		}
		var err error
		first, err = domain.DecodeFields[domain.StepFields](run.Steps[0])
		require.NoError(t, err)
		assert.Empty(t, run.Steps[0].Pool, "native commands must not be claimed by agents")
		return first.Result == "not_yet" && run.Steps[0].Holder == "" && !run.Steps[0].DeferUntil.IsZero()
	}, "command did not defer")
	require.NotNil(t, first.CommandSpec)
	assert.Equal(t, "dev", first.CommandSpec.Pool)
	assert.Equal(t, "2m", first.CommandSpec.RetryAfter)
	assert.Equal(t, []string{"emu"}, first.Resources)
	assert.Zero(t, first.Failures)
	_, err := os.Stat(filepath.Join(root.Dir, "attempted"))
	require.NoError(t, err)
	clock := c.Clock
	c.Stop()
	clock.Advance(3 * time.Minute)
	c = controltest.Start(t, root, control.Options{Clock: clock})
	c.PassUntil(func() bool {
		return hookRun(t, c, started.Run.ID).Run.Status == domain.StatusClosed
	}, "persisted command did not resume and finish the dynamic workflow")
	run := hookRun(t, c, started.Run.ID)
	require.Len(t, run.Steps, 2)
	assert.Equal(t, "done", run.Run.Resolution)
	for _, ticket := range run.Steps {
		sf, err := domain.DecodeFields[domain.StepFields](ticket)
		require.NoError(t, err)
		assert.Empty(t, ticket.Pool)
		assert.Equal(t, domain.StatusClosed, ticket.Status)
		if sf.Step == "gate" {
			assert.Equal(t, "with spaces\n\n{{literal}}\naccepted input\noutcome: ready\n", sf.Output)
			assert.Zero(t, sf.Failures)
		}
	}
}

// TestDynamicCommandInterruptedAttempt stops the controller during a
// callback command: on_restart retry reruns it under the same attempt key
// with no failure, fail counts one.
func TestDynamicCommandInterruptedAttempt(t *testing.T) {
	for _, mode := range []string{"retry", "fail"} {
		t.Run(mode, func(t *testing.T) {
			root := hookTestRoot(t, fmt.Sprintf(`def begin(run):
    return start_step(id="gate", type="command", pool="dev", on_restart=%q, retries=2, run="echo \"$JOBSHOP_STEP_KEY\" >> keys; if [ ! -f go ]; then mkfifo wait; read line < wait; fi")
def completed(run, step):
    return finish_run(outcome="done")
def handle(event, args):
    return []
`, mode))
			c := controltest.Start(t, root, control.Options{})
			var started rpc.RunResult
			c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic"}, &started)
			c.PassUntil(func() bool { return len(keyLines(t, root)) == 1 }, "command did not start")
			clock := c.Clock
			c.Stop()
			require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))
			c = controltest.Start(t, root, control.Options{Clock: clock})
			c.PassUntil(func() bool {
				return hookRun(t, c, started.Run.ID).Run.Status == domain.StatusClosed
			}, "interrupted command did not resume")
			keys := keyLines(t, root)
			require.Len(t, keys, 2)
			step := hookRun(t, c, started.Run.ID).Steps[0]
			sf, err := domain.DecodeFields[domain.StepFields](step)
			require.NoError(t, err)
			if mode == "retry" {
				assert.Equal(t, keys[0], keys[1])
				assert.Zero(t, sf.Failures)
			} else {
				assert.NotEqual(t, keys[0], keys[1])
				assert.Equal(t, 1, sf.Failures)
			}
		})
	}
}

// TestDynamicCommandSandboxWorktree: worktree=True runs the command in the
// run's worktree inside the pool's sandbox, not in the pool's repo.
func TestDynamicCommandSandboxWorktree(t *testing.T) {
	bwrapWorks(t)
	root, repo := worktreeRoot(t, true)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "pk", "hooks.star"), []byte(`def begin(run):
    return start_step(id="check", type="command", pool="dev", worktree=True, retries=0, run=["sh", "-c", "echo owned > note; pwd; test ! -e ../../../jobshop.db || echo db-visible; printf 'outcome: pass\\n'"])
def completed(run, step):
    return finish_run(outcome=step.outcome)
`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "pk", packpkg.FileName), []byte(`workflows:
  dynamic:
    project: js
    on_start: hooks.star:begin
    on_step_completed: hooks.star:completed
`), 0o600))
	c := controltest.Start(t, root, control.Options{})
	var started rpc.RunResult
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic"}, &started)
	c.PassUntil(func() bool {
		return hookRun(t, c, started.Run.ID).Run.Status == domain.StatusClosed
	}, "sandboxed callback command did not finish")
	run := hookRun(t, c, started.Run.ID)
	require.Len(t, run.Steps, 1)
	sf, err := domain.DecodeFields[domain.StepFields](run.Steps[0])
	require.NoError(t, err)
	assert.Equal(t, "pass", run.Run.Resolution)
	assert.True(t, strings.HasPrefix(sf.Output, filepath.Join(root.Dir, "runs", run.Run.ID, "work")+"\n"), sf.Output)
	assert.NotContains(t, sf.Output, "db-visible")
	_, err = os.Stat(filepath.Join(repo, "note"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

// TestDynamicCommandDefinitionRequiresAdmin: a session cannot change what
// a command step runs, or turn another step into one, through a fields
// update. It can still edit the step's data, and the admin socket can
// replace the definition.
func TestDynamicCommandDefinitionRequiresAdmin(t *testing.T) {
	root := hookTestRoot(t, `def begin(run):
    return [start_step(id="approval", pool="dev", prompt="Approve command"), start_step(id="command", type="command", pool="dev", run=["true"], needs=["approval"])]
def completed(run, step):
    return []
def handle(event, args):
    return []
`)
	c := controltest.Start(t, root, control.Options{})
	var started rpc.RunResult
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic"}, &started)
	c.PassUntil(func() bool { return len(hookRun(t, c, started.Run.ID).Steps) == 2 }, "command plan not accepted")
	steps := hookRun(t, c, started.Run.ID).Steps
	var command, approval domain.Ticket
	for _, step := range steps {
		f, err := domain.DecodeFields[domain.StepFields](step)
		require.NoError(t, err)
		if f.CommandSpec != nil {
			command = step
		} else {
			approval = step
		}
	}
	require.NotEmpty(t, command.ID)
	var session rpc.SessionNewResult
	c.Call(rpc.MethodSessionNew, rpc.SessionNewParams{Agent: "external", Pool: "dev"}, &session)
	client := c.Dial(session.Socket)
	for _, mutate := range []func(*domain.StepFields){
		func(f *domain.StepFields) { f.CommandSpec.Argv = []string{"false"} },
		func(f *domain.StepFields) { f.CommandSpec.Pool = "another" },
		func(f *domain.StepFields) { f.CommandSpec = nil },
		func(f *domain.StepFields) { f.Type = "" },
	} {
		f, err := domain.DecodeFields[domain.StepFields](command)
		require.NoError(t, err)
		mutate(&f)
		fields, err := json.Marshal(f)
		require.NoError(t, err)
		err = client.Call(t.Context(), rpc.MethodTicketUpdate, rpc.UpdateParams{ID: command.ID, Fields: fields}, nil)
		require.ErrorIs(t, err, rpc.ErrForbidden)
	}
	f, err := domain.DecodeFields[domain.StepFields](approval)
	require.NoError(t, err)
	f.Type, f.CommandSpec = config.StepCommand, &domain.CommandSpec{Pool: "dev", Argv: []string{"true"}}
	fields, err := json.Marshal(f)
	require.NoError(t, err)
	require.ErrorIs(t, client.Call(t.Context(), rpc.MethodTicketUpdate, rpc.UpdateParams{ID: approval.ID, Fields: fields}, nil), rpc.ErrForbidden)
	// Data-only updates retain the accepted execution definition.
	f, err = domain.DecodeFields[domain.StepFields](command)
	require.NoError(t, err)
	f.Inputs = map[string]any{"note": "reviewed"}
	fields, err = json.Marshal(f)
	require.NoError(t, err)
	require.NoError(t, client.Call(t.Context(), rpc.MethodTicketUpdate, rpc.UpdateParams{ID: command.ID, Fields: fields}, nil))
	f.CommandSpec.Argv = []string{"false"}
	fields, err = json.Marshal(f)
	require.NoError(t, err)
	c.Call(rpc.MethodTicketUpdate, rpc.UpdateParams{ID: command.ID, Fields: fields}, nil)
}
