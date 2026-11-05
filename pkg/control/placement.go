package control

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/store"
)

// Placement (#409, docs/notes/409-step-placement.md). A step with
// placement options starts on its top option. placeSteps runs every
// tick: once the step could start and prefer_for has passed, it moves the
// open, unclaimed step to the highest-weight option whose resource slots
// are free, back up when a higher one frees first.

// resolvePlacement fills a step's placement from run vars.
func resolvePlacement(cfg *config.Config, w *config.Workflow, p *config.Placement, vars map[string]any) (*store.StepPlacement, error) {
	opts, err := cfg.ResolvePlacement(w, p, vars)
	if err != nil {
		return nil, err
	}
	sp := &store.StepPlacement{PreferMS: p.PreferDuration().Milliseconds()}
	for _, o := range opts {
		sp.Options = append(sp.Options, store.PlacementOption{Pool: o.Pool, Resources: o.Resources, Weight: o.Weight})
	}
	return sp, nil
}

// placeSteps re-places open steps with placement options.
func (c *Controller) placeSteps() {
	ctx := c.ctx
	steps, err := c.store.List(ctx, store.Filter{Kind: store.KindStep, Status: store.StatusOpen, Limit: store.DefaultLimit})
	if err != nil {
		return
	}
	var free func(key string) bool
	for _, t := range steps {
		if t.Holder != "" {
			continue
		}
		sf, err := store.DecodeFields[store.StepFields](t)
		if err != nil || sf.Placement == nil || len(sf.Placement.Options) < 2 {
			continue
		}
		if free == nil {
			if free, err = c.slotsFree(ctx); err != nil {
				return
			}
		}
		if err := c.placeStep(ctx, t, sf, free); err != nil {
			c.log.Warn("placement", "step", t.ID, "err", err)
		}
	}
}

// slotsFree returns whether a resource key has a free slot now.
func (c *Controller) slotsFree(ctx context.Context) (func(string) bool, error) {
	rep, err := c.store.Resources(ctx, "")
	if err != nil {
		return nil, err
	}
	caps := map[string]int{}
	for name, n := range c.config().ResourceCaps() {
		caps[name] = n
	}
	held := map[string]int{}
	for _, u := range rep.Use {
		held[u.Key] = u.Held
	}
	return func(key string) bool {
		n, ok := caps[key]
		if !ok {
			n, ok = caps[store.ResourceName(key)]
		}
		return !ok || held[key] < n
	}, nil
}

func (c *Controller) placeStep(ctx context.Context, t store.Ticket, sf store.StepFields, free func(string) bool) error {
	now := c.now()
	p := sf.Placement
	if !t.DeferUntil.IsZero() && t.DeferUntil.After(now) {
		return nil
	}
	if p.Since.IsZero() {
		blocked, err := c.stepBlocked(ctx, t)
		if err != nil || blocked {
			return err
		}
		p.Since = now.UTC().Truncate(time.Millisecond)
	}
	target := 0
	if now.Sub(p.Since) >= time.Duration(p.PreferMS)*time.Millisecond {
		target = p.Chosen // none feasible: stay where it waits
		for i, o := range p.Options {
			if allFree(o.Resources, free) {
				target = i
				break
			}
		}
	}
	from := p.Chosen
	p.Chosen = target
	sf.Resources = p.Options[target].Resources
	b, err := json.Marshal(sf)
	if err != nil {
		return err
	}
	if string(b) == string(t.Fields) {
		return nil
	}
	u := store.Update{Fields: b, IfRevision: t.Revision}
	if target != from {
		pool := p.Options[target].Pool
		u.Pool = &pool
	}
	if _, err := c.store.Update(ctx, "controller", t.ID, u); err != nil {
		return err
	}
	if target != from {
		c.log.Info("step placed", "step", t.ID, "pool", p.Options[target].Pool, "fallback", target > 0)
		_, _ = store.AddComment(ctx, c.store, "controller", t.ID, fmt.Sprintf("Placement: moved to pool %s (option %d of %d, weight %d).",
			p.Options[target].Pool, target+1, len(p.Options), p.Options[target].Weight))

	}
	return nil
}

func allFree(keys []string, free func(string) bool) bool {
	for _, k := range keys {
		if !free(k) {
			return false
		}
	}
	return true
}

// stepBlocked reports whether a blocker of t is unfinished.
func (c *Controller) stepBlocked(ctx context.Context, t store.Ticket) (bool, error) {
	edges, err := store.Edges(ctx, c.store, t.ID)
	if err != nil {
		return false, err
	}
	for _, e := range edges {
		if e.From != t.ID || (e.Kind != store.EdgeBlocks && e.Kind != store.EdgeWaitsFor) {
			continue
		}
		b, err := c.store.Get(ctx, e.To)
		if err != nil {
			return false, err
		}
		if b.Status != store.StatusClosed || (b.Kind == store.KindStep && !store.StepFinished(b)) {
			return true, nil
		}
	}
	return false, nil
}
