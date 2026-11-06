package control

import (
	"context"
	"os"
	"sync"

	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/elee1766/jobshop/pkg/store"
)

// reapOrphans stops hosts whose session is not open: leftovers from a crash
// between spawn and record, or from a session closed while the controller
// was down.
func (c *Controller) reapOrphans(open []store.Ticket) {
	for _, id := range c.orphans(open) {
		c.log.Warn("stopping orphaned host", "session", id)
		c.stopHost(c.root.SessionDir(id))
	}
}

// orphans lists session directories with a live host whose session is not
// in open.
func (c *Controller) orphans(open []store.Ticket) []string {
	keep := map[string]bool{}
	for _, t := range open {
		keep[t.ID] = true
	}
	entries, err := os.ReadDir(c.root.SessionDir(""))
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if !keep[e.Name()] && host.Alive(c.root.SessionDir(e.Name())) {
			out = append(out, e.Name())
		}
	}
	return out
}

// stopAll is the full shutdown: end every session, then stop.
func (c *Controller) stopAll() {
	c.mu.Lock()
	c.drain = true
	c.mu.Unlock()
	ctx := context.WithoutCancel(c.ctx)
	open, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Status: store.StatusOpen})
	if err != nil {
		c.log.Error("shutdown: list sessions", "err", err)
	}
	var wg sync.WaitGroup
	for _, t := range open {
		wg.Go(func() {
			c.stopRunner(t.ID)
			if t2, err := store.Close(ctx, c.store, "controller", t.ID, nil, "shutdown"); err != nil {
				c.log.Error("shutdown: close session", "session", t.ID, "err", err)
			} else {
				c.countClose(ctx, t2)
			}
			c.endSession(ctx, t.ID, "controller shutdown")
		})
	}
	wg.Wait()
	c.log.Info("all sessions stopped", "count", len(open))
	c.cancel()
}
