package telemetry

import (
	"context"
	"errors"
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Scope is the instrumentation scope name for jobshop's tracer and logger.
const Scope = "github.com/gfx-labs/jobshop"

// Tracing owns the tracer provider and, with an OTLP endpoint, the log
// provider behind the slog bridge.
type Tracing struct {
	TP *sdktrace.TracerProvider
	lp *sdklog.LoggerProvider // nil without an endpoint
}

// Export says where to send traces and logs. Empty URLs export nothing for
// that signal (spans are still made, so trace context still propagates).
type Export struct {
	TracesURL string // full OTLP/HTTP URL, e.g. http://127.0.0.1:4318/v1/traces
	LogsURL   string
	Headers   map[string]string // sent on both signals
	// Resource is on every span and log record. Nil: service.name=jobshop
	// only (tests and processes without a config).
	Resource *resource.Resource
}

// orDefault returns res, or the minimal jobshop resource when res is nil.
func orDefault(res *resource.Resource) *resource.Resource {
	if res != nil {
		return res
	}
	return resource.NewSchemaless(attrServiceName.String("jobshop"))
}

// NewTracing builds the providers.
func NewTracing(ctx context.Context, ex Export) (*Tracing, error) {
	res := orDefault(ex.Resource)
	opts := []sdktrace.TracerProviderOption{sdktrace.WithResource(res)}
	t := &Tracing{}
	var te *otlptrace.Exporter
	if ex.TracesURL != "" {
		var err error
		te, err = otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(ex.TracesURL), otlptracehttp.WithHeaders(ex.Headers))
		if err != nil {
			return nil, err
		}
		opts = append(opts, sdktrace.WithBatcher(te))
	}
	if ex.LogsURL != "" {
		le, err := otlploghttp.New(ctx, otlploghttp.WithEndpointURL(ex.LogsURL), otlploghttp.WithHeaders(ex.Headers))
		if err != nil {
			if te != nil {
				err = errors.Join(err, te.Shutdown(ctx))
			}
			return nil, err
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
