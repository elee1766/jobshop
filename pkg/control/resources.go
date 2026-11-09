package control

import (
	"context"
	"fmt"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/triggers"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// resourceWaitSince is when t started waiting: its last activity, or
// when its last blocker closed, whichever is later (#408).
func (c *Controller) resourceWaitSince(ctx context.Context, t domain.Ticket) (time.Time, error) {
	last, err := triggers.LastActivity(ctx, c.store, t)
	if err != nil {
		return last, err
	}
	return triggers.LastUnblocked(ctx, c.store, t, last)
}

// waitText is how run show and findings describe a wait.
func waitText(w domain.ResourceWait) string {
	return fmt.Sprintf("waiting on %s, %d ahead (%d/%d held)", w.Key, w.Ahead, w.Held, w.Capacity)
}

// signalResourceWait: tickets held back only by a full resource for
// after (#408). They are out of the ready frontier, so ready_unclaimed
// does not see them.
func (c *Controller) signalResourceWait(ctx context.Context, now time.Time, after time.Duration) ([]triggers.Finding, error) {
	rep, err := c.store.Resources(ctx, "")
	if err != nil {
		return nil, err
	}
	var out []triggers.Finding
	for _, w := range rep.Waits {
		t, err := c.store.Get(ctx, w.Ticket)
		if err != nil {
			return nil, err
		}
		since, err := c.resourceWaitSince(ctx, t)
		if err != nil {
			return nil, err
		}
		if now.Sub(since) >= after {
			out = append(out, triggers.Finding{Signal: config.SignalResourceWait, Subject: t.ID, Detail: fmt.Sprintf("%s for %s", waitText(w), triggers.Ago(now, since))})
		}
	}
	return out, nil
}

// resourceAttention lists resource waits longer than
// attention.resource_wait (#408). Off when unset.
func (c *Controller) resourceAttention(ctx context.Context, cfg *config.Config) ([]rpc.AttentionItem, error) {
	after := config.Duration(cfg.Attention.ResourceWait, 0)
	if after == 0 {
		return nil, nil
	}
	rep, err := c.store.Resources(ctx, "")
	if err != nil {
		return nil, err
	}
	now := c.now()
	var out []rpc.AttentionItem
	for _, w := range rep.Waits {
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
func (c *Controller) runWaits(ctx context.Context, steps []domain.Ticket) ([]domain.ResourceWait, error) {
	mine := map[string]bool{}
	for _, s := range steps {
		if s.Status == domain.StatusOpen {
			mine[s.ID] = true
		}
	}
	if len(mine) == 0 {
		return nil, nil
	}
	rep, err := c.store.Resources(ctx, "")
	if err != nil {
		return nil, err
	}
	var out []domain.ResourceWait
	for _, w := range rep.Waits {
		if mine[w.Ticket] {
			out = append(out, w)
		}
	}
	return out, nil
}
