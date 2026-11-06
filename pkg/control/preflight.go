package control

import (
	"context"
	"encoding/json"
	"maps"
	"slices"
	"sync"

	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/elee1766/acphost/pkg/sandbox"
	"github.com/elee1766/acphost/pkg/userns"

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
		s.Sandbox = &sandbox.Sandbox{Network: sb.Network, Deny: sb.Deny}
		if sb.DNS != nil {
			s.Sandbox.DNS = &sandbox.DNS{Forward: sb.DNS.Forward, Host: sb.DNS.Host}
		}
		if sb.ResolvConf != "" {
			s.Sandbox.ResolvConf = expandHome(sb.ResolvConf)
		}
	}
	iso := cfg.IsolationFor(pool)
	if u := iso.User; u != nil {
		hu := &userns.User{UID: u.UID, GID: u.GIDOrUID(), MapUID: u.MapUID, MapGID: u.MapGID, MapCount: u.Slots()}
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

// preflightParallel bounds the preflight checks run at once. A dns_check
// spawns pasta and can take seconds.
const preflightParallel = 8

// preflightResult is one distinct check's report.
type preflightResult struct {
	rep host.PreflightReport
	err error
}

// preflightPools runs host.Preflight for every isolated pool at start and
// after a reload, and quarantines a pool whose host is missing something
// (#728). The escalation then names each failed check and its fix before
// any session spawns, instead of after the first spawn fails.
//
// Pools that ask for the same isolation share one check, and distinct
// checks run in parallel: a reload holds the tick lock through this, and
// one serial dns_check per pool took a 58-pool config past the CLI's
// deadline (JOB-102).
func (c *Controller) preflightPools(ctx context.Context) {
	cfg := c.config()
	check := c.opt.Preflight
	if check == nil {
		check = host.Preflight
	}
	type job struct {
		spec host.Spec
		dns  string
	}
	jobs := map[string]job{}
	keyOf := map[string]string{}
	for _, pool := range slices.Sorted(maps.Keys(cfg.Pools)) {
		spec, ok := c.preflightSpec(cfg, pool)
		if !ok {
			continue
		}
		dns := cfg.SandboxFor(pool).DNSCheck
		key, err := json.Marshal(struct {
			Sandbox, User, Cgroup any
			DNS                   string
		}{spec.Sandbox, spec.User, spec.Cgroup, dns})
		if err != nil {
			// Not expected: the spec is plain data. Check it alone.
			key = []byte("pool:" + pool)
		}
		keyOf[pool] = string(key)
		jobs[string(key)] = job{spec: spec, dns: dns}
	}
	results := make(map[string]preflightResult, len(jobs))
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, preflightParallel)
	for key, j := range jobs {
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()
			var opts []host.PreflightOption
			if j.dns != "" {
				// Spawns pasta, so only here, once per start or reload.
				opts = append(opts, host.WithDNSCheck(j.dns))
			}
			rep, err := check(j.spec, opts...)
			mu.Lock()
			results[key] = preflightResult{rep: rep, err: err}
			mu.Unlock()
		})
	}
	wg.Wait()
	for _, pool := range slices.Sorted(maps.Keys(keyOf)) {
		res := results[keyOf[pool]]
		if res.err != nil {
			c.log.WarnContext(ctx, "isolation preflight", "pool", pool, "err", res.err)
			continue
		}
		if !res.rep.OK() {
			c.quarantinePool(ctx, pool, "isolation preflight:\n"+res.rep.String())
		}
	}
}
