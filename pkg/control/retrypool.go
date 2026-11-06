package control

import (
	"context"
	"fmt"

	"github.com/elee1766/jobshop/pkg/store"
)

// escalatePool moves a step to its retry_pool (JOB-59) when it has one
// and is not on it yet. The move is one way: nothing moves a step back,
// so a run never routes down after a failure. A held step is left alone
// (its holder may still be working); the next event after it reopens
// moves it.
func (c *Controller) escalatePool(ctx context.Context, step *store.Ticket, why string) {
	if step.Kind != store.KindStep || step.Status == store.StatusInProgress {
		return
	}
	sf, err := store.DecodeFields[store.StepFields](*step)
	if err != nil || sf.RetryPool == "" || step.Pool == sf.RetryPool {
		return
	}
	if _, ok := c.config().Pools[sf.RetryPool]; !ok {
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
func (c *Controller) escalateFailedSteps(ctx context.Context, runID string) {
	steps, err := store.Children(ctx, c.store, runID)
	if err != nil {
		return
	}
	for i := range steps {
		s := &steps[i]
		if s.Kind != store.KindStep || s.Status != store.StatusOpen {
			continue
		}
		sf, err := store.DecodeFields[store.StepFields](*s)
		if err != nil || sf.Failures == 0 || sf.RetryPool == "" || s.Pool == sf.RetryPool {
			continue
		}
		c.escalatePool(ctx, s, fmt.Sprintf("attempt %d failed", sf.Failures))
	}
}
