package control

import (
	"context"
	"slices"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// Ready-work wakes (JOB-197). An idle runner used to ask the store for
// work every Options.Poll, so a new ticket waited up to a poll and an
// idle pool cost steady queries. Now a ticket that becomes claimable
// wakes its pool's idle runners; Poll is only the fallback for a wake
// nobody sent (a deferral running out, a provider slot turning ready).
//
// A pool's wake is a channel closed and replaced on each wake, so it
// reaches every runner waiting on it, however many. A runner takes the
// channel before it asks the store for work: a ticket committed after the
// ask closes the channel it holds, so no wake falls between the ask and
// the wait.

// readyChan is the channel the next wake of pool closes.
func (c *Controller) readyChan(pool string) <-chan struct{} {
	c.readyMu.Lock()
	defer c.readyMu.Unlock()
	ch, ok := c.readyWakes[pool]
	if !ok {
		if c.readyWakes == nil {
			c.readyWakes = map[string]chan struct{}{}
		}
		ch = make(chan struct{})
		c.readyWakes[pool] = ch
	}
	return ch
}

// wakePool ends the idle wait of every runner of pool.
func (c *Controller) wakePool(pool string) {
	c.readyMu.Lock()
	defer c.readyMu.Unlock()
	if ch, ok := c.readyWakes[pool]; ok {
		close(ch)
		delete(c.readyWakes, pool)
	}
}

// wakeAllPools wakes every pool. A close can unblock a ticket of another
// pool, and a reload can change resource caps, so those wake them all.
func (c *Controller) wakeAllPools() {
	c.readyMu.Lock()
	defer c.readyMu.Unlock()
	for pool, ch := range c.readyWakes {
		close(ch)
		delete(c.readyWakes, pool)
	}
}

// readyOps are the history ops that can make a ticket claimable: a new
// ticket, one handed back (release, handoff, an expired lease), reopened,
// or changed (a wait cleared, a deferral lifted, a pool or holder set),
// or one whose blocker edge went.
var readyOps = []string{"create", "release", "handoff", "lease_expired", "reopen", "update", "edge_remove"}

// noteChange wakes who a committed history row can give work to. t is
// the ticket now, nil if it is gone. It runs in the events pump, which
// reads every committed row once, so a write anywhere wakes the runners
// without its writer knowing about them.
func (c *Controller) noteChange(h store.HistoryEntry, t *domain.Ticket) {
	if t == nil {
		return
	}
	switch t.Kind {
	case domain.KindMessage:
		if h.Op == "create" {
			c.wakeMessage(t)
		}
	case domain.KindSession:
		// Never claimed, and closing one frees no slot.
	default:
		switch {
		case h.Op == "close":
			// Closing a blocker readies its dependents, closing a holder
			// frees its resource slots: either can be in another pool.
			c.wakeAllPools()
		case slices.Contains(readyOps, h.Op) && t.Status == domain.StatusOpen && t.Pool != "":
			c.wakePool(t.Pool)
		}
		// Work pre-assigned to a session is claimed by that session whatever
		// the ticket's pool.
		if t.Holder != "" && t.Status == domain.StatusOpen {
			c.mu.Lock()
			r := c.runners[t.Holder]
			c.mu.Unlock()
			if r != nil {
				c.wakePool(r.pool)
			}
		}
	}
}

// wakeMessage wakes the idle runners a new message is for, so it is
// delivered now and not at the next poll: a pool message wakes the pool,
// one to a session id wakes that session's pool, one to a key (which
// pool holds it is not known here) wakes every pool.
func (c *Controller) wakeMessage(t *domain.Ticket) {
	f, err := domain.DecodeFields[domain.MessageFields](*t)
	if err != nil || f.To == "" {
		return
	}
	if pool, ok := cutPoolAddr(f.To); ok {
		c.wakePool(pool)
		return
	}
	c.mu.Lock()
	r := c.runners[f.To]
	c.mu.Unlock()
	if r != nil {
		c.wakePool(r.pool)
		return
	}
	c.wakeAllPools()
}

// cutPoolAddr is the pool of a "pool:<name>" message address.
func cutPoolAddr(to string) (string, bool) {
	const prefix = "pool:"
	if len(to) > len(prefix) && to[:len(prefix)] == prefix {
		return to[len(prefix):], true
	}
	return "", false
}

// heldSet is the holders of every in-progress ticket, in one query: a
// session is working when it is in the set, idle when it is not. The
// gauges, the scale plan and the heartbeats read it once per pass instead
// of asking per session (JOB-197).
func (c *Controller) heldSet(ctx context.Context) (map[string]bool, error) {
	ts, err := c.inProgress(ctx)
	if err != nil {
		return nil, err
	}
	return holdersOf(ts), nil
}

// inProgress lists every in-progress ticket.
func (c *Controller) inProgress(ctx context.Context) ([]domain.Ticket, error) {
	return c.store.List(ctx, store.Filter{Status: domain.StatusInProgress, Limit: store.NoLimit})
}

func holdersOf(ts []domain.Ticket) map[string]bool {
	held := make(map[string]bool, len(ts))
	for _, t := range ts {
		if t.Holder != "" {
			held[t.Holder] = true
		}
	}
	return held
}
