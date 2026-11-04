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

// sandboxSystemRO are host paths every sandboxed agent reads: the Nix
// store and system profile for tools, /etc for users, certs and DNS
// config, the nix-ld loader, and the resolver sockets nsswitch talks to
// (#262). Missing ones are skipped, so non-NixOS hosts work too.
var sandboxSystemRO = []string{
	"/nix/store", "/etc", "/run/current-system", "/bin", "/usr/bin", "/lib64",
	"/run/systemd/resolve", "/run/nscd",
}

// sandboxEnvKeep are controller env names passed into the sandbox on top
// of the spec env; bwrap clears the rest so host secrets stay out.
var sandboxEnvKeep = []string{"HOME", "USER", "LOGNAME", "LANG", "LC_ALL", "TZ", "TERM", "SSL_CERT_FILE", "NIX_SSL_CERT_FILE", "NIX_PATH", "NIX_LD", "NIX_LD_LIBRARY_PATH", "GOPATH", "GOCACHE", "GOFLAGS", "GOTOOLCHAIN"}

// sandboxEnvKeepPrefixes keep whole families: git identity and config
// (GIT_CONFIG_*), mise settings.
var sandboxEnvKeepPrefixes = []string{"GIT_", "MISE_"}

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
	s := &host.Sandbox{Backend: config.SandboxBubblewrap, Network: sb.Network}
	ro := func(p string) { s.RO = append(s.RO, host.Mount{Src: p, SkipMissing: true}) }
	for _, p := range sandboxSystemRO {
		ro(p)
	}
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
		if slices.Contains(sandboxEnvKeep, k) || slices.ContainsFunc(sandboxEnvKeepPrefixes, func(p string) bool { return strings.HasPrefix(k, p) }) {
			s.EnvKeep = append(s.EnvKeep, k)
		}
	}
	slices.Sort(s.EnvKeep)
	return s, nil
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
		" rw=" + strings.Join(sb.RW, ",") + " hide=" + strings.Join(sb.Hide, ",")}
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
	if spec.Sandbox != nil {
		argv, err := host.SandboxArgs(*spec.Sandbox, spec.Command, spec.Args, spec.Dir)
		if err != nil {
			res.Error = err.Error()
		}
		res.Argv = argv
	}
	return res, nil
}
