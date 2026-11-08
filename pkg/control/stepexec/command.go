package stepexec

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/elee1766/acphost/pkg/proc"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/store"
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

// command runs one attempt of a command step.
func (e Executor) command(ctx context.Context, cfg *config.Config, t store.Ticket, s config.Step, sf store.StepFields, tpl Templates, attemptKey string) Result {
	argv, err := s.Argv()
	if err != nil {
		return failed("", err)
	}
	for i, a := range argv {
		if argv[i], err = tpl.Fill(a); err != nil {
			return failed("", err)
		}
	}
	extra := append(Env(tpl.Vars, attemptKey, t), e.Host.SlotEnv(ctx, t)...)
	// The controller's own binary comes first on PATH and is JOBSHOP_BIN,
	// as for sessions: units run with a PATH that has no jobshop (#669).
	extra = append(extra, e.Host.HostEnv()...)
	env := append(withoutPath(os.Environ()), extra...)
	red := cfg.Redact.Scanner().Redactor(env)
	res := Result{Command: red.Text(strings.Join(argv, " "))}
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
