package control

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// lockTick takes tickMu for who ("tick", "webhook", "event", "check"),
// records how long it waited as a histogram and, when the wait was long
// enough to matter, as a span under ctx. It returns the unlock func.
func (c *Controller) lockTick(ctx context.Context, who string) func() {
	start := time.Now()
	c.tickMu.Lock()
	c.metrics.SinceLock(ctx, "tick", who, start)
	telemetry.RecordSlow(ctx, "wait tickMu", start, lockSpanThreshold, nil,
		telemetry.AttrLock.String("tickMu"), telemetry.AttrLockHolder.String(who))
	return c.tickMu.Unlock
}

// lockSpanThreshold is how long a lock wait must be to get a span.
const lockSpanThreshold = 10 * time.Millisecond

// step runs one named part of a pass under a child span of ctx's span and
// records its duration (jobshop.tick.step.duration{step}), so a slow tick
// names its slow part. fn gets the step's ctx: the store, HTTP and exec
// spans below it nest under the step.
func (c *Controller) step(ctx context.Context, name string, fn func(ctx context.Context)) {
	ctx, span := telemetry.StartSpan(ctx, "tick "+name, attribute.String("jobshop.tick.step", name))
	start := time.Now()
	defer func() {
		c.metrics.SinceStep(ctx, name, start)
		span.End()
	}()
	fn(ctx)
}

// passStep is step for a pass function that takes the config the pass
// loaded once (JOB-104).
func (c *Controller) passStep(ctx context.Context, cfg *config.Config, name string, fn func(ctx context.Context, cfg *config.Config)) {
	c.step(ctx, name, func(ctx context.Context) { fn(ctx, cfg) })
}
