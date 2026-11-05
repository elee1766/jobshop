package control

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Provider-owned resource slots (#407). The store keeps slot state and
// takes a slot in the claim transaction; this file is the provider side:
// the admin RPC a provider calls to add, remove and fail slots, and the
// supervisor pass that resets slots after claims, sends demand and asks
// for idle slots to be removed.

// providerCall is the JSON a provider gets on stdin or as the POST body.
type providerCall struct {
	Op       string            `json:"op"` // reset, demand, remove
	Resource string            `json:"resource"`
	Slot     string            `json:"slot,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
	Demand   *store.SlotDemand `json:"demand,omitempty"`
}

// providerState is tick-only state for demand and idle release.
type providerState struct {
	demandAt   map[string]time.Time
	demandLast map[string]store.SlotDemand
	// removing marks slots already asked to be removed, so a slow
	// provider is not asked every pass.
	removing map[string]time.Time
}

// callProvider runs one provider call and reports whether it succeeded.
func (c *Controller) callProvider(ctx context.Context, p *config.Provider, call providerCall) error {
	ctx, cancel := context.WithTimeout(ctx, p.TimeoutDuration())
	defer cancel()
	body, err := json.Marshal(call)
	if err != nil {
		return err
	}
	if p.URL != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.URL, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		_ = resp.Body.Close()
		if resp.StatusCode/100 != 2 {
			return fmt.Errorf("provider %s: %s", call.Op, resp.Status)
		}
		return nil
	}
	args := []string{"-c", p.Command + ` "$@"`, "provider", call.Op}
	if call.Slot != "" {
		args = append(args, call.Slot)
	}
	cmd := exec.CommandContext(ctx, "sh", args...)
	cmd.Dir = c.root.Dir
	cmd.Stdin = bytes.NewReader(body)
	cmd.Env = append(hostInheritedEnv(), "JOBSHOP_RESOURCE="+call.Resource, "JOBSHOP_SLOT="+call.Slot)
	var errOut limitedBuffer
	cmd.Stderr = &errOut
	cmd.WaitDelay = 5 * time.Second
	if err := cmd.Run(); err != nil {
		if tail := strings.TrimSpace(errOut.String()); tail != "" {
			return fmt.Errorf("%w: %s", err, lastLines(tail, 5))
		}
		return err
	}
	return nil
}

// providerPass is one supervisor pass over provider resources.
func (c *Controller) providerPass() {
	ctx := c.ctx
	cfg := c.config()
	if len(cfg.ResourceProviders()) == 0 {
		return
	}
	if c.prov.demandAt == nil {
		c.prov = providerState{demandAt: map[string]time.Time{}, demandLast: map[string]store.SlotDemand{}, removing: map[string]time.Time{}}
	}
	c.resetSlots(ctx, cfg)
	now := c.now()
	rep, err := c.store.Resources(ctx, "")
	if err != nil {
		return
	}
	slots := rep.Slots
	// Forget remove marks of slots that are gone or no longer idle.
	idleNow := map[string]bool{}
	for _, sl := range slots {
		if sl.State == store.SlotReady {
			idleNow[sl.Pool+"/"+sl.ID] = true
		}
	}
	maps.DeleteFunc(c.prov.removing, func(k string, _ time.Time) bool { return !idleNow[k] })
	for _, name := range cfg.ResourceProviders() {
		p := cfg.Resources[name].Provider
		prep, err := c.store.Resources(ctx, name)
		if err != nil {
			continue
		}
		d := prep.Demand
		if last, ok := c.prov.demandLast[name]; !ok || last != d || now.Sub(c.prov.demandAt[name]) >= p.DemandEveryDuration() {
			c.prov.demandAt[name], c.prov.demandLast[name] = now, d
			c.wg.Go(func() { c.providerAsync(p, providerCall{Op: "demand", Resource: name, Demand: &d}) })
		}
		idle := p.IdleReleaseDuration()
		if idle <= 0 || d.Waiting > 0 {
			continue
		}
		for _, sl := range slots {
			k := sl.Pool + "/" + sl.ID
			if sl.Pool != name || sl.State != store.SlotReady || now.Sub(sl.UpdatedAt) < idle {
				continue
			}
			if at, ok := c.prov.removing[k]; ok && now.Sub(at) < idle {
				continue
			}
			c.prov.removing[k] = now
			c.wg.Go(func() { c.providerAsync(p, providerCall{Op: "remove", Resource: name, Slot: sl.ID, Env: sl.Env}) })
		}
	}
}

// providerAsync is a fire-and-forget provider call (demand, remove): the
// provider answers a remove with `resource remove`.
func (c *Controller) providerAsync(p *config.Provider, call providerCall) {
	if err := c.callProvider(c.ctx, p, call); err != nil && c.ctx.Err() == nil {
		c.log.Warn("resource provider", "resource", call.Resource, "op", call.Op, "slot", call.Slot, "err", err)
	}
}

// resetSlots asks the provider to reset each slot whose claim ended,
// once: ok -> ready, failure -> gone. Calls run in this pass so a slot
// is never reset twice; providers bound them with timeout.
func (c *Controller) resetSlots(ctx context.Context, cfg *config.Config) {
	rel, err := c.store.ReleaseSlots(ctx)
	if err != nil {
		if ctx.Err() == nil {
			c.log.Error("release slots", "err", err)
		}
		return
	}
	for _, sl := range rel {
		r, ok := cfg.Resources[sl.Pool]
		if !ok || r.Provider == nil || sl.Removing {
			// No provider any more, or removed while held: nothing to
			// reset. SlotReset deletes a removing slot.
			_, _, _ = c.store.Slot(ctx, "", store.SlotEvent{Op: store.SlotReset, Pool: sl.Pool, ID: sl.ID, OK: !ok || r.Provider == nil})
			continue
		}
		err := c.callProvider(ctx, r.Provider, providerCall{Op: "reset", Resource: sl.Pool, Slot: sl.ID, Env: sl.Env})
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			c.log.Warn("resource reset failed: slot gone", "resource", sl.Pool, "slot", sl.ID, "err", err)
		}
		if _, _, serr := c.store.Slot(ctx, "", store.SlotEvent{Op: store.SlotReset, Pool: sl.Pool, ID: sl.ID, OK: err == nil}); serr != nil {
			c.log.Error("record slot reset", "resource", sl.Pool, "slot", sl.ID, "err", serr)
		}
	}
}

// slotEnv is the env of the slots t's claim holds, sorted, as KEY=value.
func (c *Controller) slotEnv(ctx context.Context, t store.Ticket) []string {
	rep, err := c.store.Resources(ctx, "")
	if err != nil {
		return nil
	}
	var out []string
	for _, sl := range store.TicketSlots(rep, t) {
		out = append(out, "JOBSHOP_SLOT_"+strings.ToUpper(sl.Pool)+"="+sl.ID)
		for _, k := range slices.Sorted(maps.Keys(sl.Env)) {
			out = append(out, k+"="+sl.Env[k])
		}
	}
	return out
}

// slotPrompt tells an agent which slots it holds and their env. A host
// outlives claims, so the env goes in the prompt, not the process.
func slotPrompt(env []string) string {
	if len(env) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("# Resource slots\n\nThis claim holds resource slots. Export this env in the shell you use for the work:\n\n```sh\n")
	for _, e := range env {
		k, v, _ := strings.Cut(e, "=")
		fmt.Fprintf(&b, "export %s=%s\n", k, shellQuote(v))
	}
	b.WriteString("```\n\n")
	return b.String()
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// providerResource checks name is a provider resource in the live config.
func (c *Controller) providerResource(name string) error {
	r, ok := c.config().Resources[name]
	if !ok || r.Provider == nil {
		return fmt.Errorf("%w: %q is not a resource with a provider", store.ErrInvalid, name)
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
	sl, _, err := c.store.Slot(ctx, sc.actor, store.SlotEvent{Op: store.SlotAdd, Pool: p.Resource, ID: p.Slot, Env: p.Env})
	return sl, err
}

func (c *Controller) resourceRemove(ctx context.Context, sc scope, p rpc.SlotParams) (any, error) {
	if err := c.requireGrant(sc, rpc.MethodResourceRemove, "resource remove"); err != nil {
		return nil, err
	}
	sl, _, err := c.store.Slot(ctx, sc.actor, store.SlotEvent{Op: store.SlotRemove, Pool: p.Resource, ID: p.Slot})
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
	sl, reopened, err := c.store.Slot(ctx, sc.actor, store.SlotEvent{Op: store.SlotUnhealthy, Pool: p.Resource, ID: p.Slot, Reason: reason})
	if err != nil {
		return nil, err
	}
	if reopened != "" {
		c.log.Info("resource unhealthy: claim requeued", "resource", p.Resource, "slot", p.Slot, "ticket", reopened)
		// The holder's session must stop working the requeued ticket.
		c.stepEvents(ctx, []string{reopened})
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
