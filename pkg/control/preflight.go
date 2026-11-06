package control

import (
	"context"
	"maps"
	"slices"

	host "github.com/elee1766/acphost"

	"github.com/elee1766/jobshop/pkg/config"
)

// preflightSpec is what pool's sessions ask the host for, enough for
// host.Preflight (#728): the sandbox network, the mapped user and the
// cgroup limits. Nil when the pool uses none of them.
func (c *Controller) preflightSpec(cfg *config.Config, pool string) (host.Spec, bool) {
	var s host.Spec
	if sb := cfg.SandboxFor(pool); sb.Backend == config.SandboxBubblewrap {
		// Deny makes Preflight check nft; DNS and ResolvConf are what
		// its sandbox.dns check sets up (acphost v0.9.0).
		s.Sandbox = &host.Sandbox{Network: sb.Network, Deny: sb.Deny}
		if sb.DNS != nil {
			s.Sandbox.DNS = &host.DNS{Forward: sb.DNS.Forward, Host: sb.DNS.Host}
		}
		if sb.ResolvConf != "" {
			s.Sandbox.ResolvConf = expandHome(sb.ResolvConf)
		}
	}
	iso := cfg.IsolationFor(pool)
	if u := iso.User; u != nil {
		hu := &host.User{UID: u.UID, GID: u.GIDOrUID(), MapUID: u.MapUID, MapGID: u.MapGID, MapCount: u.Slots()}
		for _, g := range u.Groups {
			if gid, err := config.LookupGroup(g); err == nil {
				hu.Groups = append(hu.Groups, gid)
			}
		}
		s.User = hu
	}
	if l := iso.Limits; l != nil {
		// The parent is the sessions cgroup the controller set up; a
		// setup failure is reported at spawn (applyIsolation).
		if cg := c.ensureCgroup(); cg.err == nil {
			if hc, err := cgroupFor(*l, cg.sessions); err == nil {
				s.Cgroup = hc
			}
		}
	}
	return s, s.Sandbox != nil || s.User != nil || s.Cgroup != nil
}

// preflightPools runs host.Preflight for every isolated pool at start and
// after a reload, and quarantines a pool whose host is missing something
// (#728). The escalation then names each failed check and its fix before
// any session spawns, instead of after the first spawn fails.
func (c *Controller) preflightPools(ctx context.Context) {
	cfg := c.config()
	check := c.opt.Preflight
	if check == nil {
		check = host.Preflight
	}
	for _, pool := range slices.Sorted(maps.Keys(cfg.Pools)) {
		spec, ok := c.preflightSpec(cfg, pool)
		if !ok {
			continue
		}
		var opts []host.PreflightOption
		if name := cfg.SandboxFor(pool).DNSCheck; name != "" {
			// Spawns pasta, so only here, once per start or reload.
			opts = append(opts, host.WithDNSCheck(name))
		}
		rep, err := check(spec, opts...)
		if err != nil {
			c.log.WarnContext(ctx, "isolation preflight", "pool", pool, "err", err)
			continue
		}
		if !rep.OK() {
			c.quarantinePool(ctx, pool, "isolation preflight:\n"+rep.String())
		}
	}
}
