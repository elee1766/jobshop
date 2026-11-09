// Package stepexec runs the execution half of a controller-run step
// (command, http and script, JOB-157): templating, run dir placement and
// one attempt. It takes a resolved config.Step and a Host for what needs
// the controller (isolated commands, worktrees, slot env) and returns a
// Result. Claims, leases, the step record, escalation and stop-on-shutdown
// stay in pkg/orchestration/control, which this package never imports.
package stepexec

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/elee1766/acphost/pkg/proc"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
)

// ExitNotYet is the command exit code for "not yet, retry after
// retry_after" (EX_TEMPFAIL). It does not count as a failure.
const ExitNotYet = 75

// Step results.
const (
	ResultOK     = "ok"
	ResultNotYet = "not_yet"
	ResultFailed = "failed"
)

// Result is one attempt's outcome.
type Result struct {
	Result     string
	Output     string
	Outcome    string
	Command    string
	RetryAfter time.Duration
	Err        string
	// Structured is a script step's result(output=...).
	Structured map[string]any
}

// failed is a failed Result with err.
func failed(command string, err error) Result {
	return Result{Result: ResultFailed, Command: command, Err: err.Error()}
}

// Cmd is one command an attempt runs.
type Cmd struct {
	Argv []string
	// Pool picks the sandbox, user and cgroup the pool's sessions get;
	// empty runs the command as the controller.
	Pool string
	// Dir is the working dir and Own the run dir bound rw in a sandbox.
	Dir string
	Own string
	// Env is added to the environment (set inside a sandbox).
	Env    []string
	Stdout io.Writer
	Stderr io.Writer
	// Cfg is the config of the pass that started the attempt.
	Cfg *config.Config
}

// Host is what an attempt needs from the controller.
type Host interface {
	// Run runs c, isolated as c.Pool says. A non-zero exit is not an error.
	Run(ctx context.Context, c Cmd) (proc.Exit, error)
	// Worktree makes a locked, detached worktree of repo at ref in dir
	// unless dir exists. what names the owner in errors.
	Worktree(repo, ref, dir, reason, what string) error
	// SlotEnv is the env of the slots t's claim holds.
	SlotEnv(ctx context.Context, t domain.Ticket) []string
	// HostEnv is JOBSHOP_BIN and the PATH that has the controller's binary
	// first, as for sessions (#669). Empty when the binary is unknown.
	HostEnv() []string
}

// Executor runs attempts.
type Executor struct {
	// Root is the controller's root dir: the default working dir and the
	// parent of runs/.
	Root string
	Host Host
	// HTTP is the traced client http steps use.
	HTTP *http.Client
	Now  func() time.Time
}
