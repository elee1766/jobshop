package stepexec

import (
	"context"
	"fmt"
	"maps"
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

// ScriptConfig is a script step's static settings, from its fields.
func ScriptConfig(sf store.StepFields) config.Step {
	s := config.Step{ID: sf.Step, Type: store.StepScript, Pool: sf.ScriptPool, Worktree: sf.Worktree, Timeout: sf.Timeout}
	// The script owns its retries through start_step; a restart reruns it.
	s.OnRestart = config.OnRestartRetry
	return s
}

// ScriptInputs is what the controller loads for a script attempt: the run
// and step envelopes and the program the run pinned.
type ScriptInputs struct {
	Run     *hooks.Run
	Step    *hooks.Step
	Program *hooks.Program
}

// script runs one attempt of a script step. inputs is called before the
// worktree is made, so a run whose code changed does not leave one.
func (e Executor) script(ctx context.Context, cfg *config.Config, t store.Ticket, s config.Step, sf store.StepFields, tpl Templates, attemptKey string, engine *hooks.Engine, inputs func(context.Context) (ScriptInputs, error)) Result {
	res := Result{Command: sf.Script}
	in, err := inputs(ctx)
	if err != nil {
		res.Result, res.Err = ResultFailed, err.Error()
		return res
	}
	pl, err := e.Place(cfg, s, sf, tpl)
	if err != nil {
		res.Result, res.Err = ResultFailed, err.Error()
		return res
	}
	key := hooks.StepAttemptKey(t.ID, t.Fence, e.Now())
	out, err := engine.StepAttempt(ctx, key, in.Program, sf.Script, s.TimeoutDur(), e.scriptExec(cfg, t, pl, attemptKey), in.Run, in.Step)
	res.Output = out.Logs
	if err != nil {
		res.Result, res.Err = ResultFailed, err.Error()
		return res
	}
	if out.Outcome != "" && !config.OutcomePattern.MatchString(out.Outcome) {
		res.Result, res.Err = ResultFailed, fmt.Sprintf("invalid outcome %q", out.Outcome)
		return res
	}
	if len(out.Actions) > 0 {
		res.Result, res.Err = ResultFailed, "a step function returns result(outcome=..., output=...), not actions"
		return res
	}
	if err := hooks.CheckSchema(sf.OutputSchema, out.Output, true); err != nil {
		res.Result, res.Err = ResultFailed, err.Error()
		return res
	}
	res.Result, res.Outcome, res.Structured = ResultOK, out.Outcome, out.Output
	return res
}

// scriptExec runs a step function's shell calls: in the pool's sandbox
// when it has one, with the step worktree (or pool dir) as the default
// cwd, and the same environment as a command step.
func (e Executor) scriptExec(cfg *config.Config, t store.Ticket, pl Place, attemptKey string) api.Exec {
	return func(ctx context.Context, req api.ExecRequest) (int, error) {
		dir := pl.Dir
		if req.Dir != "" {
			dir = req.Dir
			if !filepath.IsAbs(dir) {
				dir = filepath.Join(pl.Dir, dir)
			}
		}
		extra := append([]string{"JOBSHOP_STEP_KEY=" + attemptKey, "JOBSHOP_TICKET=" + t.ID}, e.Host.SlotEnv(ctx, t)...)
		extra = append(extra, e.Host.HostEnv()...)
		for _, k := range slices.Sorted(maps.Keys(req.Env)) {
			extra = append(extra, k+"="+req.Env[k])
		}
		ex, err := e.Host.Run(ctx, Cmd{Argv: req.Argv, Pool: pl.Pool, Dir: dir, Own: pl.Own, Env: extra, Stdout: req.Stdout, Stderr: req.Stderr, Cfg: cfg})
		if err != nil {
			return 0, err
		}
		if ex.Signal != "" {
			return 0, fmt.Errorf("killed by signal %s", ex.Signal)
		}
		return ex.ExitCode, nil
	}
}

// ApplyScript records the structured output of a finished attempt.
func ApplyScript(sf store.StepFields, res Result) store.StepFields {
	if sf.Type != store.StepScript {
		return sf
	}
	sf.StructuredOutput = nil
	if res.Result == ResultOK {
		sf.StructuredOutput = res.Structured
	}
	return sf
}
