package telemetry

import (
	"context"
	"log/slog"

	"github.com/elee1766/jobshop/pkg/logctx"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
)

// Process is one jobshop process's logger and tracer. Every long-running
// command builds one with Start, so OTLP export is wired in one place.
type Process struct {
	Log    *slog.Logger
	Tracer trace.Tracer
	// Resource is the one on the traces and logs; the controller passes it
	// to NewMetrics so target_info matches (JOB-106).
	Resource *resource.Resource
	t        *Tracing
}

// Start builds the logger and tracer for the process named name (serve,
// desk, ingest, ...). Logs go to base (the --log-level/--log-format
// handler) and, when ex has a logs URL, to OTLP. Records carry
// process=name and the ids in ctx (logctx, outermost so both sinks get
// them).
func Start(ctx context.Context, name string, ex Export, base slog.Handler) (*Process, error) {
	t, err := NewTracing(ctx, ex)
	if err != nil {
		return nil, err
	}
	log := slog.New(logctx.NewHandler(t.LogHandler(logctx.Unwrap(base)))).With("process", name)
	return &Process{Log: log, Tracer: t.Tracer(), Resource: orDefault(ex.Resource), t: t}, nil
}

// Tracing returns the providers, for a process that hands them on.
func (p *Process) Tracing() *Tracing { return p.t }

// Shutdown flushes pending spans and log records.
func (p *Process) Shutdown(ctx context.Context) error { return p.t.Shutdown(ctx) }
