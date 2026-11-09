package control

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

<<<<<<< HEAD
	"github.com/elee1766/jobshop/pkg/integrations/identity"
	"github.com/elee1766/jobshop/pkg/store"
=======
	"github.com/elee1766/jobshop/pkg/identity"
	"github.com/elee1766/jobshop/pkg/storage/store"
>>>>>>> 7fea25a3 (refactor(storage): group store and journal packages under pkg/storage (JOB-173))
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// tracingYAML has a trigger with a check command and a Linear intake
// scan trigger, so one pass runs an exec and an outbound HTTP call.
const tracingYAML = `projects:
  js: {}
agents:
  fake: {command: /bin/false}
identities:
  lin:
    kind: linear
    token_file: %[1]s/token
    workspace: acme
    teams: [ENG]
pools:
  dev: {agent: fake, dir: %[1]q}
  plan: {agent: fake, dir: %[1]q}
workflows:
  triage:
    vars: [signal, subject, detail]
    escalate_pool: plan
    steps:
      - {id: diagnose, pool: plan, retries: 1, title: "Finding", prompt: "x"}
triggers:
  probe:
    when:
      check: echo SECRET-SCRIPT-OUTPUT; exit 1
    ticket: {title: probe, pool: dev}
  scan:
    when:
      scan:
        - {signal: linear_issue_no_run, team: ENG, states: "Todo", labels: agent, poll: 5m}
    run:
      workflow: triage
      vars: {signal: "{{.signal}}", subject: "{{.subject}}", detail: "{{.detail}}"}
`

// TestTickTraces (JOB-105): one supervisor pass is a tick span with a
// child span per step; a trigger check is an exec span naming the command
// without its script; a call to Linear is an HTTP client span under the
// run_triggers step that shows the host and no query or token; and the
// step and outbound histograms record.
func TestTickTraces(t *testing.T) {
	t.Parallel()
	gql := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		_, _ = w.Write([]byte(`{"data":{"issues":{"nodes":[],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}`))
	}))
	t.Cleanup(gql.Close)

	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	t.Cleanup(func() { _ = tp.Shutdown(context.WithoutCancel(t.Context())) })

	c, _, dir := scanControllerYAML(t, tracingYAML)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "token"), []byte("lin_api_SECRET\n"), 0o600))
	c.minter, c.linearURL = identity.NewMinter(), gql.URL
	c.tracer = tp.Tracer("test")
	m, err := telemetry.NewMetrics()
	require.NoError(t, err)
	c.metrics = m
	c.ctx = telemetry.WithMetrics(t.Context(), m)
	c.store = store.Traced(c.store)

	c.tick()

	spans := exp.GetSpans()
	byName := map[string]tracetest.SpanStub{}
	for _, s := range spans {
		byName[s.Name] = s
	}
	tick, ok := byName["supervisor tick"]
	require.True(t, ok, "tick span")

	steps := []string{
		"sweep", "provider_pass", "run_triggers", "place_steps", "supervise", "run_auto_steps",
		"report_orphans", "scan_sessions", "sweep_runs", "sweep_escalation_chats", "sweep_attention_chats",
		"sweep_batches", "sweep_linear_asks", "sweep_linear_agent_sessions", "prune_inbox", "prune_write_keys",
		"prune_artifacts", "compress_journals",
	}
	for _, s := range steps {
		sp, ok := byName["tick "+s]
		if assert.True(t, ok, "step span %s", s) {
			assert.Equal(t, tick.SpanContext.SpanID(), sp.Parent.SpanID(), "%s is a child of the tick", s)
		}
	}

	step := byName["tick run_triggers"]
	trig, ok := byName["trigger probe"]
	require.True(t, ok, "per-trigger span")
	assert.Equal(t, step.SpanContext.SpanID(), trig.Parent.SpanID())
	assert.Equal(t, "held", attrOf(trig, "jobshop.trigger.result"))

	exec, ok := byName["exec sh"]
	require.True(t, ok, "exec span for the trigger check")
	assert.Equal(t, trig.SpanContext.SpanID(), exec.Parent.SpanID(), "the check runs under its trigger")
	assert.Equal(t, "trigger-check", attrOf(exec, "jobshop.exec.label"))
	assert.Equal(t, "sh", attrOf(exec, "process.command"))
	assert.Equal(t, int64(1), intAttrOf(exec, "process.exit.code"))
	for _, kv := range exec.Attributes {
		assert.NotContains(t, kv.Value.String(), "SECRET-SCRIPT-OUTPUT", "the script and its output stay off the span")
	}

	var httpSpan tracetest.SpanStub
	for _, s := range spans {
		if strings.HasPrefix(s.Name, "HTTP ") {
			httpSpan = s
		}
	}
	require.NotEmpty(t, httpSpan.Name, "client span for the Linear call")
	assert.Equal(t, "linear", attrOf(httpSpan, "jobshop.outbound.kind"))
	assert.Contains(t, httpSpan.Name, strings.TrimPrefix(gql.URL, "http://"))
	assert.True(t, descends(spans, httpSpan, step), "the HTTP span nests under the run_triggers step")
	for _, kv := range httpSpan.Attributes {
		v := kv.Value.String()
		assert.NotContains(t, v, "lin_api_SECRET")
		assert.NotContains(t, v, "?")
	}

	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := rec.Body.String()
	assert.Contains(t, body, `jobshop_tick_step_duration_seconds_count{`)
	assert.Contains(t, body, `step="run_triggers"`)
	assert.Contains(t, body, `jobshop_outbound_duration_seconds_count{`)
	assert.Contains(t, body, `kind="linear"`)
	assert.Contains(t, body, `jobshop_lock_wait_seconds_count{`)
}

// descends reports whether span is a descendant of ancestor.
func descends(spans tracetest.SpanStubs, span, ancestor tracetest.SpanStub) bool {
	for span.Parent.IsValid() {
		if span.Parent.SpanID() == ancestor.SpanContext.SpanID() {
			return true
		}
		i := slices.IndexFunc(spans, func(s tracetest.SpanStub) bool { return s.SpanContext.SpanID() == span.Parent.SpanID() })
		if i < 0 {
			return false
		}
		span = spans[i]
	}
	return false
}

func attrOf(s tracetest.SpanStub, key string) string {
	for _, kv := range s.Attributes {
		if kv.Key == attribute.Key(key) {
			return kv.Value.AsString()
		}
	}
	return ""
}

func intAttrOf(s tracetest.SpanStub, key string) int64 {
	for _, kv := range s.Attributes {
		if kv.Key == attribute.Key(key) {
			return kv.Value.AsInt64()
		}
	}
	return -1
}
