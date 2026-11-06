package telemetry

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTracingCustomPathAndHeaders(t *testing.T) {
	var mu sync.Mutex
	got := map[string]string{} // path -> Authorization
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		got[r.URL.Path] = r.Header.Get("Authorization")
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx := context.Background()
	tr, err := NewTracing(ctx, Export{
		TracesURL: srv.URL + "/insert/opentelemetry/v1/traces",
		Headers:   map[string]string{"Authorization": "Bearer tok"},
	})
	require.NoError(t, err)
	_, span := tr.Tracer().Start(ctx, "test span")
	span.End()
	require.NoError(t, tr.Shutdown(ctx))

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, map[string]string{"/insert/opentelemetry/v1/traces": "Bearer tok"}, got)
}

func TestStartExportsLogs(t *testing.T) {
	var mu sync.Mutex
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx := context.Background()
	var stderr strings.Builder
	p, err := Start(ctx, "desk", Export{LogsURL: srv.URL + "/v1/logs"}, slog.NewTextHandler(&stderr, nil))
	require.NoError(t, err)
	p.Log.Info("hello")
	require.NoError(t, p.Shutdown(ctx))

	assert.Contains(t, stderr.String(), "process=desk")
	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []string{"/v1/logs"}, paths)
}
