package control

import (
	"context"
	"fmt"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// escalatePool moves a step to its retry_pool (JOB-59) when it has one
// and is not on it yet. The move is one way: nothing moves a step back,
// so a run never routes down after a failure. A held step is left alone
// (its holder may still be working); the next event after it reopens
// moves it.
func (c *Controller) escalatePool(ctx context.Context, cfg *config.Config, step *domain.Ticket, why string) {
	if step.Kind != domain.KindStep || step.Status == domain.StatusInProgress {
		return
	}
	sf, err := domain.DecodeFields[domain.StepFields](*step)
	if err != nil || sf.RetryPool == "" || step.Pool == sf.RetryPool {
		return
	}
	if _, ok := cfg.Pools[sf.RetryPool]; !ok {
		c.log.WarnContext(ctx, "retry pool not configured", "step", step.ID, "pool", sf.RetryPool)
		return
	}
	from, to := step.Pool, sf.RetryPool
	t, err := c.store.Update(ctx, "controller", step.ID, store.Update{
		Pool:       &to,
		IfRevision: step.Revision,
		Comment:    fmt.Sprintf("Moved from %s to retry pool %s (%s).", from, to, why),
	})
	if err != nil {
		c.log.WarnContext(ctx, "retry pool move", "step", step.ID, "err", err)
		return
	}
	*step = t
	c.log.InfoContext(ctx, "step moved to retry pool", "step", step.ID, "from", from, "to", to)
}

// escalateFailedSteps moves each open step of a run that has failed an
// attempt to its retry pool. Called on every run event, so a failed
// close, a release and an expired lease all land here.
func (c *Controller) escalateFailedSteps(ctx context.Context, cfg *config.Config, runID string) {
	steps, err := store.Children(ctx, c.store, runID)
	if err != nil {
		return
	}
	for i := range steps {
		s := &steps[i]
		if s.Kind != domain.KindStep || s.Status != domain.StatusOpen {
			continue
		}
		sf, err := domain.DecodeFields[domain.StepFields](*s)
		if err != nil || sf.Failures == 0 || sf.RetryPool == "" || s.Pool == sf.RetryPool {
			continue
		}
		c.escalatePool(ctx, cfg, s, fmt.Sprintf("attempt %d failed", sf.Failures))
	}
}
