package telemetry

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// memExporter records exported log records. block, if set, makes Export
// wait until it is closed (a stalled endpoint).
type memExporter struct {
	mu      sync.Mutex
	records []sdklog.Record
	block   chan struct{}
}

func (m *memExporter) Export(ctx context.Context, rs []sdklog.Record) error {
	if m.block != nil {
		select {
		case <-m.block:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range rs {
		m.records = append(m.records, r.Clone())
	}
	return nil
}
func (m *memExporter) Shutdown(context.Context) error   { return nil }
func (m *memExporter) ForceFlush(context.Context) error { return nil }

func (m *memExporter) all() []sdklog.Record {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]sdklog.Record(nil), m.records...)
}

func attrs(r sdklog.Record) map[attribute.Key]string {
	out := map[attribute.Key]string{}
	r.WalkAttributes(func(kv attribute.KeyValue) bool {
		out[kv.Key] = kv.Value.String()
		return true
	})
	return out
}

func newTestExporter(t *testing.T, cfg ConvConfig) (*ConvExporter, *memExporter) {
	t.Helper()
	mem := &memExporter{}
	if cfg.Exporter == nil {
		cfg.Exporter = mem
	}
	cfg.Session, cfg.Pool, cfg.Project = "js-s1", "dev", "js"
	e, err := NewConvExporter(t.Context(), cfg)
	require.NoError(t, err)
	return e, mem
}

func closeExporter(t *testing.T, e *ConvExporter) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, e.Close(ctx))
}

// A GitHub token pattern the default gitleaks rules detect.
const ghToken = "ghp_" + "Zx9Qw3Er5Ty7Ui9Op1As3Df5Gh7Jk9Lz1Xc3V"

func TestConvExportRecords(t *testing.T) {
	e, mem := newTestExporter(t, ConvConfig{Env: []string{"MY_API_TOKEN=launch-secret-value-123", "SHORT_TOKEN=x"}})
	frames := []struct{ from, line string }{
		{"controller", `{"jsonrpc":"2.0","id":1,"method":"session/prompt","params":{"sessionId":"a","_meta":{"jobshop/ticket":"js-t1"},"prompt":[{"type":"text","text":"do it"}]}}`},
		{"agent", `{"jsonrpc":"2.0","method":"session/update","params":{"update":{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"hello "}}}}`},
		{"agent", `{"jsonrpc":"2.0","method":"session/update","params":{"update":{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"key launch-secret-value-123 and ` + ghToken + `"}}}}`},
		{"agent", `{"jsonrpc":"2.0","method":"session/update","params":{"update":{"sessionUpdate":"tool_call","toolCallId":"c1","title":"curl","kind":"execute","status":"pending","rawInput":{"url":"http://x","password":"hunter2","nested":{"Authorization":"Bearer abc"}}}}}`},
		{"agent", `{"jsonrpc":"2.0","method":"session/update","params":{"update":{"sessionUpdate":"tool_call_update","toolCallId":"c1","content":[]}}}`},
		{"agent", `{"jsonrpc":"2.0","method":"session/update","params":{"update":{"sessionUpdate":"tool_call_update","toolCallId":"c1","status":"completed","rawOutput":{"stdout":"ok"}}}}`},
		{"agent", `{"jsonrpc":"2.0","id":"h1","result":{"stopReason":"end_turn"}}`},
	}
	for _, f := range frames {
		e.Observe(f.from, []byte(f.line))
	}
	closeExporter(t, e)

	recs := mem.all()
	var events []string
	for _, r := range recs {
		a := attrs(r)
		events = append(events, a[AttrConvEvent])
		assert.Equal(t, "js-s1", a[AttrSession])
		assert.Equal(t, "dev", a[AttrPool])
		assert.Equal(t, "js-t1", a[AttrTicket], "ticket from prompt _meta")
	}
	require.Equal(t, []string{"prompt", "agent_message", "tool_call", "tool_result", "turn_end"}, events,
		"chunks joined, status-less tool updates skipped")

	assert.Equal(t, "do it", recs[0].Body().AsString())
	msg := recs[1].Body().AsString()
	assert.True(t, strings.HasPrefix(msg, "hello key "), msg)
	assert.NotContains(t, msg, "launch-secret-value-123", "launch env value redacted")
	assert.NotContains(t, msg, ghToken, "gitleaks finding redacted")
	assert.Contains(t, msg, redacted)

	in := attrs(recs[2])[AttrToolInput]
	assert.Contains(t, in, `"url":"http://x"`)
	assert.NotContains(t, in, "hunter2", "secret-named key redacted")
	assert.NotContains(t, in, "Bearer abc", "nested secret-named key redacted")
	assert.Equal(t, `{"stdout":"ok"}`, attrs(recs[3])[AttrToolOutput])
	assert.Equal(t, "end_turn", attrs(recs[4])[AttrStopReason])
}

func TestConvExportTruncation(t *testing.T) {
	e, mem := newTestExporter(t, ConvConfig{FieldCap: 64})
	long := strings.Repeat("é", 100) // 200 bytes, 2-byte runes
	e.Observe("controller", []byte(`{"jsonrpc":"2.0","id":1,"method":"session/prompt","params":{"prompt":[{"type":"text","text":"`+long+`"}]}}`))
	e.Observe("controller", []byte(`{"jsonrpc":"2.0","id":2,"method":"session/prompt","params":{"prompt":[{"type":"text","text":"short"}]}}`))
	closeExporter(t, e)
	recs := mem.all()
	require.Len(t, recs, 2)
	body := recs[0].Body().AsString()
	assert.Len(t, body, 64)
	assert.Equal(t, strings.Repeat("é", 32), body, "cut on a rune boundary")
	assert.Equal(t, "true", attrs(recs[0])[AttrConvTruncated])
	_, ok := attrs(recs[1])[AttrConvTruncated]
	assert.False(t, ok)
	assert.Empty(t, attrs(recs[1])[AttrTicket], "no _meta, no ticket")
	_, truncated := e.TakeCounts()
	assert.Equal(t, int64(1), truncated)
}

// TestConvExportOverflow stalls the exporter so the queue fills: Observe
// never blocks, and the overflow is counted as dropped.
func TestConvExportOverflow(t *testing.T) {
	stall := &memExporter{block: make(chan struct{})}
	e, _ := newTestExporter(t, ConvConfig{Exporter: stall})
	line := []byte(`{"jsonrpc":"2.0","method":"session/update","params":{"update":{"sessionUpdate":"tool_call","toolCallId":"c","title":"t","status":"pending"}}}`)
	done := make(chan struct{})
	go func() {
		for range 200_000 {
			e.Observe("agent", line)
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("Observe blocked on a stalled exporter")
	}
	dropped, _ := e.TakeCounts()
	assert.Positive(t, dropped)
	close(stall.block)
	closeExporter(t, e)
}

func TestSecretEnvValues(t *testing.T) {
	got := secretEnvValues([]string{"PATH=/usr/bin:/bin", "GH_TOKEN=abcdefgh", "DB_PASSWORD=abcdefghijkl", "API_KEY=short", "NAME=jobshop-user"})
	assert.Equal(t, []string{"abcdefghijkl", "abcdefgh"}, got, "secret-named, 8+ bytes, longest first")
}
