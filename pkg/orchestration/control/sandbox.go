package control

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gfx-labs/acphost/pkg/cgroup"
	"github.com/gfx-labs/acphost/pkg/sandbox"

	"github.com/gfx-labs/jobshop/pkg/domain"
	execsandbox "github.com/gfx-labs/jobshop/pkg/execution/sandbox"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// The sandbox spec is built by execsandbox.BuildSandbox from config; the
// methods here read the live config and pass it in.

// sandboxFor computes the session's sandbox from config, or nil when the
// pool is not sandboxed. exe is the jobshop binary agents call. It may
// create configured rw dirs, so it runs on spawn and in the debug view,
// not for env staleness checks (sandboxKey covers those).
// home, when set, is the mapped user's home (JOB-19).
func (c *Controller) sandboxFor(t domain.Ticket, exe, cwd, home string) (*sandbox.Sandbox, error) {
	return c.poolSandbox(t.Pool, c.root.SessionDir(t.ID), exe, cwd, home)
}

// poolSandbox is pool's sandbox for a process whose own writable dir is
// own (a session dir, or a command step's dir) and which runs in cwd.
func (c *Controller) poolSandbox(pool, own, exe, cwd, home string) (*sandbox.Sandbox, error) {
	return execsandbox.BuildSandbox(c.config(), pool, own, exe, cwd, home)
}

// sandboxKey is what of the sandbox config feeds EnvHash (see
// execsandbox.Key).
func (c *Controller) sandboxKey(pool string) []string {
	return execsandbox.Key(c.config(), pool)
}

// checkSandboxDNS runs the pool's pasta DNS check once per setting.
func (c *Controller) checkSandboxDNS(pool string, s *sandbox.Sandbox) error {
	return c.dnsChecks.Check(c.config(), pool, s)
}

// sessionSandbox shows the sandbox a session's next spawn would get: the
// resolved mounts and the bwrap argv (#262).
func (c *Controller) sessionSandbox(ctx context.Context, sc scope, p rpc.SessionSandboxParams) (any, error) {
	if err := requireAdmin(sc, "session sandbox"); err != nil {
		return nil, err
	}
	if p.ID == "" {
		return nil, fmt.Errorf("%w: session id is required", domain.ErrInvalid)
	}
	t, err := c.store.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if t.Kind != domain.KindSession {
		return nil, fmt.Errorf("%w: %s is a %s, not a session", domain.ErrInvalid, p.ID, t.Kind)
	}
	f, err := domain.DecodeFields[domain.SessionFields](t)
	if err != nil {
		return nil, err
	}
	spec, exe, err := c.hostSpec(t, f)
	if err != nil {
		return nil, err
	}
	home := ""
	if _, _, ok, _ := c.mappedUser(t.Pool, f.Slot); ok {
		home = c.mappedHomeDir(t.Pool, f.Slot)
	}
	if spec.Sandbox, err = c.sandboxFor(t, exe, spec.Dir, home); err != nil {
		return nil, err
	}
	res := rpc.SessionSandboxResult{Session: t.ID, Pool: t.Pool, Sandbox: spec.Sandbox}
	if err := c.applyIsolation(&spec.Spec, t.Pool); err != nil {
		res.Error = err.Error()
	} else if u := c.config().IsolationFor(t.Pool).User; u != nil && u.Slots() > 0 {
		if err := execsandbox.ApplySlot(spec.User, *u, f.Slot); err != nil {
			res.Error = err.Error()
		}
	}
	res.Slot = f.Slot
	res.User, res.Cgroup = spec.User, spec.Cgroup
	if res.Cgroup != nil {
		if res.Cgroup.Driver == cgroup.DriverSystemd {
			// The live scope, if any: its name holds the holder's pid.
			res.CgroupPath = c.sessionCgroupDir(c.config(), t.ID)
		} else {
			res.CgroupPath = filepath.Join(res.Cgroup.Parent, t.ID)
		}
	}
	if spec.Sandbox != nil {
		argv, err := sandbox.Args(*spec.Sandbox, spec.Command, spec.Args, spec.Dir)
		if err != nil && res.Error == "" {
			res.Error = err.Error()
		}
		res.Argv = argv
	}
	return res, nil
}
