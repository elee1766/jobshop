package control

import (
	"context"
	"time"

	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// Ingest drain RPCs (#335, #679, docs/notes/679-ingest-queue.md). The
// spool, the drain loop and the notifier are pkg/control/ingestqueue;
// these handlers check the caller and call it.

// pumpDrain drains on a timer and on wakeups until the controller stops.
// stop closes the spool once every writer is done.
func (c *Controller) pumpDrain() { c.queue().RunDrain(c.ctx) }

// ingestDrainMethod is `jobshop ingest drain`: one drain pass through the
// admin socket, for running it on demand or from outside the controller.
func (c *Controller) ingestDrainMethod(ctx context.Context, sc scope, p rpc.IngestDrainParams) (any, error) {
	if err := requireAdmin(sc, "ingest drain"); err != nil {
		return nil, err
	}
	var res ingest.DrainResult
	var err error
	if p.Spool != "" {
		res, err = c.queue().DrainFile(ctx, c.root.Path(p.Spool))
	} else {
		res, err = c.queue().DrainOnce(ctx)
	}
	if err != nil {
		return nil, err
	}
	return rpc.IngestDrainResult{Moved: res.Moved, Fresh: res.Fresh, Bad: res.Bad}, nil
}

// ingestNotify is a receiver's append signal (#679): `ingest serve`
// calls it after each new spool row, so the drain runs now.
func (c *Controller) ingestNotify(_ context.Context, sc scope, _ struct{}) (any, error) {
	if err := requireAdmin(sc, "ingest notify"); err != nil {
		return nil, err
	}
	c.queue().WakeDrain()
	return nil, nil
}

// ingestNext is the inbox long poll (#679): it blocks until the inbox
// changed since p.After (an append or an admission), the timeout, or
// controller stop. External workers run a pass after each return.
func (c *Controller) ingestNext(ctx context.Context, sc scope, p rpc.IngestNextParams) (any, error) {
	if err := requireAdmin(sc, "ingest next"); err != nil {
		return nil, err
	}
	return c.queue().Next(ctx, p.After, time.Duration(p.TimeoutMS)*time.Millisecond), nil
}
