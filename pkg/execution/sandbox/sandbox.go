package sandbox

import (
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/gfx-labs/acphost/pkg/sandbox"
	"github.com/gfx-labs/acphost/pkg/xerr"

	"github.com/gfx-labs/jobshop/pkg/config"
)

// System mounts and kept env names are host-specific and come from config
// (sandbox ro and env_keep, or sandbox_profiles), not from code (#625).

// ExpandHome turns a leading "~/" into the controller's HOME.
func ExpandHome(p string) string {
	if rest, ok := strings.CutPrefix(p, "~/"); ok {
		if h, err := os.UserHomeDir(); err == nil {
			return filepath.Join(h, rest)
		}
	}
	return p
}

// Within reports whether p is root or a path under it.
func Within(p, root string) bool {
	return p == root || strings.HasPrefix(p, root+string(filepath.Separator))
}

// GitCommonDir is repo's shared git dir, or "" if git does not know it.
func GitCommonDir(repo string) string {
	out, err := exec.Command("git", "-C", repo, "rev-parse", "--path-format=absolute", "--git-common-dir").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// BuildSandbox computes pool's sandbox under cfg, or nil when the pool is
// not sandboxed. exe is the jobshop binary agents call. own is the
// process's own writable dir (a session dir, or a command step's dir) and
// cwd where it runs. home, when set, is the mapped user's home (JOB-19):
// "~/" rw paths bind from under it instead of from the controller's HOME.
// It may create configured rw dirs and overlay dirs, so it runs on spawn
// and in the debug view, not for env staleness checks (Key covers
// those).
func BuildSandbox(cfg *config.Config, pool, own, exe, cwd, home string) (*sandbox.Sandbox, error) {
	sb := cfg.SandboxFor(pool)
	if sb.Backend == "" {
		return nil, nil
	}
	s := &sandbox.Sandbox{Backend: config.SandboxBubblewrap, Network: sb.Network, Deny: sb.Deny}
	if sb.DNS != nil {
		s.DNS = &sandbox.DNS{Forward: sb.DNS.Forward, Host: sb.DNS.Host}
	}
	if sb.ResolvConf != "" {
		s.ResolvConf = ExpandHome(sb.ResolvConf)
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
		ro(ExpandHome(p))
	}
	rw := func(p string) { s.RW = append(s.RW, sandbox.Mount{Src: p}) }
	// The session dir holds cli.sock, the only controller socket bound in.
	// cwd is the worktree under it, or the pool dir.
	rw(own)
	if !Within(cwd, own) {
		rw(cwd)
	}
	// Worktree commits write objects and refs in the main checkout's .git:
	// the one cwd belongs to (a keyed repo's checkout, JOB-207), else the
	// pool repo's (a worktree not made yet).
	if p := cfg.Pools[pool]; p.Repo != "" {
		gd := GitCommonDir(cwd)
		if gd == "" {
			gd = GitCommonDir(p.Repo)
		}
		if gd != "" {
			rw(gd)
		}
	}
	for _, p := range sb.RW {
		m := sandbox.Mount{Src: ExpandHome(p)}
		mode := os.FileMode(0o700)
		if rel, ok := strings.CutPrefix(p, "~/"); ok && home != "" {
			// Same path inside, the mapped user's copy outside. 0770: the
			// home's default ACL takes its mask from the group bits.
			m = sandbox.Mount{Src: filepath.Join(home, rel), Dst: m.Src}
			mode = 0o770
		}
		if err := os.MkdirAll(m.Src, mode); err != nil {
			return nil, xerr.Permanentf("sandbox rw %s: %w", m.Src, err)
		}
		s.RW = append(s.RW, m)
	}
	if home != "" {
		// Credentials outside every rw path still need to be seen.
		for _, p := range agent.Credentials {
			rel, ok := strings.CutPrefix(p, "~/")
			if !ok || slices.ContainsFunc(sb.RW, func(r string) bool { return Within(p, r) }) {
				continue
			}
			s.RO = append(s.RO, sandbox.Mount{Src: filepath.Join(home, rel), Dst: ExpandHome(p), SkipMissing: true})
		}
	}
	// acphost binds parents first, so a guidance file under an rw dir
	// (~/.jcode/prompt-overlay.md under the mapped ~/.jcode) is bound over
	// it. At equal depth rw binds come last and would win: refuse that.
	for _, p := range sb.Guidance {
		if slices.Contains(sb.RW, p) {
			return nil, xerr.Permanentf("sandbox guidance %s is also an rw path", p)
		}
		if c := slices.IndexFunc(agent.Credentials, func(cr string) bool { return Within(ExpandHome(cr), ExpandHome(p)) }); c >= 0 {
			return nil, xerr.Permanentf("sandbox guidance %s holds credentials file %s; credentials are copied, not bound", p, agent.Credentials[c])
		}
		ro(ExpandHome(p))
	}
	for _, p := range sb.Hide {
		s.Hide = append(s.Hide, ExpandHome(p))
	}
	ov, err := overlayMounts(sb.Overlay, own)
	if err != nil {
		return nil, err
	}
	s.Overlay = ov
	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		if KeepEnv(sb.EnvKeep, k) {
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
		m := sandbox.Overlay{Dst: filepath.Clean(ExpandHome(dst)), Layers: o.Layers, Mode: o.Mode}
		for _, l := range o.Layers {
			if st, err := os.Stat(l); err != nil || !st.IsDir() {
				return nil, xerr.Permanentf("sandbox overlay %s: layer %s is missing or not a directory", dst, l)
			}
		}
		if o.Mode == config.OverlayPersist {
			base := filepath.Join(own, "overlay", strings.ReplaceAll(strings.TrimPrefix(m.Dst, "/"), "/", "_"))
			m.Upper, m.Work = filepath.Join(base, "upper"), filepath.Join(base, "work")
			if err := os.RemoveAll(m.Work); err != nil {
				return nil, xerr.Permanentf("sandbox overlay %s: %w", dst, err)
			}
			for _, d := range []string{m.Upper, m.Work} {
				if err := os.MkdirAll(d, 0o700); err != nil {
					return nil, xerr.Permanentf("sandbox overlay %s: %w", dst, err)
				}
			}
		}
		out = append(out, m)
	}
	return out, nil
}

// KeepEnv reports whether name matches keep: an exact name, or a prefix
// written with a trailing *.
func KeepEnv(keep []string, name string) bool {
	return slices.ContainsFunc(keep, func(k string) bool {
		if p, ok := strings.CutSuffix(k, "*"); ok {
			return strings.HasPrefix(name, p)
		}
		return k == name
	})
}

// Key is what of pool's sandbox config feeds EnvHash, so a reload
// that changes the pool's sandbox respawns idle sessions like an env
// change does. Nil when the pool is not sandboxed.
func Key(cfg *config.Config, pool string) []string {
	sb := cfg.SandboxFor(pool)
	if sb.Backend == "" {
		return nil
	}
	return []string{"JOBSHOP_SANDBOX=" + sb.Backend + " net=" + sb.Network +
		" ro=" + strings.Join(append(slices.Clone(sb.RO), cfg.SharedTools...), ",") +
		" rw=" + strings.Join(sb.RW, ",") + " hide=" + strings.Join(sb.Hide, ",") + " guidance=" + strings.Join(sb.Guidance, ",") + " deny=" + strings.Join(sb.Deny, ",") +
		" env=" + strings.Join(sb.EnvKeep, ",") + " " + dnsKey(sb) + overlayKey(sb)}
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

// DNSChecks remembers which pool and DNS settings passed the sandbox DNS
// check, so it runs once per setting. The zero value is ready; it is safe
// for concurrent use.
type DNSChecks struct {
	mu sync.Mutex
	ok map[string]bool
}

// Check resolves the pool's dns_check name inside a pasta namespace set
// up like the agent's (JOB-20), once per pool and DNS setting: it spawns
// pasta, so it is not run per session. A failure is permanent and
// quarantines the pool with an error naming sandbox.dns, instead of every
// turn of the agent failing on name resolution. s is the pool's sandbox
// as BuildSandbox made it; cfg is the config it was built from.
func (d *DNSChecks) Check(cfg *config.Config, pool string, s *sandbox.Sandbox) error {
	sb := cfg.SandboxFor(pool)
	if s == nil || sb.Network != config.NetworkPasta || sb.DNSCheck == "" {
		return nil
	}
	key := pool + "\x00" + dnsKey(sb)
	d.mu.Lock()
	ok := d.ok[key]
	d.mu.Unlock()
	if ok {
		return nil
	}
	if err := sandbox.CheckDNS(*s, sb.DNSCheck); err != nil {
		return xerr.Permanentf("no DNS in sandbox: %w", err)
	}
	d.mu.Lock()
	if d.ok == nil {
		d.ok = map[string]bool{}
	}
	d.ok[key] = true
	d.mu.Unlock()
	return nil
}

// Len is how many settings have passed.
func (d *DNSChecks) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.ok)
}
