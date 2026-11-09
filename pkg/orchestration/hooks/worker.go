package hooks

import (
	"context"
	"log/slog"
	"sync/atomic"
)

// drainBudget is how many invocations one worker run processes before it
// reschedules itself, so a long queue cannot hold the worker forever.
const drainBudget = 64

// Worker runs the bounded callback drain: enqueue what runs are waiting
// for, process invocations one at a time, enqueue again (an applied batch
// can make the next one due), until nothing is left. At most one drain
// runs; a Kick during one makes it go round again. The host supplies the
// goroutine home and the two steps; the worker holds no host lock itself.
type Worker struct {
	// Go starts f tracked by the host, so shutdown waits for it.
	Go func(f func())
	// Context is the host's lifetime context. A nil or ended one makes
	// Kick a no-op.
	Context func() context.Context
	// Enqueue records the invocations that are due.
	Enqueue func(ctx context.Context) error
	// Process runs one durable invocation (Engine.Process).
	Process func(ctx context.Context) (bool, error)
	Log     *slog.Logger

	running, pending atomic.Bool
}

// Kick asks for a drain. It returns at once.
func (w *Worker) Kick() {
	ctx := w.Context()
	if ctx == nil || ctx.Err() != nil {
		return
	}
	w.pending.Store(true)
	if !w.running.CompareAndSwap(false, true) {
		return
	}
	w.Go(func() {
		w.pending.Store(false)
		defer func() {
			w.running.Store(false)
			if w.pending.Swap(false) {
				w.Kick()
			}
		}()
		if err := w.Enqueue(ctx); err != nil {
			w.Log.ErrorContext(ctx, "enqueue workflow hooks", "err", err)
			return
		}
		for range drainBudget {
			processed, err := w.Process(ctx)
			if err != nil {
				w.Log.ErrorContext(ctx, "process hook", "err", err)
				return
			}
			if !processed {
				return
			}
			if err := w.Enqueue(ctx); err != nil {
				return
			}
		}
		// Continue draining when the bounded pass exhausted its budget.
		w.pending.Store(true)
	})
}
