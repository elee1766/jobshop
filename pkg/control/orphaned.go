package control

import (
	"context"
	"sort"

	"go.opentelemetry.io/otel/metric"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// Reasons a ticket is orphaned (the gauge's reason label).
const (
	orphanNoPool      = "no_pool"
	orphanUnknownPool = "unknown_pool"
)

// orphaned returns open, unclaimed tickets that no configured pool will
// claim: their pool is empty or not in the loaded config.
func (c *Controller) orphaned(ctx context.Context, project string, limit int) ([]store.Ticket, error) {
	pools := make([]string, 0, len(c.config().Pools))
	for name := range c.config().Pools {
		pools = append(pools, name)
	}
	sort.Strings(pools)
	return c.store.Orphaned(ctx, project, pools, limit)
}

func orphanReason(t store.Ticket) string {
	if t.Pool == "" {
		return orphanNoPool
	}
	return orphanUnknownPool
}

// reportOrphans runs each tick and after a reload. It warns once per
// ticket and reason, so a stuck ticket does not log every tick, and sets
// the jobshop.tickets.orphaned gauge.
func (c *Controller) reportOrphans() {
	ctx := c.ctx
	ts, err := c.orphaned(ctx, "", 0)
	if err != nil {
		if ctx.Err() == nil {
			c.log.Error("orphaned tickets", "err", err)
		}
		return
	}
	counts := map[string]int64{orphanNoPool: 0, orphanUnknownPool: 0}
	seen := map[string]string{}
	c.orphanMu.Lock()
	for _, t := range ts {
		r := orphanReason(t)
		counts[r]++
		seen[t.ID] = r
		if c.orphanWarned[t.ID] != r {
			c.log.Warn("ticket has no claimable pool", "ticket", t.ID, "project", t.Project,
				"pool", t.Pool, "reason", r)
		}
	}
	// Dropping fixed tickets means a ticket that becomes orphaned again
	// warns again.
	c.orphanWarned = seen
	c.orphanMu.Unlock()
	for r, n := range counts {
		c.metrics.Orphaned.Record(ctx, n, metric.WithAttributes(telemetry.AttrReason.String(r)))
	}
}
