package control

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/elee1766/acphost/pkg/proc"
	"github.com/elee1766/acphost/pkg/userns"

	"github.com/elee1766/jobshop/pkg/config"
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
var errSetup = errors.New("isolation setup failed")

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
		cleanup, err := c.isolateSpec(&spec, ic)
		if cleanup != nil {
			defer cleanup()
		}
		if err != nil {
			return proc.Exit{}, fmt.Errorf("%w: %w", errSetup, err)
		}
	}
	// The parent-death signal follows the thread that started it, so
	// keep this goroutine on it until the command is done.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	opts := []proc.Option{}
	if ic.stdin != nil {
		opts = append(opts, proc.Stdin(ic.stdin))
	}
	if ic.stdout != nil {
		opts = append(opts, proc.Stdout(ic.stdout))
	}
	if ic.stderr != nil {
		opts = append(opts, proc.Stderr(ic.stderr))
	}
	p, err := proc.Start(ctx, spec, opts...)
	if err != nil {
		return proc.Exit{}, err
	}
	ex, err := p.Wait()
	if err == nil && ex.SetupError != "" {
		err = fmt.Errorf("%w: %s", errSetup, ex.SetupError)
	}
	return ex, err
}

// isolateSpec adds pool's sandbox, user and cgroup to spec. The cleanup
// removes the scratch dir it made.
func (c *Controller) isolateSpec(spec *proc.Spec, ic isolatedCmd) (func(), error) {
	cfg := ic.cfg
	if cfg == nil {
		return nil, fmt.Errorf("isolated command for pool %s: no config", ic.pool)
	}
	var cleanup func()
	sbCfg := cfg.SandboxFor(ic.pool)
	if sbCfg.Backend != "" {
		own := ic.own
		if own == "" {
			base := filepath.Join(c.root.Dir, "tmp")
			if err := os.MkdirAll(base, 0o700); err != nil {
				return nil, err
			}
			d, err := os.MkdirTemp(base, "cmd-")
			if err != nil {
				return nil, err
			}
			own, cleanup = d, func() { _ = os.RemoveAll(d) }
		}
		if spec.Dir == "" {
			spec.Dir = own
		}
		exe, err := c.hostExe()
		if err != nil {
			return cleanup, err
		}
		sb, err := c.poolSandboxIn(cfg, ic.pool, own, exe, spec.Dir, "")
		if err != nil {
			return cleanup, err
		}
		if sb != nil {
			// bwrap clears the env and PATH is not in sandboxEnvKeep
			// (sessions set their own), so pass the controller's.
			if !hasEnv(spec.Env, "PATH") {
				spec.Env = append([]string{"PATH=" + os.Getenv("PATH")}, spec.Env...)
			}
			spec.Sandbox = sb
		}
	}
	iso := cfg.IsolationFor(ic.pool)
	if u := iso.User; u != nil {
		hu := &userns.User{UID: u.UID, GID: u.GIDOrUID(), MapUID: u.MapUID, MapGID: u.MapGID}
		for _, g := range u.Groups {
			gid, err := config.LookupGroup(g)
			if err != nil {
				return cleanup, fmt.Errorf("isolation user: %w", err)
			}
			hu.Groups = append(hu.Groups, gid)
		}
		spec.User = hu
	}
	if l := iso.Limits; l != nil {
		cg := c.ensureCgroup()
		if cg.err != nil {
			return cleanup, fmt.Errorf("isolation limits: no delegated cgroup: %w", cg.err)
		}
		hc, err := cgroupFor(*l, cg.sessions)
		if err != nil {
			return cleanup, fmt.Errorf("isolation limits: %w", err)
		}
		b := make([]byte, 6)
		_, _ = rand.Read(b)
		spec.Cgroup, spec.Name = hc, firstNonEmpty(ic.name, "cmd")+"-"+hex.EncodeToString(b)
	}
	return cleanup, nil
}

func hasEnv(env []string, key string) bool {
	for _, kv := range env {
		if strings.HasPrefix(kv, key+"=") {
			return true
		}
	}
	return false
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
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
