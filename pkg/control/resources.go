package control

import (
	"context"
	"fmt"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// resourceWaitSince is when t started waiting: its last activity, or
// when its last blocker closed, whichever is later (#408).
func (c *Controller) resourceWaitSince(ctx context.Context, t store.Ticket) (time.Time, error) {
	last, err := c.lastActivity(ctx, t)
	if err != nil {
		return last, err
	}
	return c.lastUnblocked(ctx, t, last)
}

// waitText is how run show and stalls describe a wait.
func waitText(w store.ResourceWait) string {
	return fmt.Sprintf("waiting on %s, %d ahead (%d/%d held)", w.Key, w.Ahead, w.Held, w.Capacity)
}

// stallResourceWait: tickets held back only by a full resource for
// after (#408). They are out of the ready frontier, so ready_unclaimed
// does not see them.
func (c *Controller) stallResourceWait(ctx context.Context, now time.Time, after time.Duration) ([]stall, error) {
	waits, err := c.store.ResourceWaits(ctx)
	if err != nil {
		return nil, err
	}
	var out []stall
	for _, w := range waits {
		t, err := c.store.Get(ctx, w.Ticket)
		if err != nil {
			return nil, err
		}
		since, err := c.resourceWaitSince(ctx, t)
		if err != nil {
			return nil, err
		}
		if now.Sub(since) >= after {
			out = append(out, stall{config.StallResourceWait, t.ID, fmt.Sprintf("%s for %s", waitText(w), ago(now, since))})
		}
	}
	return out, nil
}

// resourceAttention lists resource waits longer than
// attention.resource_wait (#408). Off when unset.
func (c *Controller) resourceAttention(ctx context.Context) ([]rpc.AttentionItem, error) {
	after := config.Duration(c.config().Attention.ResourceWait, 0)
	if after == 0 {
		return nil, nil
	}
	waits, err := c.store.ResourceWaits(ctx)
	if err != nil {
		return nil, err
	}
	now := c.now()
	var out []rpc.AttentionItem
	for _, w := range waits {
		t, err := c.store.Get(ctx, w.Ticket)
		if err != nil {
			return nil, err
		}
		since, err := c.resourceWaitSince(ctx, t)
		if err != nil {
			return nil, err
		}
		if now.Sub(since) < after {
			continue
		}
		it := ticketItem(rpc.AttentionResource, t)
		it.Title = t.Title + ": " + waitText(w)
		it.Since = since
		out = append(out, it)
	}
	return out, nil
}

// runWaits are the resource waits of a run's steps.
func (c *Controller) runWaits(ctx context.Context, steps []store.Ticket) ([]store.ResourceWait, error) {
	mine := map[string]bool{}
	for _, s := range steps {
		if s.Status == store.StatusOpen {
			mine[s.ID] = true
		}
	}
	if len(mine) == 0 {
		return nil, nil
	}
	waits, err := c.store.ResourceWaits(ctx)
	if err != nil {
		return nil, err
	}
	var out []store.ResourceWait
	for _, w := range waits {
		if mine[w.Ticket] {
			out = append(out, w)
		}
	}
	return out, nil
}
