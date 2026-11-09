package control

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/support/telemetry"
)

// lockHolder is who holds tickMu now, set under the lock: the kind (tick,
// webhook, event, check), the trigger an event or check job is for, and
// the holder's span, so a waiter can link it (JOB-109).
type lockHolder struct {
	who, trigger string
	span         trace.SpanContext
}

// lockTick takes tickMu for who ("tick", "webhook", "event", "check"),
// records how long it waited as a histogram and, when the wait was long
// enough to matter, as a span under ctx. The span names the holder it
// waited behind (jobshop.lock.for, plus jobshop.lock.holder.trigger for a
// trigger job) and links the holder's span, so a slow pass or reload says
// what blocked it; the waiter is jobshop.lock.waiter. It returns the
// unlock func.
func (c *Controller) lockTick(ctx context.Context, who string) func() {
	return c.lockTickFor(ctx, who, "")
}

// lockTickFor is lockTick for a holder that works for one trigger.
func (c *Controller) lockTickFor(ctx context.Context, who, trigger string) func() {
	start := time.Now()
	// Read before waiting: the holder at the moment this call began is
	// what it waits behind. It may be gone by the time Lock returns, and
	// then the wait was short.
	c.holderMu.Lock()
	prev := c.holder
	c.holderMu.Unlock()
	if c.lockWaitHook != nil {
		c.lockWaitHook()
	}
	c.tickMu.Lock()
	c.metrics.SinceLock(ctx, "tick", who, start)
	if prev.who != "" {
		attrs := []attribute.KeyValue{
			telemetry.AttrLock.String("tickMu"),
			telemetry.AttrLockHolder.String(prev.who),
			telemetry.AttrLockWaiter.String(who),
		}
		if prev.trigger != "" {
			attrs = append(attrs, telemetry.AttrLockHolderTrigger.String(prev.trigger))
		}
		var links []trace.Link
		if prev.span.IsValid() {
			links = append(links, trace.Link{SpanContext: prev.span})
		}
		telemetry.RecordSlowLinks(ctx, "wait tickMu", start, c.lockSpanFloor(), nil, links, attrs...)
	} else {
		telemetry.RecordSlow(ctx, "wait tickMu", start, c.lockSpanFloor(), nil,
			telemetry.AttrLock.String("tickMu"), telemetry.AttrLockWaiter.String(who))
	}
	c.holderMu.Lock()
	c.holder = lockHolder{who: who, trigger: trigger, span: trace.SpanContextFromContext(ctx)}
	c.holderMu.Unlock()
	return func() {
		c.holderMu.Lock()
		c.holder = lockHolder{}
		c.holderMu.Unlock()
		c.tickMu.Unlock()
	}
}

// rootSpan starts a root span (a new trace) for work that runs in its own
// goroutine, so what it does is not an orphan: an event job or a trigger
// check replay. The controller's tracer is nil in tests that build one by
// hand, which get a non-recording span.
func (c *Controller) rootSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	tr := c.tracer
	if tr == nil {
		tr = noop.NewTracerProvider().Tracer(telemetry.Scope)
	}
	return tr.Start(ctx, name, trace.WithNewRoot(), trace.WithAttributes(attrs...))
}

// lockSpanThreshold is how long a lock wait must be to get a span.
const lockSpanThreshold = 10 * time.Millisecond

func (c *Controller) lockSpanFloor() time.Duration {
	if c.lockSpanMin > 0 {
		return c.lockSpanMin
	}
	return lockSpanThreshold
}

// step runs one named part of a pass under a child span of ctx's span and
// records its duration (jobshop.tick.step.duration{step}), so a slow tick
// names its slow part. fn gets the step's ctx: the store, HTTP and exec
// spans below it nest under the step.
func (c *Controller) step(ctx context.Context, name string, fn func(ctx context.Context)) {
	ctx, span := telemetry.StartSpan(ctx, "tick "+name, attribute.String("jobshop.tick.step", name))
	start := time.Now()
	defer func() {
		c.metrics.SinceStep(ctx, name, start)
		c.statusMu.Lock()
		if c.tickSteps == nil {
			c.tickSteps = map[string]float64{}
		}
		c.tickSteps[name] = time.Since(start).Seconds()
		c.statusMu.Unlock()
		span.End()
	}()
	fn(ctx)
}

// passStep is step for a pass function that takes the config the pass
// loaded once (JOB-104).
func (c *Controller) passStep(ctx context.Context, cfg *config.Config, name string, fn func(ctx context.Context, cfg *config.Config)) {
	c.step(ctx, name, func(ctx context.Context) { fn(ctx, cfg) })
}
