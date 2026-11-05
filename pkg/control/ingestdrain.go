package control

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Ingest queue (#335, #679, docs/notes/679-ingest-queue.md). Every
// source takes one path: a receiver appends to the spool (`ingest
// serve`, the controller's Linear listener, `ingest put`) or a poller
// appends to the inbox; the drain moves the spool into the inbox. Every
// inbox append goes through inboxAppended, which advances the inbox
// notifier and wakes the waits the new items name. The admission pump
// and external workers (`ingest.next`) block on the notifier. Timers
// stay as a reconcile backstop.

// drainEvery is the drain timer. Receivers wake the drain on append:
// the controller's own listener directly, `ingest serve` over the
// admin socket (ingest.notify), so this only bounds the delay when the
// controller was unreachable at append time.
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

// spoolExists reports whether the spool is open or its file exists.
func (c *Controller) spoolExists() bool {
	c.ingest.mu.Lock()
	open := c.ingest.sp != nil
	c.ingest.mu.Unlock()
	if open {
		return true
	}
	_, err := os.Stat(c.config().Ingest.SpoolPath(c.root))
	return err == nil
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

// inboxNotifier is the inbox change signal, made on first use so test
// controllers built as struct literals have one.
func (c *Controller) inboxNotifier() *ingest.Notifier {
	c.inboxOnce.Do(func() {
		if c.inbox == nil {
			c.inbox = ingest.NewNotifier()
		}
	})
	return c.inbox
}

// inboxAppended is the one hook after any inbox append, from any source
// (#679): it wakes the inbox's consumers and the waits whose conditions
// name a new item's subject. Items are the new ones only; dedupes do not
// call it.
func (c *Controller) inboxAppended(items []store.InboxItem) {
	if len(items) == 0 {
		return
	}
	c.inboxNotifier().Advance()
	c.wakeWaitsFor(items)
}

// drainOnce runs one drain until the spool is empty.
func (c *Controller) drainOnce(ctx context.Context) (ingest.DrainResult, error) {
	// No spool file means nothing was ever spooled. Opening it would
	// create and migrate it, which every test controller paid (#433).
	// Writers (the Linear listener, `ingest serve`) create it.
	if !c.spoolExists() {
		return ingest.DrainResult{}, nil
	}
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
	c.inboxAppended(res.Items)
	c.linearDelivered(res.Items)
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
		c.inboxAppended(res.Items)
		c.linearDelivered(res.Items)
		if err != nil {
			return nil, err
		}
		return drainResult(res), nil
	}
	res, err := c.drainOnce(ctx)
	if err != nil {
		return nil, err
	}
	return drainResult(res), nil
}

func drainResult(r ingest.DrainResult) rpc.IngestDrainResult {
	return rpc.IngestDrainResult{Moved: r.Moved, Fresh: r.Fresh, Bad: r.Bad}
}

// ingestNotify is a receiver's append signal (#679): `ingest serve`
// calls it after each new spool row, so the drain runs now.
func (c *Controller) ingestNotify(_ context.Context, sc scope, _ struct{}) (any, error) {
	if err := requireAdmin(sc, "ingest notify"); err != nil {
		return nil, err
	}
	c.wakeDrain()
	return nil, nil
}

// ingestNextMax caps one ingest.next wait below the CLI's 30s call
// deadline.
const ingestNextMax = 25 * time.Second

// ingestNext is the inbox long poll (#679): it blocks until the inbox
// changed since p.After (an append or an admission), the timeout, or
// controller stop. External workers run a pass after each return.
func (c *Controller) ingestNext(ctx context.Context, sc scope, p rpc.IngestNextParams) (any, error) {
	if err := requireAdmin(sc, "ingest next"); err != nil {
		return nil, err
	}
	d := time.Duration(p.TimeoutMS) * time.Millisecond
	if d <= 0 || d > ingestNextMax {
		d = ingestNextMax
	}
	wctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		select {
		case <-c.after(d):
			cancel()
		case <-wctx.Done():
		}
	}()
	seq := c.inboxNotifier().Wait(p.After, wctx.Done())
	return rpc.IngestNextResult{Seq: seq, Changed: seq != p.After}, nil
}
