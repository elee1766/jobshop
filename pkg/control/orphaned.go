package control

import (
	"context"
	"fmt"
	"sort"
	"time"

	"go.opentelemetry.io/otel/metric"

	"github.com/elee1766/jobshop/pkg/rpc"
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
	ts, err := c.store.List(ctx, store.Filter{
		Project: project, Kinds: store.ClaimableKinds, Status: store.StatusOpen, Unheld: true,
		PoolNotIn: pools, Limit: limit,
	})
	if err != nil {
		return nil, err
	}
	// Command and http steps have no pool on purpose: the controller
	// runs them (#396).
	out := ts[:0]
	for _, t := range ts {
		if f, err := store.DecodeFields[store.StepFields](t); t.Kind == store.KindStep && err == nil && f.Auto() {
			continue
		}
		out = append(out, t)
	}
	return out, nil
}

func orphanReason(t store.Ticket) string {
	if t.Pool == "" {
		return orphanNoPool
	}
	return orphanUnknownPool
}

// staleAttention lists orphaned tickets older than attention.orphan_age
// (#495): ad-hoc tickets nobody will claim otherwise sit in ready forever.
// The title carries the command that closes one.
func (c *Controller) staleAttention(ctx context.Context) ([]rpc.AttentionItem, error) {
	age := c.config().Attention.OrphanAgeDuration()
	if age == 0 {
		return nil, nil
	}
	ts, err := c.orphaned(ctx, "", 0)
	if err != nil {
		return nil, err
	}
	now := c.now()
	var out []rpc.AttentionItem
	for _, t := range ts {
		if now.Sub(t.CreatedAt) < age {
			continue
		}
		it := ticketItem(rpc.AttentionStale, t)
		it.Title = fmt.Sprintf("%s: no claimable pool (%s) for %s; if fixed: jobshop close %s -r cancelled",
			t.Title, orphanReason(t), now.Sub(t.CreatedAt).Truncate(time.Hour), t.ID)
		out = append(out, it)
	}
	return out, nil
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
