package control

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/elee1766/acphost/pkg/proc"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/execution/sandbox"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// isolatedCmd is a command the controller runs for a hook, prompt rule,
// Desk exec, trigger check or command step (#810). pool picks the
// sandbox, mapped user and cgroup limits sessions of that pool get;
// empty runs it unisolated (still its own process group, killed with
// ctx and with the controller).
type isolatedCmd struct {
	argv []string
	pool string
	// dir is the working dir. With a sandbox it must be inside own or be
	// bound rw by the pool; empty with a sandbox is a fresh scratch dir.
	dir string
	// own is bound rw in the sandbox (a run dir). Empty: a scratch dir
	// under the root, removed after.
	own    string
	env    []string // added to the environment (set inside a sandbox)
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
	name   string // cgroup name prefix
	// cfg is the config the pool's isolation comes from, required with
	// pool: the pass's in pass code (JOB-104), else c.config().
	cfg *config.Config
}

// errSetup is a sandbox or isolation failure: the command never ran.
var errSetup = sandbox.ErrSetup

// runIsolated runs ic and returns its exit. err is a start failure, a
// setup failure (wrapping errSetup) or a wait error; a non-zero exit is
// not an error.
func (c *Controller) runIsolated(ctx context.Context, ic isolatedCmd) (proc.Exit, error) {
	// One exec span per command: the program, what it is for (ic.name) and
	// the exit, never the script, env or stdin (JOB-105).
	ctx, done := telemetry.StartExec(ctx, ic.argv[0], ic.name)
	ex, err := c.runIsolatedCmd(ctx, ic)
	code := ex.ExitCode
	if err != nil {
		code = -1
	}
	done(code, err)
	return ex, err
}

func (c *Controller) runIsolatedCmd(ctx context.Context, ic isolatedCmd) (proc.Exit, error) {
	spec := proc.Spec{Command: ic.argv[0], Args: ic.argv[1:], Dir: ic.dir, Env: ic.env}
	if ic.pool != "" {
		cleanup, err := sandbox.IsolateSpec(&spec, sandbox.CommandSetup{
			Cfg: ic.cfg, Pool: ic.pool, Own: ic.own, Name: ic.name,
			TmpDir: filepath.Join(c.root.Dir, "tmp"),
			Exe:    c.hostExe,
			Cgroup: c.poolCgroup,
		})
		if cleanup != nil {
			defer cleanup()
		}
		if err != nil {
			return proc.Exit{}, fmt.Errorf("%w: %w", errSetup, err)
		}
	}
	return sandbox.Run(ctx, spec, sandbox.Streams{Stdin: ic.stdin, Stdout: ic.stdout, Stderr: ic.stderr})
}

// shArgv is sh -c command.
func shArgv(command string) []string { return []string{"sh", "-c", command} }

// isolates reports whether pool has a sandbox, mapped user or cgroup
// limits for runIsolated to apply. A pool with none runs commands as
// before, in the root dir.
func (c *Controller) isolates(cfg *config.Config, pool string) bool {
	if pool == "" {
		return false
	}
	iso := cfg.IsolationFor(pool)
	return cfg.SandboxFor(pool).Backend != "" || iso.User != nil || iso.Limits != nil
}

// exitText is "code N" or "signal S" for a command's exit.
func exitText(ex proc.Exit) string {
	if ex.Signal != "" {
		return "signal " + ex.Signal
	}
	return fmt.Sprintf("%d", ex.ExitCode)
}
