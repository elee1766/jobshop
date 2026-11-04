package control

import (
	"context"
	"sync"
	"time"

	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// Ingest drain (#335): the controller moves spooled webhook deliveries
// into the inbox. The spool is its own SQLite file, written by `ingest
// serve` or the controller's Linear listener; the drain is its only
// deleter.

// drainEvery is the drain poll interval. The controller's own listener
// also wakes it, so this only bounds the delay for `ingest serve`
// writes, which come from another process.
const drainEvery = 5 * time.Second

// ingestDrain holds the controller's open spool and its wakeup.
type ingestDrain struct {
	mu   sync.Mutex // one drain at a time
	sp   *ingest.Spool
	wake chan struct{}
}

// spool opens the spool on first use and keeps it open.
func (c *Controller) spool(ctx context.Context) (*ingest.Spool, error) {
	c.ingest.mu.Lock()
	defer c.ingest.mu.Unlock()
	if c.ingest.sp != nil {
		return c.ingest.sp, nil
	}
	sp, err := ingest.OpenSpool(ctx, c.config().Ingest.SpoolPath(c.root))
	if err != nil {
		return nil, err
	}
	c.ingest.sp = sp
	return sp, nil
}

// closeSpool closes the spool if it was opened.
func (c *Controller) closeSpool() {
	c.ingest.mu.Lock()
	defer c.ingest.mu.Unlock()
	if c.ingest.sp != nil {
		_ = c.ingest.sp.Close()
		c.ingest.sp = nil
	}
}

// wakeDrain asks the drain loop for a pass now.
func (c *Controller) wakeDrain() {
	select {
	case c.ingest.wake <- struct{}{}:
	default:
	}
}

// drainOnce runs one drain until the spool is empty.
func (c *Controller) drainOnce(ctx context.Context) (ingest.DrainResult, error) {
	sp, err := c.spool(ctx)
	if err != nil {
		return ingest.DrainResult{}, err
	}
	c.ingest.mu.Lock()
	defer c.ingest.mu.Unlock()
	res, err := ingest.Drain(ctx, sp, c.store, c.now, c.log)
	if res.Moved > 0 || res.Bad > 0 {
		c.log.Info("ingest drain", "moved", res.Moved, "fresh", res.Fresh, "bad", res.Bad)
	}
	return res, err
}

// pumpDrain drains on a timer and on wakeups until the controller stops.
// stop closes the spool once every writer is done.
func (c *Controller) pumpDrain() {
	t := c.clk().NewTicker(drainEvery)
	defer t.Stop()
	for {
		if _, err := c.drainOnce(c.ctx); err != nil && c.ctx.Err() == nil {
			c.log.Warn("ingest drain", "err", err)
		}
		select {
		case <-c.ctx.Done():
			return
		case <-t.C():
		case <-c.ingest.wake:
		}
	}
}

// ingestDrainMethod is `jobshop ingest drain`: one drain pass through the
// admin socket, for running it on demand or from outside the controller.
func (c *Controller) ingestDrainMethod(ctx context.Context, sc scope, p rpc.IngestDrainParams) (any, error) {
	if err := requireAdmin(sc, "ingest drain"); err != nil {
		return nil, err
	}
	if p.Spool != "" {
		sp, err := ingest.OpenSpool(ctx, c.root.Path(p.Spool))
		if err != nil {
			return nil, err
		}
		defer func() { _ = sp.Close() }()
		res, err := ingest.Drain(ctx, sp, c.store, c.now, c.log)
		if err != nil {
			return nil, err
		}
		return rpc.IngestDrainResult(res), nil
	}
	res, err := c.drainOnce(ctx)
	if err != nil {
		return nil, err
	}
	return rpc.IngestDrainResult(res), nil
}
