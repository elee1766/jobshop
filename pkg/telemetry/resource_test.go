package telemetry

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
)

// TestResourceOnAllSignals checks traces, logs and metrics carry one
// resource (JOB-106). OTLP/HTTP bodies are uncompressed protobuf, so
// attribute keys and values appear in them as plain bytes.
func TestResourceOnAllSignals(t *testing.T) {
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "")
	t.Setenv("OTEL_SERVICE_NAME", "")
	var mu sync.Mutex
	bodies := map[string]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		bodies[r.URL.Path] += string(b)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx := context.Background()
	res, err := NewResource(ctx, Identity{
		Instance: "selfdev", Root: "/var/lib/jobshop", Version: "0.6.18", Commit: "abc123",
		Attributes: map[string]string{"deployment.environment": "prod"},
	})
	require.NoError(t, err)
	p, err := Start(ctx, "serve", Export{TracesURL: srv.URL + "/v1/traces", LogsURL: srv.URL + "/v1/logs", Resource: res},
		slog.DiscardHandler)
	require.NoError(t, err)
	_, span := p.Tracer.Start(ctx, "test span")
	span.End()
	p.Log.Info("hello")
	require.NoError(t, p.Shutdown(ctx))

	want := []string{
		"service.name", "jobshop", "service.version", "0.6.18+abc123", "jobshop.instance",
		"service.instance.id", "selfdev", "jobshop.commit", "abc123", "deployment.environment", "prod",
	}
	mu.Lock()
	for _, path := range []string{"/v1/traces", "/v1/logs"} {
		for _, w := range want {
			assert.Contains(t, bodies[path], w, path)
		}
	}
	mu.Unlock()

	m, err := NewMetricsFor(p.Resource)
	require.NoError(t, err)
	defer func() { _ = m.Shutdown(ctx) }()
	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))
	var info string
	for l := range strings.SplitSeq(rec.Body.String(), "\n") {
		if strings.HasPrefix(l, "target_info{") {
			info = l
		}
	}
	for _, w := range []string{
		`service_name="jobshop"`, `service_version="0.6.18+abc123"`, `jobshop_instance="selfdev"`,
		`service_instance_id="selfdev"`, `deployment_environment="prod"`,
	} {
		assert.Contains(t, info, w)
	}
}

// TestResourceEnvWins pins the precedence: OTEL_SERVICE_NAME and
// OTEL_RESOURCE_ATTRIBUTES override the config, which overrides the
// defaults.
func TestResourceEnvWins(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "jobshop-canary")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "jobshop.instance=from-env,deployment.environment=staging")
	res, err := NewResource(context.Background(), Identity{
		Instance: "selfdev", Version: "v1",
		Attributes: map[string]string{"deployment.environment": "prod", "team": "a"},
	})
	require.NoError(t, err)
	got := map[attribute.Key]string{}
	for _, kv := range res.Attributes() {
		got[kv.Key] = kv.Value.String()
	}
	assert.Equal(t, "jobshop-canary", got["service.name"])
	assert.Equal(t, "from-env", got["jobshop.instance"])
	assert.Equal(t, "selfdev", got["service.instance.id"], "env did not set it")
	assert.Equal(t, "staging", got["deployment.environment"])
	assert.Equal(t, "a", got["team"])
	assert.Equal(t, "v1", got["service.version"])
}

func TestDefaultInstance(t *testing.T) {
	assert.Equal(t, "selfdev", DefaultInstance("/srv/selfdev"))
	got := DefaultInstance("/var/lib/jobshop")
	assert.True(t, strings.HasSuffix(got, ":/var/lib/jobshop"), got)
	assert.NotEqual(t, ":/var/lib/jobshop", got)
}

func TestServiceVersion(t *testing.T) {
	assert.Equal(t, "0.6.18+abc", serviceVersion("0.6.18+abc", "abc"), "nix build already has it")
	assert.Equal(t, "dev+abc", serviceVersion("dev", "abc"))
	assert.Equal(t, "dev", serviceVersion("dev", "unknown"))
}
