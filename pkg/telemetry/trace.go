package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Scope is the instrumentation scope name for jobshop's tracer and logger.
const Scope = "github.com/elee1766/jobshop"

// Tracing owns the tracer provider and, with an OTLP endpoint, the log
// provider behind the slog bridge.
type Tracing struct {
	TP *sdktrace.TracerProvider
	lp *sdklog.LoggerProvider // nil without an endpoint
}

// NewTracing builds the providers. endpoint is an OTLP/HTTP base URL such
// as http://127.0.0.1:4318; empty exports nothing (spans are still made, so
// trace context still propagates). extra processors are for tests.
func NewTracing(ctx context.Context, endpoint string, extra ...sdktrace.SpanProcessor) (*Tracing, error) {
	res := resource.NewSchemaless(attribute.String("service.name", "jobshop"))
	opts := []sdktrace.TracerProviderOption{sdktrace.WithResource(res)}
	for _, p := range extra {
		opts = append(opts, sdktrace.WithSpanProcessor(p))
	}
	t := &Tracing{}
	if endpoint != "" {
		base := strings.TrimSuffix(endpoint, "/")
		te, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(base+"/v1/traces"))
		if err != nil {
			return nil, err
		}
		opts = append(opts, sdktrace.WithBatcher(te))
		le, err := otlploghttp.New(ctx, otlploghttp.WithEndpointURL(base+"/v1/logs"))
		if err != nil {
			return nil, errors.Join(err, te.Shutdown(ctx))
		}
		t.lp = sdklog.NewLoggerProvider(sdklog.WithResource(res), sdklog.WithProcessor(sdklog.NewBatchProcessor(le)))
	}
	t.TP = sdktrace.NewTracerProvider(opts...)
	return t, nil
}

// Tracer returns jobshop's tracer.
func (t *Tracing) Tracer() trace.Tracer { return t.TP.Tracer(Scope) }

// LogHandler returns base, plus the OTel bridge when logs are exported.
// Records logged with a span in their context carry its trace and span ids.
func (t *Tracing) LogHandler(base slog.Handler) slog.Handler {
	if t.lp == nil {
		return base
	}
	return slog.NewMultiHandler(base, otelslog.NewHandler(Scope, otelslog.WithLoggerProvider(t.lp)))
}

// Shutdown flushes and stops the providers.
func (t *Tracing) Shutdown(ctx context.Context) error {
	err := t.TP.Shutdown(ctx)
	if t.lp != nil {
		err = errors.Join(err, t.lp.Shutdown(ctx))
	}
	return err
}

var propagator = propagation.TraceContext{}

// TraceParent formats the span context in ctx as a W3C traceparent, or "".
func TraceParent(ctx context.Context) string {
	c := propagation.MapCarrier{}
	propagator.Inject(ctx, c)
	return c["traceparent"]
}

// WithTraceParent returns ctx with tp as its remote parent. An empty or
// malformed tp leaves ctx unchanged.
func WithTraceParent(ctx context.Context, tp string) context.Context {
	if tp == "" {
		return ctx
	}
	return propagator.Extract(ctx, propagation.MapCarrier{"traceparent": tp})
}
