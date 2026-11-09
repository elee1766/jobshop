package desk_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/controltest"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/transport/desk"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// This boundary catches audit changes that never invalidate Desk before a ticket exists.
func TestAutomationAuditLive(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	arrived, release := make(chan struct{}), make(chan struct{})
	var released sync.Once
	t.Cleanup(func() { released.Do(func() { close(release) }) })
	var requests atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		last := requests.Add(1) == 12
		if last {
			close(arrived)
			select {
			case <-release:
			case <-r.Context().Done():
				return
			}
		}
		body := "ok"
		if last {
			body = strings.Repeat("x", 70<<10)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"authorization": r.Header.Get("Authorization"), "body": body})
	}))
	defer upstream.Close()
	defer released.Do(func() { close(release) })
	root := controltest.NewRoot(t, fmt.Sprintf(`pack: pack.yaml
projects: {js: {}}
agents: {fake: {command: "false"}}
automation:
  http: true
  credentials:
    audit:
      secret: credential
      origins: [%q]
      header: Authorization
      prefix: "Bearer "
`, upstream.URL))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "credential"), []byte("audit-fixture-key"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "pack.yaml"), []byte(`pools:
  dev: {project: js, agent: fake, max_sessions: 0, dir: "."}
workflows:
  fixed:
    project: js
    steps: [{id: work, title: Work, pool: dev, prompt: "Review event"}]
triggers:
  incoming:
    on: [{source: generic, event: tasks}]
    handler: hooks.star:handle
`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "hooks.star"), fmt.Appendf(nil, `def handle(event, args):
    for i in range(12):
        http.get(%q, auth="audit")
    return [start_run(workflow="fixed", key=event.delivery)]
`, upstream.URL), 0o600))
	_, err := config.Load(root)
	require.NoError(t, err)
	controller := controltest.StartWith(t, root)
	server := httptest.NewServer(desk.New(desk.Config{Root: root.Dir}, desk.Socket{Path: root.AdminSocket()}, nil).Handler())
	defer server.Close()
	type message struct {
		kind string
		data json.RawMessage
	}
	connect := func(query string) (<-chan message, func()) {
		t.Helper()
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/live"+query, nil)
		require.NoError(t, err)
		response, err := http.DefaultClient.Do(request)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)
		messages := make(chan message, 32)
		go func() {
			defer close(messages)
			scanner := bufio.NewScanner(response.Body)
			scanner.Buffer(make([]byte, 4096), 2<<20)
			var kind string
			for scanner.Scan() {
				line := scanner.Text()
				if value, ok := strings.CutPrefix(line, "event: "); ok {
					kind = value
				}
				if value, ok := strings.CutPrefix(line, "data: "); ok {
					select {
					case messages <- message{kind, json.RawMessage(value)}:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
		return messages, func() { _ = response.Body.Close() }
	}
	next := func(messages <-chan message, kind string) json.RawMessage {
		t.Helper()
		for {
			select {
			case msg, ok := <-messages:
				require.True(t, ok, "live stream ended before %s", kind)
				if msg.kind == kind {
					return msg.data
				}
			case <-ctx.Done():
				t.Fatalf("no pushed %s state: %v", kind, ctx.Err())
			}
		}
	}
	messages, disconnect := connect("")
	defer disconnect()
	var initial struct {
		Automation desk.AutomationList `json:"automation"`
	}
	require.NoError(t, json.Unmarshal(next(messages, "system"), &initial))
	require.Empty(t, initial.Automation.Items)

	spoolPath := filepath.Join(root.Dir, "audit-spool.db")
	spool, err := ingest.OpenSpool(ctx, spoolPath)
	require.NoError(t, err)
	defer func() { _ = spool.Close() }()
	_, err = ingest.Put(ctx, spool, config.IngestEndpoint{Name: "local", Source: ingest.SourceGeneric}, "desk-event", "tasks", []byte(`{"task":"review"}`), controller.Clock.Now())
	require.NoError(t, err)
	controller.Call(rpc.MethodIngestDrain, rpc.IngestDrainParams{Spool: spoolPath}, nil)
	controller.Pass(1)
	select {
	case <-arrived:
	case <-ctx.Done():
		t.Fatal("hook did not reach the held HTTP call")
	}
	var listed struct {
		Automation desk.AutomationList `json:"automation"`
	}
	for len(listed.Automation.Items) == 0 {
		require.NoError(t, json.Unmarshal(next(messages, "system"), &listed))
	}
	require.Len(t, listed.Automation.Items, 1)
	id := listed.Automation.Items[0].Key
	assert.Equal(t, "generic:desk-event", listed.Automation.Items[0].Event.Delivery)
	assert.NotContains(t, string(nextSummary(t, listed.Automation)), "payload")
	var runs []domain.Ticket
	controller.Call(rpc.MethodTicketList, rpc.ListParams{Kind: domain.KindRun}, &runs)
	require.Empty(t, runs, "audit must be available before the hook creates tickets")
	disconnect()
	messages, disconnect = connect("?audit=" + id)
	defer disconnect()
	var trace desk.AutomationAudit
	require.NoError(t, json.Unmarshal(next(messages, "automation"), &trace))
	require.Len(t, trace.Calls, 12)
	assert.Equal(t, "UNKNOWN", trace.Calls[11].State)
	for i, call := range trace.Calls {
		var value struct {
			Observation struct{ Index int } `json:"observation"`
		}
		require.NoError(t, json.Unmarshal(call.Value, &value))
		assert.Equal(t, i+1, value.Observation.Index)
		assert.NotContains(t, string(call.Value), "audit-fixture-key")
	}
	released.Do(func() { close(release) })
	for trace.Invocation.State != "done" || len(trace.Attempts) == 0 || trace.Attempts[0].State != "done" {
		require.NoError(t, json.Unmarshal(next(messages, "automation"), &trace))
	}
	assert.Equal(t, "completed", trace.Calls[11].State)
	assert.Contains(t, string(trace.Calls[11].Value), `"artifact"`)
	assert.Less(t, len(trace.Calls[11].Value), 64<<10)
	assert.NotContains(t, string(trace.Invocation.Value), "audit-fixture-key")
	var result struct {
		Targets []string `json:"targets"`
	}
	require.NoError(t, json.Unmarshal(trace.Invocation.Value, &result))
	require.NotEmpty(t, result.Targets)
	assert.NotContains(t, string(trace.Invocation.Value), `"plan"`)
	var shown rpc.ShowResult
	controller.Call(rpc.MethodTicketShow, rpc.IDParams{ID: result.Targets[0]}, &shown)
	assert.Equal(t, domain.KindRun, shown.Ticket.Kind)
}

func nextSummary(t *testing.T, value desk.AutomationList) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	require.NoError(t, err)
	return data
}
