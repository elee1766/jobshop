package stepexec

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/elee1766/acphost/pkg/proc"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
)

// withoutPath drops PATH from env: the step's own PATH comes from HostEnv.
func withoutPath(env []string) []string {
	out := env[:0:0]
	for _, kv := range env {
		if !strings.HasPrefix(kv, "PATH=") {
			out = append(out, kv)
		}
	}
	return out
}

// CommandConfig is a callback command step's settings, from the spec its
// fields recorded. ok is false for a workflow command step, whose
// settings are in the workflow.
func CommandConfig(sf domain.StepFields) (config.Step, bool) {
	spec := sf.CommandSpec
	if sf.Type != config.StepCommand || spec == nil {
		return config.Step{}, false
	}
	return config.Step{
		ID: sf.Step, Type: config.StepCommand, Pool: spec.Pool, Run: slices.Clone(spec.Argv),
		Timeout: spec.Timeout, RetryAfter: spec.RetryAfter, OnRestart: spec.OnRestart, Worktree: spec.Worktree,
	}, true
}

// command runs one attempt of a command step.
func (e Executor) command(ctx context.Context, cfg *config.Config, t domain.Ticket, s config.Step, sf domain.StepFields, tpl Templates, attemptKey string) Result {
	argv, err := s.Argv()
	if err != nil {
		return failed("", err)
	}
	// A callback's arguments are values, not Go templates.
	if sf.CommandSpec == nil {
		for i, a := range argv {
			if argv[i], err = tpl.Fill(a); err != nil {
				return failed("", err)
			}
		}
	}
	extra := append(Env(tpl.Vars, attemptKey, t), e.Host.SlotEnv(ctx, t)...)
	// The controller's own binary comes first on PATH and is JOBSHOP_BIN,
	// as for sessions: units run with a PATH that has no jobshop (#669).
	extra = append(extra, e.Host.HostEnv()...)
	env := append(withoutPath(os.Environ()), extra...)
	red := cfg.Redact.Scanner().Redactor(env)
	res := Result{Command: red.Text(strings.Join(argv, " "))}
	if spec := sf.CommandSpec; spec != nil {
		// The pool was checked when the batch was accepted; a reload since
		// may have removed it or moved it to another project.
		if pool, ok := cfg.Pools[spec.Pool]; !ok || pool.Project != t.Project {
			res.Result, res.Err = ResultFailed, fmt.Sprintf("command pool %s is missing or outside step project %s", spec.Pool, t.Project)
			return res
		}
	}
	pl, err := e.Place(cfg, s, sf, tpl)
	if err != nil {
		res.Result, res.Err = ResultFailed, err.Error()
		return res
	}
	var out, errOut LimitedBuffer
	cmd := Cmd{Argv: argv, Pool: pl.Pool, Dir: pl.Dir, Own: pl.Own, Env: extra, Stdout: &out, Stderr: &errOut, Cfg: cfg}
	ex, runErr := e.Host.Run(ctx, cmd)
	res.Output = red.Text(out.String())
	res.Outcome = OutcomeOf(res.Output)
	switch {
	case ctx.Err() != nil:
		res.Result, res.Err = ResultFailed, "timed out"
	case runErr == nil && ex.ExitCode == 0 && ex.Signal == "":
		res.Result = ResultOK
	case runErr == nil && ex.Signal == "" && ex.ExitCode == ExitNotYet:
		res.Result = ResultNotYet
	default:
		res.Result = ResultFailed
		if runErr != nil {
			res.Err = runErr.Error()
		} else {
			res.Err = "exit " + ExitText(ex)
		}
		if tail := strings.TrimSpace(red.Text(errOut.String())); tail != "" {
			res.Err += ": " + LastLines(tail, 5)
		}
	}
	return res
}

// ExitText is the exit code, or "signal S", for a command's exit.
func ExitText(ex proc.Exit) string {
	if ex.Signal != "" {
		return "signal " + ex.Signal
	}
	return strconv.Itoa(ex.ExitCode)
}
