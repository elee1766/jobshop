package control

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/elee1766/jobshop/pkg/store"
)

func tracedController(t *testing.T, yaml string) (*Controller, *tracetest.InMemoryExporter) {
	t.Helper()
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	t.Cleanup(func() { _ = tp.Shutdown(context.WithoutCancel(t.Context())) })
	c, _, _ := scanControllerYAML(t, yaml)
	c.tracer = tp.Tracer("test")
	return c, exp
}

// TestLockWaitNamesHolder (JOB-109): a wait for tickMu records who held
// it (kind and trigger), who waited, and links the holder's span.
func TestLockWaitNamesHolder(t *testing.T) {
	t.Parallel()
	c, exp := tracedController(t, eventTriggerYAML)
	c.lockSpanMin = time.Nanosecond // any wait gets a span; no real wait needed

	holderCtx, holderSpan := c.rootSpan(t.Context(), "trigger event red", attribute.String("jobshop.trigger", "red"))
	unlock := c.lockTickFor(holderCtx, "event", "red")

	// The waiter has read the holder when the hook fires; only then is the
	// lock released, so the wait is behind the event job.
	committed := make(chan struct{})
	c.lockWaitHook = func() { close(committed) }
	waiterCtx, waiterSpan := c.rootSpan(t.Context(), "config reload")
	done := make(chan struct{})
	go func() {
		defer close(done)
		c.lockTick(waiterCtx, "tick")()
	}()
	<-committed
	unlock()
	<-done
	waiterSpan.End()
	holderSpan.End()

	var wait tracetest.SpanStub
	for _, s := range exp.GetSpans() {
		if s.Name == "wait tickMu" {
			wait = s
		}
	}
	require.NotEmpty(t, wait.Name, "the wait got a span")
	assert.Equal(t, "event", attrOf(wait, "jobshop.lock.for"), "the holder")
	assert.Equal(t, "red", attrOf(wait, "jobshop.lock.holder.trigger"))
	assert.Equal(t, "tick", attrOf(wait, "jobshop.lock.waiter"))
	require.Len(t, wait.Links, 1, "a link to the holder's span")
	assert.Equal(t, holderSpan.SpanContext().SpanID(), wait.Links[0].SpanContext.SpanID())
	assert.Equal(t, holderSpan.SpanContext().TraceID(), wait.Links[0].SpanContext.TraceID())
	assert.Equal(t, waiterSpan.SpanContext().SpanID(), wait.Parent.SpanID(), "the wait is a child of the waiter")

	// Released: the next lock has no holder to name.
	c.lockWaitHook = nil
	c.lockTick(t.Context(), "tick")()
	c.holderMu.Lock()
	assert.Empty(t, c.holder.who, "the holder is cleared on unlock")
	c.holderMu.Unlock()
}

// TestEventJobRootSpan (JOB-109): an event job runs in its own goroutine
// and gets its own root span naming the triggers, so what it does (waits,
// store calls, commands) belongs to a trace.
func TestEventJobRootSpan(t *testing.T) {
	t.Parallel()
	c, exp := tracedController(t, eventTriggerYAML)
	c.lockSpanMin = time.Nanosecond

	c.triggerEvents([]store.InboxItem{workflowRun("d1", "master", "success", "aaa")})
	c.wg.Wait()

	var job tracetest.SpanStub
	for _, s := range exp.GetSpans() {
		if s.Name == "trigger event green" {
			job = s
		}
	}
	require.NotEmpty(t, job.Name, "root span for the event job")
	assert.False(t, job.Parent.IsValid(), "a root span")
	assert.Equal(t, []string{"green"}, stringsOf(job, "jobshop.triggers"))
	assert.Len(t, triggerTickets(t, c, "green"), 1, "the job ran the trigger")
}

func stringsOf(s tracetest.SpanStub, key string) []string {
	for _, kv := range s.Attributes {
		if kv.Key == attribute.Key(key) {
			return kv.Value.AsStringSlice()
		}
	}
	return nil
}
