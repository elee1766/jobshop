package control

import (
	"context"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/elee1766/acphost/pkg/sandbox"

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
// home, when set, is the mapped user's home (JOB-19): "~/" rw paths bind
// from under it instead of from the controller's HOME.
func (c *Controller) sandboxFor(t store.Ticket, exe, cwd, home string) (*sandbox.Sandbox, error) {
	return c.poolSandbox(t.Pool, c.root.SessionDir(t.ID), exe, cwd, home)
}

// poolSandbox is pool's sandbox for a process whose own writable dir is
// own (a session dir, or a command step's dir) and which runs in cwd.
func (c *Controller) poolSandbox(pool, own, exe, cwd, home string) (*sandbox.Sandbox, error) {
	cfg := c.config()
	sb := cfg.SandboxFor(pool)
	if sb.Backend == "" {
		return nil, nil
	}
	s := &sandbox.Sandbox{Backend: config.SandboxBubblewrap, Network: sb.Network, Deny: sb.Deny}
	if sb.DNS != nil {
		s.DNS = &sandbox.DNS{Forward: sb.DNS.Forward, Host: sb.DNS.Host}
	}
	if sb.ResolvConf != "" {
		s.ResolvConf = expandHome(sb.ResolvConf)
	}
	ro := func(p string) { s.RO = append(s.RO, sandbox.Mount{Src: p, SkipMissing: true}) }
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
	rw := func(p string) { s.RW = append(s.RW, sandbox.Mount{Src: p}) }
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
		m := sandbox.Mount{Src: expandHome(p)}
		mode := os.FileMode(0o700)
		if rel, ok := strings.CutPrefix(p, "~/"); ok && home != "" {
			// Same path inside, the mapped user's copy outside. 0770: the
			// home's default ACL takes its mask from the group bits.
			m = sandbox.Mount{Src: filepath.Join(home, rel), Dst: m.Src}
			mode = 0o770
		}
		if err := os.MkdirAll(m.Src, mode); err != nil {
			return nil, host.Permanentf("sandbox rw %s: %w", m.Src, err)
		}
		s.RW = append(s.RW, m)
	}
	if home != "" {
		// Credentials outside every rw path still need to be seen.
		for _, p := range agent.Credentials {
			rel, ok := strings.CutPrefix(p, "~/")
			if !ok || slices.ContainsFunc(sb.RW, func(r string) bool { return within(p, r) }) {
				continue
			}
			s.RO = append(s.RO, sandbox.Mount{Src: filepath.Join(home, rel), Dst: expandHome(p), SkipMissing: true})
		}
	}
	// acphost binds parents first, so a guidance file under an rw dir
	// (~/.jcode/prompt-overlay.md under the mapped ~/.jcode) is bound over
	// it. At equal depth rw binds come last and would win: refuse that.
	for _, p := range sb.Guidance {
		if slices.Contains(sb.RW, p) {
			return nil, host.Permanentf("sandbox guidance %s is also an rw path", p)
		}
		if c := slices.IndexFunc(agent.Credentials, func(cr string) bool { return within(expandHome(cr), expandHome(p)) }); c >= 0 {
			return nil, host.Permanentf("sandbox guidance %s holds credentials file %s; credentials are copied, not bound", p, agent.Credentials[c])
		}
		ro(expandHome(p))
	}
	for _, p := range sb.Hide {
		s.Hide = append(s.Hide, expandHome(p))
	}
	ov, err := overlayMounts(sb.Overlay, own)
	if err != nil {
		return nil, err
	}
	s.Overlay = ov
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

// overlayMounts turns the merged overlay config into acphost overlays,
// sorted by target (JOB-55). A "~/" target is the same path inside the
// sandbox as an rw bind's. A persist overlay keeps its upper dir under
// own/overlay; its work dir must be empty, so it is recreated each spawn.
// A layer gone since load is a permanent error: the pool is quarantined.
func overlayMounts(ov map[string]config.SandboxOverlay, own string) ([]sandbox.Overlay, error) {
	var out []sandbox.Overlay
	for _, dst := range slices.Sorted(maps.Keys(ov)) {
		o := ov[dst]
		m := sandbox.Overlay{Dst: filepath.Clean(expandHome(dst)), Layers: o.Layers, Mode: o.Mode}
		for _, l := range o.Layers {
			if st, err := os.Stat(l); err != nil || !st.IsDir() {
				return nil, host.Permanentf("sandbox overlay %s: layer %s is missing or not a directory", dst, l)
			}
		}
		if o.Mode == config.OverlayPersist {
			base := filepath.Join(own, "overlay", strings.ReplaceAll(strings.TrimPrefix(m.Dst, "/"), "/", "_"))
			m.Upper, m.Work = filepath.Join(base, "upper"), filepath.Join(base, "work")
			if err := os.RemoveAll(m.Work); err != nil {
				return nil, host.Permanentf("sandbox overlay %s: %w", dst, err)
			}
			for _, d := range []string{m.Upper, m.Work} {
				if err := os.MkdirAll(d, 0o700); err != nil {
					return nil, host.Permanentf("sandbox overlay %s: %w", dst, err)
				}
			}
		}
		out = append(out, m)
	}
	return out, nil
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
		" rw=" + strings.Join(sb.RW, ",") + " hide=" + strings.Join(sb.Hide, ",") + " guidance=" + strings.Join(sb.Guidance, ",") + " deny=" + strings.Join(sb.Deny, ",") +
		" env=" + strings.Join(sb.EnvKeep, ",") + " " + dnsKey(sb) + overlayKey(sb)}
	return key
}

// overlayKey lists each overlay target's mode and layers.
func overlayKey(sb config.Sandbox) string {
	var k string
	for _, dst := range slices.Sorted(maps.Keys(sb.Overlay)) {
		o := sb.Overlay[dst]
		k += " overlay=" + dst + ":" + o.Mode + ":" + strings.Join(o.Layers, ",")
	}
	return k
}

// dnsKey is the part of sb the pasta DNS check depends on.
func dnsKey(sb config.Sandbox) string {
	k := "resolv=" + sb.ResolvConf + " dnscheck=" + sb.DNSCheck + " deny=" + strings.Join(sb.Deny, ",")
	if sb.DNS != nil {
		k += " dns=" + sb.DNS.Forward + ">" + sb.DNS.Host
	}
	return k
}

// checkSandboxDNS resolves the pool's dns_check name inside a pasta
// namespace set up like the agent's (JOB-20), once per pool and DNS
// setting: it spawns pasta, so it is not run per session. A failure is
// permanent and quarantines the pool with an error naming sandbox.dns,
// instead of every turn of the agent failing on name resolution.
func (c *Controller) checkSandboxDNS(pool string, s *sandbox.Sandbox) error {
	sb := c.config().SandboxFor(pool)
	if s == nil || sb.Network != config.NetworkPasta || sb.DNSCheck == "" {
		return nil
	}
	key := pool + "\x00" + dnsKey(sb)
	c.mu.Lock()
	ok := c.dnsChecked[key]
	c.mu.Unlock()
	if ok {
		return nil
	}
	if err := sandbox.CheckDNS(*s, sb.DNSCheck); err != nil {
		return host.Permanentf("no DNS in sandbox: %w", err)
	}
	c.mu.Lock()
	if c.dnsChecked == nil {
		c.dnsChecked = map[string]bool{}
	}
	c.dnsChecked[key] = true
	c.mu.Unlock()
	return nil
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
		argv, err := sandbox.Args(*spec.Sandbox, spec.Command, spec.Args, spec.Dir)
		if err != nil && res.Error == "" {
			res.Error = err.Error()
		}
		res.Argv = argv
	}
	return res, nil
}
