package telemetry

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// TestHTTPClientHidesURL (JOB-105): a call is a client span with the host
// and the route template, and nothing of the real path or query, where a
// webhook URL or token may sit. The request on the wire is unchanged.
func TestHTTPClientHidesURL(t *testing.T) {
	t.Parallel()
	var got atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got.Store(r.URL.RequestURI())
		_, _ = io.Copy(io.Discard, r.Body)
	}))
	t.Cleanup(srv.Close)

	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	t.Cleanup(func() { _ = tp.Shutdown(context.WithoutCancel(t.Context())) })
	m, err := NewMetrics()
	require.NoError(t, err)

	ctx, root := tp.Tracer("t").Start(WithMetrics(t.Context(), m), "root")
	ctx = WithRoute(ctx, "/hook/{id}")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, srv.URL+"/hook/SECRET-TOKEN?key=SECRET-QUERY", strings.NewReader("x"))
	require.NoError(t, err)
	resp, err := HTTPClient("hook", time.Minute).Do(req)
	require.NoError(t, err)
	_, _ = io.Copy(io.Discard, resp.Body)
	require.NoError(t, resp.Body.Close())
	root.End()

	assert.Equal(t, "/hook/SECRET-TOKEN?key=SECRET-QUERY", got.Load(), "the real URL goes on the wire")
	var client tracetest.SpanStub
	for _, s := range exp.GetSpans() {
		if strings.HasPrefix(s.Name, "HTTP ") {
			client = s
		}
	}
	require.NotEmpty(t, client.Name)
	assert.Equal(t, "HTTP POST "+strings.TrimPrefix(srv.URL, "http://")+"/hook/{id}", client.Name)
	assert.Equal(t, root.SpanContext().SpanID(), client.Parent.SpanID())
	for _, kv := range client.Attributes {
		v := kv.Value.String()
		assert.NotContains(t, v, "SECRET", string(kv.Key))
	}
}

// TestRecordSlow: only work past the threshold, under a recording span,
// becomes a span.
func TestRecordSlow(t *testing.T) {
	t.Parallel()
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	ctx, root := tp.Tracer("t").Start(t.Context(), "root")

	RecordSlow(ctx, "fast", time.Now(), time.Hour, nil)
	RecordSlow(ctx, "slow", time.Now().Add(-time.Second), time.Millisecond, nil)
	RecordSlow(context.Background(), "orphan", time.Now().Add(-time.Second), time.Millisecond, nil)
	root.End()

	var names []string
	for _, s := range exp.GetSpans() {
		names = append(names, s.Name)
	}
	assert.ElementsMatch(t, []string{"slow", "root"}, names)
}
