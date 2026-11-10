package control

import (
	"cmp"
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/conversation/attention"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/integrations/identity"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// Host outages (JOB-194). When an identity cannot get credentials (the
// token command fails, GitHub or Linear answers 401, 403 or 5xx, the
// endpoint is unreachable) the fault is the host's, not the work's. The
// controller then pauses the pools that use the identity, like a rate
// limit backoff: they claim nothing and scale up nothing, a step released
// or closed failed meanwhile counts no attempt, and no escalation is
// filed. A probe on the controller's clock mints a token again and lifts
// the pause when it works; the steps then run with no undefer.

// hostOutage is one identity that cannot get credentials.
type hostOutage struct {
	identity  string
	pools     []string // pools paused by it, chat pools excluded
	reason    string   // the first error, one line
	last      string   // the latest probe or mint error, one line
	since     time.Time
	nextProbe time.Time
}

// outageProbeTimeout bounds one probe: a mint plus one GET.
const outageProbeTimeout = 45 * time.Second

// outageProbeKey is the off-pass call key of an identity's probe.
func outageProbeKey(name string) string { return "outage\x00" + name }

// outageLine is err as one short line.
func outageLine(err error) string {
	s, _, _ := strings.Cut(strings.TrimSpace(err.Error()), "\n")
	if len(s) > 300 {
		s = s[:300]
	}
	return s
}

// outagePools lists the ticket pools whose sessions may act as identity
// name: through the pool's own identities or its project's. A chat pool
// is left out, so the operator can still talk to a controller whose
// credentials are broken.
func outagePools(cfg *config.Config, name string) []string {
	var out []string
	for pn, p := range cfg.Pools {
		if p.Chat() {
			continue
		}
		proj := cmp.Or(p.Project, cfg.DefaultProject)
		if slices.Contains(p.Identities, name) || slices.Contains(cfg.Projects[proj].Identities, name) {
			out = append(out, pn)
		}
	}
	slices.Sort(out)
	return out
}

// noteIdentityError pauses the pools of identity name when err says the
// host cannot get credentials. Other errors (a refused target, a bad
// request) are the caller's and change nothing.
func (c *Controller) noteIdentityError(ctx context.Context, name string, err error) {
	if !identity.Outage(err) {
		return
	}
	pools := outagePools(c.config(), name)
	if len(pools) == 0 {
		return
	}
	line := outageLine(err)
	now := c.now()
	c.mu.Lock()
	o, ok := c.outages[name]
	if !ok {
		if c.outages == nil {
			c.outages = map[string]*hostOutage{}
		}
		o = &hostOutage{identity: name, since: now, reason: line, nextProbe: now.Add(c.opt.OutageProbe)}
		c.outages[name] = o
	}
	o.pools, o.last = pools, line
	c.mu.Unlock()
	if !ok {
		c.log.WarnContext(ctx, "host outage: pools paused", "identity", name, "pools", pools, "err", line)
	}
}

// poolOutaged reports whether pool is paused by a host outage.
func (c *Controller) poolOutaged(pool string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, o := range c.outages {
		if slices.Contains(o.pools, pool) {
			return true
		}
	}
	return false
}

// outageHold reports whether a step of t's pool is held up by a host
// outage, and why. A failed close or a plain release of such a step is
// not an attempt of the step's own: the caller hands it back uncounted.
func (c *Controller) outageHold(t domain.Ticket) (string, bool) {
	if t.Kind != domain.KindStep || t.Pool == "" {
		return "", false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, name := range slices.Sorted(maps.Keys(c.outages)) {
		if o := c.outages[name]; slices.Contains(o.pools, t.Pool) {
			return fmt.Sprintf("identity %s: %s", name, o.reason), true
		}
	}
	return "", false
}

// releaseForOutage hands the step held under cl back to its pool now,
// counting no attempt, and says why on the ticket. how names what the
// agent did (closed failed, released).
func (c *Controller) releaseForOutage(ctx context.Context, cfg *config.Config, actor string, cl store.Claim, how, why string) (domain.Ticket, error) {
	t, err := c.store.Release(ctx, actor, cl.TicketID, &cl, store.Outcome{Reason: "host outage: " + why, Until: c.now()})
	if err != nil {
		return t, err
	}
	note := fmt.Sprintf("Step %s during a host outage (%s): handed back without counting a failed attempt. The pool starts no claims until credentials work again.", how, why)
	if _, err := store.AddComment(ctx, c.store, "controller", t.ID, note); err != nil {
		c.log.WarnContext(ctx, "outage comment", "ticket", t.ID, "err", err)
	}
	c.stepEvent(ctx, cfg, t)
	return t, nil
}

// probeOutages applies finished probes and starts the due ones off the
// pass. A successful probe lifts the pause; a failed one waits another
// OutageProbe.
func (c *Controller) probeOutages(ctx context.Context, cfg *config.Config) {
	now := c.now()
	c.mu.Lock()
	var due []string
	for name, o := range c.outages {
		if !now.Before(o.nextProbe) {
			due = append(due, name)
		}
	}
	c.mu.Unlock()
	slices.Sort(due)
	for _, name := range due {
		key := outageProbeKey(name)
		if v, ok := c.offTake(key); ok {
			err, _ := v.(error)
			c.probed(ctx, name, err)
			continue
		}
		c.offGo(key, "outage", outageProbeTimeout, true, func(ctx context.Context) any {
			if err := c.probeIdentity(ctx, cfg, name); err != nil {
				return err
			}
			return nil
		})
	}
}

// probed records a probe's result.
func (c *Controller) probed(ctx context.Context, name string, err error) {
	c.mu.Lock()
	o, ok := c.outages[name]
	if !ok {
		c.mu.Unlock()
		return
	}
	if err != nil {
		o.last = outageLine(err)
		o.nextProbe = c.now().Add(c.opt.OutageProbe)
		c.mu.Unlock()
		c.log.InfoContext(ctx, "host outage: probe failed", "identity", name, "err", o.last)
		return
	}
	delete(c.outages, name)
	since, pools := o.since, o.pools
	c.mu.Unlock()
	c.log.InfoContext(ctx, "host outage over: pools resume", "identity", name, "pools", pools, "paused_for", c.now().Sub(since))
}

// probeIdentity mints a fresh token for identity name and uses it: GET
// /rate_limit for GitHub, the viewer for Linear.
func (c *Controller) probeIdentity(ctx context.Context, cfg *config.Config, name string) error {
	id, ok := cfg.Identities[name]
	if !ok {
		return nil // a reload removed it: nothing left to wait for
	}
	target := ""
	if id.Kind == config.IdentityGitHubApp {
		if owners := slices.Sorted(maps.Keys(id.Installations)); len(owners) > 0 {
			target = owners[0] + "/probe"
		}
	}
	if err := c.minter.Probe(ctx, name, id, target); err != nil {
		return err
	}
	if config.IsLinear(id.Kind) {
		if _, err := c.linearFor(cfg, name).Viewer(ctx); err != nil {
			return err
		}
	}
	return nil
}

// probeOutageNow makes the identity's probe due on the next pass.
func (c *Controller) probeOutageNow(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	o, ok := c.outages[name]
	if !ok {
		return fmt.Errorf("%w: no host outage for identity %q", domain.ErrConflict, name)
	}
	o.nextProbe = c.now()
	return nil
}

// clearOutages lifts the pause on pool, or on every pool when pool is "".
// An outage left with no pools is over.
func (c *Controller) clearOutages(pool string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for name, o := range c.outages {
		if pool == "" {
			delete(c.outages, name)
			continue
		}
		o.pools = slices.DeleteFunc(o.pools, func(p string) bool { return p == pool })
		if len(o.pools) == 0 {
			delete(c.outages, name)
		}
	}
}

// outageStatusLocked lists the paused pools for pool list. c.mu is held.
func (c *Controller) outageStatusLocked() map[string]rpc.PoolOutage {
	if len(c.outages) == 0 {
		return nil
	}
	out := map[string]rpc.PoolOutage{}
	for _, name := range slices.Sorted(maps.Keys(c.outages)) {
		o := c.outages[name]
		for _, p := range o.pools {
			if _, dup := out[p]; !dup {
				out[p] = rpc.PoolOutage{Identity: name, Reason: o.last, Since: o.since, NextProbe: o.nextProbe}
			}
		}
	}
	return out
}

// outageHealthLocked is the attention rows' view: one per identity,
// however many pools and steps it holds up. c.mu is held.
func (c *Controller) outageHealthLocked() []attention.Outage {
	var out []attention.Outage
	for _, name := range slices.Sorted(maps.Keys(c.outages)) {
		o := c.outages[name]
		out = append(out, attention.Outage{Identity: name, Pools: slices.Clone(o.pools), Why: o.last, Since: o.since})
	}
	return out
}
