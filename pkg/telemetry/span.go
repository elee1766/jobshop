package telemetry

import (
	"context"
	"os/exec"
	"path/filepath"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Attribute keys for the spans below.
const (
	AttrTickStep = attribute.Key("jobshop.tick.step")
	AttrOutKind  = attribute.Key("jobshop.outbound.kind")
	AttrLock     = attribute.Key("jobshop.lock")
	// AttrLockHolder is who held the lock when a wait for it began (tick,
	// webhook, event, check); AttrLockHolderTrigger names the trigger
	// behind an event or check holder. AttrLockWaiter is who waited.
	AttrLockHolder        = attribute.Key("jobshop.lock.for")
	AttrLockHolderTrigger = attribute.Key("jobshop.lock.holder.trigger")
	AttrLockWaiter        = attribute.Key("jobshop.lock.waiter")
	AttrExecCmd           = attribute.Key("process.command")
	AttrExecLabel         = attribute.Key("jobshop.exec.label")
	AttrExecExit          = attribute.Key("process.exit.code")
	AttrExecTimed         = attribute.Key("jobshop.exec.timed_out")
	AttrStoreOp           = attribute.Key("jobshop.store.op")
	AttrStoreColl         = attribute.Key("jobshop.store.collection")
)

type metricsKey struct{}

// WithMetrics returns ctx carrying m, so code far from the controller (an
// HTTP transport, a lock helper) can record without a global or a new
// parameter.
func WithMetrics(ctx context.Context, m *Metrics) context.Context {
	return context.WithValue(ctx, metricsKey{}, m)
}

// MetricsFrom is the Metrics in ctx, or nil.
func MetricsFrom(ctx context.Context) *Metrics {
	m, _ := ctx.Value(metricsKey{}).(*Metrics)
	return m
}

// StartSpan starts a child of the span in ctx. With no span in ctx it makes a
// non-recording one, so a package without a tracer (an HTTP client, the
// store) adds spans only under a traced caller and never needs a
// provider of its own. Sampling stays whatever the root span chose.
func StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	tr := trace.SpanFromContext(ctx).TracerProvider().Tracer(Scope)
	return tr.Start(ctx, name, trace.WithAttributes(attrs...))
}

// RecordSlow adds a finished child span of ctx's span for work that began
// at start, but only when it took at least threshold. Volume stays low for
// things done thousands of times a tick, such as store queries, while a
// slow one still shows under the step that made it. A ctx with no
// recording span records nothing.
func RecordSlow(ctx context.Context, name string, start time.Time, threshold time.Duration, err error, attrs ...attribute.KeyValue) {
	RecordSlowLinks(ctx, name, start, threshold, err, nil, attrs...)
}

// RecordSlowLinks is RecordSlow with span links, for work whose cause is
// in another trace: a lock wait links the span of the holder.
func RecordSlowLinks(ctx context.Context, name string, start time.Time, threshold time.Duration, err error, links []trace.Link, attrs ...attribute.KeyValue) {
	parent := trace.SpanFromContext(ctx)
	if !parent.IsRecording() || time.Since(start) < threshold {
		return
	}
	_, span := parent.TracerProvider().Tracer(Scope).Start(ctx, name, trace.WithTimestamp(start), trace.WithAttributes(attrs...), trace.WithLinks(links...))
	End(span, err)
}

// End ends span, marking it failed when err is not nil.
func End(span trace.Span, err error) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}

// ExecContext is exec.CommandContext under an exec span (StartExec). Run
// the command any way you like, then call done once:
//
//	cmd, done := telemetry.ExecContext(ctx, "gh-api", "gh", "api", path)
//	out, err := cmd.Output()
//	done(err)
//
// The span records the exit code from cmd, and whether ctx ended.
func ExecContext(ctx context.Context, label, name string, args ...string) (*exec.Cmd, func(err error)) {
	ctx, end := StartExec(ctx, name, label)
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd, func(err error) { end(ExitCode(cmd), err) }
}

// ExitCode is the exit code of a finished cmd, or -1 when it never ran.
func ExitCode(cmd *exec.Cmd) int {
	if cmd.ProcessState == nil {
		return -1
	}
	return cmd.ProcessState.ExitCode()
}

// StartExec starts the span for one subprocess. cmd is the program (only
// its base name is kept) and label says what it is for: a trigger or hook
// name, or the gh subcommand. Never pass the script or an env value: the
// span carries the command name, the label, the exit code and whether it
// timed out. The returned func ends the span.
func StartExec(ctx context.Context, cmd, label string) (context.Context, func(exitCode int, err error)) {
	ctx, span := StartSpan(ctx, "exec "+filepath.Base(cmd), AttrExecCmd.String(filepath.Base(cmd)), AttrExecLabel.String(label))
	return ctx, func(exitCode int, err error) {
		span.SetAttributes(AttrExecExit.Int(exitCode))
		if ctx.Err() != nil {
			span.SetAttributes(AttrExecTimed.Bool(true))
		}
		End(span, err)
	}
}

// The Since methods do nothing on a nil *Metrics, so a controller built
// without telemetry (tests) needs no guards.

// SinceStep records one tick step's duration.
func (m *Metrics) SinceStep(ctx context.Context, step string, start time.Time) {
	if m == nil {
		return
	}
	m.TickStep.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(attribute.String("step", step)))
}

// SinceLock records the time spent waiting for lock, taken for who.
func (m *Metrics) SinceLock(ctx context.Context, lock, who string, start time.Time) {
	if m == nil {
		return
	}
	m.LockWait.Record(ctx, time.Since(start).Seconds(),
		metric.WithAttributes(attribute.String("lock", lock), attribute.String("for", who)))
}

// SinceOutbound records one outbound call's duration by kind and target
// host.
func (m *Metrics) SinceOutbound(ctx context.Context, kind, target string, d time.Duration) {
	if m == nil {
		return
	}
	m.Outbound.Record(ctx, d.Seconds(),
		metric.WithAttributes(attribute.String("kind", kind), attribute.String("target", target)))
}
