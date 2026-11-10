package control

import (
	"context"
	"time"

	"github.com/gfx-labs/jobshop/pkg/config"
)

// storeMaintainEvery spaces planner statistics refreshes. Open ran one;
// this one picks up a collection that has grown since. Nothing changed
// much means a few microseconds, so the interval only bounds how long a
// grown table runs on old statistics.
const storeMaintainEvery = 6 * time.Hour

// maintainStore refreshes the store's planner statistics (JOB-191).
func (c *Controller) maintainStore(ctx context.Context, _ *config.Config) {
	now := c.now()
	if c.storeMaintainedAt.IsZero() {
		// Open just did it.
		c.storeMaintainedAt = now
		return
	}
	if now.Sub(c.storeMaintainedAt) < storeMaintainEvery {
		return
	}
	c.storeMaintainedAt = now
	if err := c.store.Maintain(ctx); err != nil && ctx.Err() == nil {
		c.log.ErrorContext(ctx, "store maintenance", "err", err)
	}
}
