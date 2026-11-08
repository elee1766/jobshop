package sessionrt

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/elee1766/acphost/pkg/cgroup"
	"github.com/elee1766/acphost/pkg/proc"
	"github.com/elee1766/acphost/pkg/userns"

	"github.com/elee1766/jobshop/pkg/config"
)

// ErrSetup is a sandbox or isolation failure: the command never ran.
var ErrSetup = errors.New("isolation setup failed")

// UserFor is the acphost user for a pool's isolation.user, with its
// groups resolved. The error is not marked permanent: callers decide.
func UserFor(u config.IsolationUser) (*userns.User, error) {
	hu := &userns.User{UID: u.UID, GID: u.GIDOrUID(), MapUID: u.MapUID, MapGID: u.MapGID}
	for _, g := range u.Groups {
		gid, err := config.LookupGroup(g)
		if err != nil {
			return nil, fmt.Errorf("isolation user: %w", err)
		}
		hu.Groups = append(hu.Groups, gid)
	}
	return hu, nil
}

// CommandSetup is what IsolateSpec needs to give a controller command
// its pool's sandbox, user and cgroup. The callbacks keep the controller
// state (binary path, delegated cgroup) on the controller's side.
type CommandSetup struct {
	// Cfg is the config the pool's isolation comes from.
	Cfg  *config.Config
	Pool string
	// Own is bound rw in the sandbox (a run dir). Empty: a scratch dir
	// under TmpDir, removed by the returned cleanup.
	Own string
	// TmpDir is where scratch dirs are made, created if missing.
	TmpDir string
	// Name is the cgroup name prefix, "cmd" when empty.
	Name string
	// Exe is the jobshop binary agents call, asked only for a sandbox.
	Exe func() (string, error)
	// Cgroup is the pool's acphost cgroup, asked only for a pool with
	// limits.
	Cgroup func(cfg *config.Config, pool string) (*cgroup.Cgroup, error)
}

// IsolateSpec adds the pool's sandbox, user and cgroup to spec. The
// cleanup, when not nil, removes the scratch dir it made and is returned
// with an error too.
func IsolateSpec(spec *proc.Spec, s CommandSetup) (func(), error) {
	cfg := s.Cfg
	if cfg == nil {
		return nil, fmt.Errorf("isolated command for pool %s: no config", s.Pool)
	}
	var cleanup func()
	if cfg.SandboxFor(s.Pool).Backend != "" {
		own := s.Own
		if own == "" {
			if err := os.MkdirAll(s.TmpDir, 0o700); err != nil {
				return nil, err
			}
			d, err := os.MkdirTemp(s.TmpDir, "cmd-")
			if err != nil {
				return nil, err
			}
			own, cleanup = d, func() { _ = os.RemoveAll(d) }
		}
		if spec.Dir == "" {
			spec.Dir = own
		}
		exe, err := s.Exe()
		if err != nil {
			return cleanup, err
		}
		sb, err := BuildSandbox(cfg, s.Pool, own, exe, spec.Dir, "")
		if err != nil {
			return cleanup, err
		}
		if sb != nil {
			// bwrap clears the env and PATH is not in sandboxEnvKeep
			// (sessions set their own), so pass the controller's.
			if !HasEnv(spec.Env, "PATH") {
				spec.Env = append([]string{"PATH=" + os.Getenv("PATH")}, spec.Env...)
			}
			spec.Sandbox = sb
		}
	}
	iso := cfg.IsolationFor(s.Pool)
	if u := iso.User; u != nil {
		hu, err := UserFor(*u)
		if err != nil {
			return cleanup, err
		}
		spec.User = hu
	}
	if iso.Limits != nil {
		hc, err := s.Cgroup(cfg, s.Pool)
		if err != nil {
			return cleanup, err
		}
		b := make([]byte, 6)
		_, _ = rand.Read(b)
		name := s.Name
		if name == "" {
			name = "cmd"
		}
		spec.Cgroup, spec.Name = hc, name+"-"+hex.EncodeToString(b)
	}
	return cleanup, nil
}

// HasEnv reports whether env sets key.
func HasEnv(env []string, key string) bool {
	for _, kv := range env {
		if strings.HasPrefix(kv, key+"=") {
			return true
		}
	}
	return false
}

// Streams are a command's optional stdio.
type Streams struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// Run starts spec and waits for it. A start failure or wait error is
// returned as is; a sandbox or isolation failure the process reported
// wraps ErrSetup. A non-zero exit is not an error.
func Run(ctx context.Context, spec proc.Spec, st Streams) (proc.Exit, error) {
	// The parent-death signal follows the thread that started it, so
	// keep this goroutine on it until the command is done.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	var opts []proc.Option
	if st.Stdin != nil {
		opts = append(opts, proc.Stdin(st.Stdin))
	}
	if st.Stdout != nil {
		opts = append(opts, proc.Stdout(st.Stdout))
	}
	if st.Stderr != nil {
		opts = append(opts, proc.Stderr(st.Stderr))
	}
	p, err := proc.Start(ctx, spec, opts...)
	if err != nil {
		return proc.Exit{}, err
	}
	ex, err := p.Wait()
	if err == nil && ex.SetupError != "" {
		err = fmt.Errorf("%w: %s", ErrSetup, ex.SetupError)
	}
	return ex, err
}
