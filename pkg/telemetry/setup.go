package telemetry

import (
	"context"
	"io"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// Process is one jobshop process's logger and tracer. Every long-running
// command builds one with Start, so OTLP export is wired in one place.
type Process struct {
	Log    *slog.Logger
	Tracer trace.Tracer
	t      *Tracing
}

// Start builds the logger and tracer for the process named name (serve,
// desk, ingest, ...). Logs go to stderr as text and, when ex has a logs
// URL, to OTLP. Records carry process=name.
func Start(ctx context.Context, name string, ex Export, stderr io.Writer) (*Process, error) {
	t, err := NewTracing(ctx, ex)
	if err != nil {
		return nil, err
	}
	log := slog.New(t.LogHandler(slog.NewTextHandler(stderr, nil))).With("process", name)
	return &Process{Log: log, Tracer: t.Tracer(), t: t}, nil
}

// Tracing returns the providers, for a process that hands them on.
func (p *Process) Tracing() *Tracing { return p.t }

// Shutdown flushes pending spans and log records.
func (p *Process) Shutdown(ctx context.Context) error { return p.t.Shutdown(ctx) }
