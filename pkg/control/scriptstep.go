package control

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/elee1766/jobshop/pkg/automation/api"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/hooks"
	"github.com/elee1766/jobshop/pkg/store"
)

// Script steps (JOB-124): a Starlark step function the controller runs
// in-process as a durable step, through the same claim, lease, retry and
// restart path as command steps. Only shell leaves the controller, and it
// runs in the step's pool sandbox and worktree.

// scriptStep is a script step's static settings, from its fields.
func scriptStep(sf store.StepFields) config.Step {
	s := config.Step{ID: sf.Step, Type: store.StepScript, Pool: sf.ScriptPool, Worktree: sf.Worktree, Timeout: sf.Timeout}
	// The script owns its retries through start_step; a restart reruns it.
	s.OnRestart = config.OnRestartRetry
	return s
}

// runScriptStep runs one attempt of a script step.
func (c *Controller) runScriptStep(ctx context.Context, cfg *config.Config, t store.Ticket, s config.Step, sf store.StepFields, st stepTemplates, attemptKey string) autoResult {
	res := autoResult{command: sf.Script}
	run, step, program, err := c.scriptInputs(ctx, t, sf)
	if err != nil {
		res.result, res.err = resultFailed, err.Error()
		return res
	}
	pl, err := c.stepPlace(cfg, s, sf, st)
	if err != nil {
		res.result, res.err = resultFailed, err.Error()
		return res
	}
	exec := c.scriptExec(cfg, t, pl, attemptKey)
	key := hooks.StepAttemptKey(t.ID, t.Fence, c.now())
	out, err := c.hooks().StepAttempt(ctx, key, program, sf.Script, s.TimeoutDur(), exec, run, step)
	res.output = out.Logs
	if err != nil {
		res.result, res.err = resultFailed, err.Error()
		return res
	}
	if out.Outcome != "" && !config.OutcomePattern.MatchString(out.Outcome) {
		res.result, res.err = resultFailed, fmt.Sprintf("invalid outcome %q", out.Outcome)
		return res
	}
	if len(out.Actions) > 0 {
		res.result, res.err = resultFailed, "a step function returns result(outcome=..., output=...), not actions"
		return res
	}
	if err := checkHookSchema(sf.OutputSchema, out.Output, true); err != nil {
		res.result, res.err = resultFailed, err.Error()
		return res
	}
	res.result, res.outcome, res.structured = resultOK, out.Outcome, out.Output
	return res
}

// scriptInputs loads the run and step envelopes and the program the run
// pinned. A run whose callback code changed since it started fails the
// attempt rather than running new code against old state.
func (c *Controller) scriptInputs(ctx context.Context, t store.Ticket, sf store.StepFields) (*hooks.Run, *hooks.Step, *hooks.Program, error) {
	runTicket, err := c.store.Get(ctx, sf.Run)
	if err != nil {
		return nil, nil, nil, err
	}
	rf, err := store.DecodeFields[store.RunFields](runTicket)
	if err != nil {
		return nil, nil, nil, err
	}
	if rf.Automation == nil {
		return nil, nil, nil, errors.New("script step outside a callback workflow")
	}
	runtime := c.live.Load().runtime
	if !hooks.RunPinned(runtime.Programs, runtime.Config, rf.Workflow, rf.Automation.Digest) {
		return nil, nil, nil, errors.New("workflow callback code changed since the run started: restore it or cancel the run")
	}
	program := runtime.Programs[sf.ScriptHook]
	if program == nil {
		return nil, nil, nil, fmt.Errorf("callback %s is not loaded", sf.ScriptHook)
	}
	children, err := store.Children(ctx, c.store, sf.Run)
	if err != nil {
		return nil, nil, nil, err
	}
	run := &hooks.Run{ID: runTicket.ID, Workflow: rf.Workflow, Key: rf.Key, Project: runTicket.Project, Inputs: rf.Inputs, Vars: rf.Vars, State: rf.State, Steps: []hooks.Step{}}
	var step hooks.Step
	for _, child := range children {
		if child.Kind != store.KindStep {
			continue
		}
		in, err := stepHookInput(child)
		if err != nil {
			return nil, nil, nil, err
		}
		run.Steps = append(run.Steps, in)
		if child.ID == t.ID {
			step = in
		}
	}
	return run, &step, program, nil
}

// scriptExec runs a step function's shell calls: in the pool's sandbox
// when it has one, with the step worktree (or pool dir) as the default
// cwd, and the same environment as a command step.
func (c *Controller) scriptExec(cfg *config.Config, t store.Ticket, pl stepPlace, attemptKey string) api.Exec {
	return func(ctx context.Context, req api.ExecRequest) (int, error) {
		dir := pl.dir
		if req.Dir != "" {
			dir = req.Dir
			if !filepath.IsAbs(dir) {
				dir = filepath.Join(pl.dir, dir)
			}
		}
		extra := append([]string{"JOBSHOP_STEP_KEY=" + attemptKey, "JOBSHOP_TICKET=" + t.ID}, c.slotEnv(ctx, t)...)
		if exe, err := c.hostExe(); err == nil {
			extra = append(extra, "JOBSHOP_BIN="+exe,
				"PATH="+filepath.Dir(exe)+string(os.PathListSeparator)+os.Getenv("PATH"))
		}
		for _, k := range slices.Sorted(maps.Keys(req.Env)) {
			extra = append(extra, k+"="+req.Env[k])
		}
		ic := isolatedCmd{argv: req.Argv, dir: dir, own: pl.own, env: extra, name: "step", cfg: cfg, stdout: req.Stdout, stderr: req.Stderr}
		if c.isolates(cfg, pl.pool) {
			ic.pool = pl.pool
		}
		ex, err := c.runIsolated(ctx, ic)
		if err != nil {
			return 0, err
		}
		if ex.Signal != "" {
			return 0, fmt.Errorf("killed by signal %s", ex.Signal)
		}
		return ex.ExitCode, nil
	}
}

// scriptStepFields records the structured output of a finished attempt.
func scriptStepFields(sf store.StepFields, res autoResult) store.StepFields {
	if sf.Type != store.StepScript {
		return sf
	}
	sf.StructuredOutput = nil
	if res.result == resultOK {
		sf.StructuredOutput = res.structured
	}
	return sf
}
