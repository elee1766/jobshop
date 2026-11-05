package control

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	host "github.com/elee1766/acphost"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// System mounts and kept env names are host-specific and come from config
// (sandbox ro and env_keep, or sandbox_profiles), not from code (#625).

// expandHome turns a leading "~/" into the controller's HOME.
func expandHome(p string) string {
	if rest, ok := strings.CutPrefix(p, "~/"); ok {
		if h, err := os.UserHomeDir(); err == nil {
			return filepath.Join(h, rest)
		}
	}
	return p
}

// sandboxFor computes the session's sandbox from config, or nil when the
// pool is not sandboxed. exe is the jobshop binary agents call. It may
// create configured rw dirs, so it runs on spawn and in the debug view,
// not for env staleness checks (sandboxKey covers those).
func (c *Controller) sandboxFor(t store.Ticket, exe, cwd string) (*host.Sandbox, error) {
	return c.poolSandbox(t.Pool, c.root.SessionDir(t.ID), exe, cwd)
}

// poolSandbox is pool's sandbox for a process whose own writable dir is
// own (a session dir, or a command step's dir) and which runs in cwd.
func (c *Controller) poolSandbox(pool, own, exe, cwd string) (*host.Sandbox, error) {
	cfg := c.config()
	sb := cfg.SandboxFor(pool)
	if sb.Backend == "" {
		return nil, nil
	}
	s := &host.Sandbox{Backend: config.SandboxBubblewrap, Network: sb.Network, Deny: sb.Deny}
	ro := func(p string) { s.RO = append(s.RO, host.Mount{Src: p, SkipMissing: true}) }
	// jobshop itself and the agent command, which may live outside the store.
	ro(filepath.Dir(exe))
	agent := cfg.Agents[cfg.Pools[pool].Agent]
	if filepath.IsAbs(agent.Command) {
		ro(filepath.Dir(agent.Command))
	}
	for _, p := range cfg.SharedTools {
		ro(p)
	}
	for _, p := range sb.RO {
		ro(expandHome(p))
	}
	rw := func(p string) { s.RW = append(s.RW, host.Mount{Src: p}) }
	// The session dir holds cli.sock, the only controller socket bound in.
	// cwd is the worktree under it, or the pool dir.
	rw(own)
	if !within(cwd, own) {
		rw(cwd)
	}
	// Worktree commits write objects and refs in the main checkout's .git.
	if p := cfg.Pools[pool]; p.Repo != "" {
		if gd := gitCommonDir(p.Repo); gd != "" {
			rw(gd)
		}
	}
	for _, p := range sb.RW {
		p = expandHome(p)
		if err := os.MkdirAll(p, 0o700); err != nil {
			return nil, host.Permanentf("sandbox rw %s: %w", p, err)
		}
		rw(p)
	}
	for _, p := range sb.Hide {
		s.Hide = append(s.Hide, expandHome(p))
	}
	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		if keepEnv(sb.EnvKeep, k) {
			s.EnvKeep = append(s.EnvKeep, k)
		}
	}
	slices.Sort(s.EnvKeep)
	s.EnvKeep = slices.Compact(s.EnvKeep)
	return s, nil
}

// keepEnv reports whether name matches keep: an exact name, or a prefix
// written with a trailing *.
func keepEnv(keep []string, name string) bool {
	return slices.ContainsFunc(keep, func(k string) bool {
		if p, ok := strings.CutSuffix(k, "*"); ok {
			return strings.HasPrefix(name, p)
		}
		return k == name
	})
}

// sandboxKey is what of the sandbox config feeds EnvHash, so a reload
// that changes the pool's sandbox respawns idle sessions like an env
// change does. Empty when the pool is not sandboxed.
func (c *Controller) sandboxKey(pool string) []string {
	cfg := c.config()
	sb := cfg.SandboxFor(pool)
	if sb.Backend == "" {
		return nil
	}
	key := []string{"JOBSHOP_SANDBOX=" + sb.Backend + " net=" + sb.Network +
		" ro=" + strings.Join(append(slices.Clone(sb.RO), cfg.SharedTools...), ",") +
		" rw=" + strings.Join(sb.RW, ",") + " hide=" + strings.Join(sb.Hide, ",") + " deny=" + strings.Join(sb.Deny, ",") +
		" env=" + strings.Join(sb.EnvKeep, ",")}
	return key
}

func within(p, root string) bool {
	return p == root || strings.HasPrefix(p, root+string(filepath.Separator))
}

// gitCommonDir is repo's shared git dir, or "" if git does not know it.
func gitCommonDir(repo string) string {
	out, err := exec.Command("git", "-C", repo, "rev-parse", "--path-format=absolute", "--git-common-dir").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// sessionSandbox shows the sandbox a session's next spawn would get: the
// resolved mounts and the bwrap argv (#262).
func (c *Controller) sessionSandbox(ctx context.Context, sc scope, p rpc.SessionSandboxParams) (any, error) {
	if err := requireAdmin(sc, "session sandbox"); err != nil {
		return nil, err
	}
	if p.ID == "" {
		return nil, fmt.Errorf("%w: session id is required", store.ErrInvalid)
	}
	t, err := c.store.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if t.Kind != store.KindSession {
		return nil, fmt.Errorf("%w: %s is a %s, not a session", store.ErrInvalid, p.ID, t.Kind)
	}
	f, err := store.DecodeFields[store.SessionFields](t)
	if err != nil {
		return nil, err
	}
	spec, exe, err := c.hostSpec(t, f)
	if err != nil {
		return nil, err
	}
	if spec.Sandbox, err = c.sandboxFor(t, exe, spec.Dir); err != nil {
		return nil, err
	}
	res := rpc.SessionSandboxResult{Session: t.ID, Pool: t.Pool, Sandbox: spec.Sandbox}
	if err := c.applyIsolation(&spec.Spec, t.Pool); err != nil {
		res.Error = err.Error()
	} else if u := c.config().IsolationFor(t.Pool).User; u != nil && u.Slots() > 0 {
		if err := applySlot(&spec.Spec, *u, f.Slot); err != nil {
			res.Error = err.Error()
		}
	}
	res.Slot = f.Slot
	res.User, res.Cgroup = spec.User, spec.Cgroup
	if res.Cgroup != nil {
		res.CgroupPath = filepath.Join(res.Cgroup.Parent, t.ID)
	}
	if spec.Sandbox != nil {
		argv, err := host.SandboxArgs(*spec.Sandbox, spec.Command, spec.Args, spec.Dir)
		if err != nil && res.Error == "" {
			res.Error = err.Error()
		}
		res.Argv = argv
	}
	return res, nil
}
