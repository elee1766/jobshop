package control

import (
	"context"
	"fmt"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/slotprovider"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// Provider-owned resource slots (#407). The store keeps slot state and
// takes a slot in the claim transaction; the provider calls, demand and
// idle release live in pkg/control/slotprovider (JOB-158). This file is
// what stays in the controller: the admin RPC a provider calls to add,
// remove and fail slots, and the pass step that resets slots after claims.

// slots is the provider client, made on first use so controllers built
// without Run get one too.
func (c *Controller) slots() *slotprovider.Client {
	c.slotsOnce.Do(func() {
		c.slotClient = slotprovider.New(slotHost{c}, c.log)
	})
	return c.slotClient
}

// slotHost is the controller as slotprovider.Host.
type slotHost struct{ c *Controller }

func (h slotHost) Now() time.Time           { return h.c.now() }
func (h slotHost) Go(f func())              { h.c.wg.Go(f) }
func (h slotHost) Context() context.Context { return h.c.ctx }
func (h slotHost) Dir() string              { return h.c.root.Dir }
func (h slotHost) Env() []string            { return hostInheritedEnv() }

func (h slotHost) Resources(ctx context.Context, pool string) (domain.ResourceReport, error) {
	return h.c.store.Resources(ctx, pool)
}

// providerPass is one supervisor pass over provider resources.
func (c *Controller) providerPass(ctx context.Context, cfg *config.Config) {
	if len(cfg.ResourceProviders()) == 0 {
		return
	}
	c.resetSlots(ctx, cfg)
	c.slots().Pass(ctx, cfg)
}

// resetSlots asks the provider to reset each slot whose claim ended,
// once: ok -> ready, failure -> gone. The call runs outside the tick
// (JOB-103), one per slot; the pass that finds it finished records it,
// so a slot is never reset twice. Providers bound calls with timeout.
func (c *Controller) resetSlots(ctx context.Context, cfg *config.Config) {
	rel, err := c.store.ReleaseSlots(ctx)
	if err != nil {
		if ctx.Err() == nil {
			c.log.ErrorContext(ctx, "release slots", "err", err)
		}
		return
	}
	for _, sl := range rel {
		r, ok := cfg.Resources[sl.Pool]
		if !ok || r.Provider == nil || sl.Removing {
			// No provider any more, or removed while held: nothing to
			// reset. SlotReset deletes a removing slot.
			_, _, _ = c.store.Slot(ctx, "", domain.SlotEvent{Op: domain.SlotReset, Pool: sl.Pool, ID: sl.ID, OK: !ok || r.Provider == nil})
			continue
		}
		p := r.Provider
		key := "slot\x00" + sl.Pool + "\x00" + sl.ID
		v, done := c.offTake(key)
		if !done {
			call := slotprovider.Call{Op: "reset", Resource: sl.Pool, Slot: sl.ID, Env: sl.Env}
			c.offGo(key, "provider", p.TimeoutDuration(), true, func(ctx context.Context) any {
				return c.slots().Call(ctx, p, call)
			})
			if v, done = c.offTake(key); !done {
				continue // recorded by a later pass
			}
		}
		if ctx.Err() != nil {
			return
		}
		err, _ := v.(error)
		if err != nil {
			c.log.WarnContext(ctx, "resource reset failed: slot gone", "resource", sl.Pool, "slot", sl.ID, "err", err)
		}
		if _, _, serr := c.store.Slot(ctx, "", domain.SlotEvent{Op: domain.SlotReset, Pool: sl.Pool, ID: sl.ID, OK: err == nil}); serr != nil {
			c.log.ErrorContext(ctx, "record slot reset", "resource", sl.Pool, "slot", sl.ID, "err", serr)
		}
	}
}

// slotEnv is the env of the slots t's claim holds, sorted, as KEY=value.
func (c *Controller) slotEnv(ctx context.Context, t domain.Ticket) []string {
	rep, err := c.store.Resources(ctx, "")
	if err != nil {
		return nil
	}
	return slotprovider.Env(rep, t)
}

// providerResource checks name is a provider resource in the live config.
func (c *Controller) providerResource(name string) error {
	r, ok := c.config().Resources[name]
	if !ok || r.Provider == nil {
		return fmt.Errorf("%w: %q is not a resource with a provider", domain.ErrInvalid, name)
	}
	return nil
}

func (c *Controller) resourceAdd(ctx context.Context, sc scope, p rpc.SlotParams) (any, error) {
	if err := c.requireGrant(sc, rpc.MethodResourceAdd, "resource add"); err != nil {
		return nil, err
	}
	if err := c.providerResource(p.Resource); err != nil {
		return nil, err
	}
	sl, _, err := c.store.Slot(ctx, sc.actor, domain.SlotEvent{Op: domain.SlotAdd, Pool: p.Resource, ID: p.Slot, Env: p.Env})
	return sl, err
}

func (c *Controller) resourceRemove(ctx context.Context, sc scope, p rpc.SlotParams) (any, error) {
	if err := c.requireGrant(sc, rpc.MethodResourceRemove, "resource remove"); err != nil {
		return nil, err
	}
	sl, _, err := c.store.Slot(ctx, sc.actor, domain.SlotEvent{Op: domain.SlotRemove, Pool: p.Resource, ID: p.Slot})
	return sl, err
}

func (c *Controller) resourceUnhealthy(ctx context.Context, sc scope, p rpc.SlotParams) (any, error) {
	if err := c.requireGrant(sc, rpc.MethodResourceUnhealthy, "resource unhealthy"); err != nil {
		return nil, err
	}
	reason := p.Reason
	if reason == "" {
		reason = "reported by provider"
	}
	sl, reopened, err := c.store.Slot(ctx, sc.actor, domain.SlotEvent{Op: domain.SlotUnhealthy, Pool: p.Resource, ID: p.Slot, Reason: reason})
	if err != nil {
		return nil, err
	}
	if reopened != "" {
		c.log.InfoContext(ctx, "resource unhealthy: claim requeued", "resource", p.Resource, "slot", p.Slot, "ticket", reopened)
		// The holder's session must stop working the requeued ticket.
		c.stepEvents(ctx, c.config(), []string{reopened})
	}
	return rpc.SlotResult{Slot: sl, Requeued: reopened}, nil
}

func (c *Controller) resourceList(ctx context.Context, sc scope, _ struct{}) (any, error) {
	if err := c.requireGrant(sc, rpc.MethodResourceList, "resource list"); err != nil {
		return nil, err
	}
	rep, err := c.store.Resources(ctx, "")
	if err != nil {
		return nil, err
	}
	return rpc.ResourceListResult{Slots: rep.Slots, Use: rep.Use}, nil
}
