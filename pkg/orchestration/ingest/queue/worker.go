package queue

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// Worker is the loop of `jobshop ingest work` (#443): claim a
// pending item from the controller, run the admission hooks here, report
// the run. It never opens the database; every write is the controller's.
type Worker struct {
	// Call is one RPC to the controller's admin socket.
	Call func(ctx context.Context, method string, params, out any) error
	// Dir is the root: hook commands run in it and secret files are
	// relative to it.
	Dir    string
	Worker string
	Host   string
	Log    *slog.Logger
	// Dispatch: after admitting, also dispatch admitted items through
	// the controller (#633). NoAdmit skips admission, for a dispatch-only
	// worker while the controller admits.
	Dispatch bool
	NoAdmit  bool
}

// DispatchOnce asks the controller to dispatch admitted items under
// this worker's lease holder, returning how many it claimed.
func (w Worker) DispatchOnce(ctx context.Context) (int, error) {
	var res rpc.IngestDispatchResult
	err := w.Call(ctx, rpc.MethodIngestDispatch, rpc.IngestClaimParams{Worker: w.Worker, Host: w.Host}, &res)
	return res.Dispatched, err
}

// Pass admits (unless NoAdmit) then dispatches (with Dispatch),
// returning the admitted and dispatched counts.
func (w Worker) Pass(ctx context.Context) (admitted, dispatched int, err error) {
	if !w.NoAdmit {
		if admitted, err = w.Once(ctx); err != nil {
			return admitted, 0, err
		}
	}
	if w.Dispatch {
		dispatched, err = w.DispatchOnce(ctx)
	}
	return admitted, dispatched, err
}

// Once claims and admits until nothing is pending, returning how many
// items it admitted.
func (w Worker) Once(ctx context.Context) (int, error) {
	ar := Runner{Dir: w.Dir}
	n := 0
	for ctx.Err() == nil {
		var cl rpc.IngestClaimResult
		if err := w.Call(ctx, rpc.MethodIngestClaim, rpc.IngestClaimParams{Worker: w.Worker, Host: w.Host}, &cl); err != nil {
			return n, err
		}
		if cl.Item == nil {
			return n, nil
		}
		// Hooks must finish inside the lease the controller granted.
		hctx, cancel := context.WithTimeout(ctx, time.Duration(cl.LeaseMS)*time.Millisecond)
		run := ar.Run(hctx, cl.Project, cl.Hooks, *cl.Item, cl.Answered)
		cancel()
		raw, err := json.Marshal(run)
		if err != nil {
			return n, err
		}
		var done json.RawMessage
		if err := w.Call(ctx, rpc.MethodIngestAdmit, rpc.IngestAdmitParams{Worker: w.Worker, ID: cl.Item.ID, Run: raw}, &done); err != nil {
			// A lost lease: another worker has or will claim the item.
			w.logger().WarnContext(ctx, "ingest work: admit", "delivery", cl.Item.DeliveryID, "err", err)
			continue
		}
		n++
	}
	return n, ctx.Err()
}

// Run calls Pass until ctx ends. Between passes it long-polls
// ingest.next (#679), so an item is admitted as it lands; a long poll
// that fails (a controller restart, or one without the method) falls
// back to waiting idle. A failed pass is logged and retried, so a
// controller restart does not stop the worker.
func (w Worker) Run(ctx context.Context, idle time.Duration, after func(time.Duration) <-chan time.Time) error {
	// After -1 never matches a seq, so the first call returns the
	// current one at once: the cursor taken before the first pass.
	seq, ok := w.next(ctx, -1)
	for ctx.Err() == nil {
		if _, _, err := w.Pass(ctx); err != nil && ctx.Err() == nil {
			w.logger().WarnContext(ctx, "ingest work", "err", err)
		}
		if ok {
			if seq, ok = w.next(ctx, seq); ok {
				continue
			}
		}
		select {
		case <-ctx.Done():
		case <-after(idle):
		}
		seq, ok = w.next(ctx, -1)
	}
	return nil
}

// next is one ingest.next call; false when it failed.
func (w Worker) next(ctx context.Context, after int64) (int64, bool) {
	var res rpc.IngestNextResult
	if err := w.Call(ctx, rpc.MethodIngestNext, rpc.IngestNextParams{After: after}, &res); err != nil {
		if ctx.Err() == nil {
			w.logger().DebugContext(ctx, "ingest work: next", "err", err)
		}
		return 0, false
	}
	return res.Seq, true
}

func (w Worker) logger() *slog.Logger {
	if w.Log == nil {
		return slog.New(slog.DiscardHandler)
	}
	return w.Log
}
