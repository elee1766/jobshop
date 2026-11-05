package desk

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// fakeCtl answers admin calls from canned results and records them.
type fakeCtl struct {
	results map[string]any
	calls   []string
	events  []rpc.Event
	session []rpc.SessionEvent
	since   int64
	actors  []string
}

func (f *fakeCtl) Call(ctx context.Context, method string, p, out any) error {
	f.actors = append(f.actors, rpc.ActorFrom(ctx))
	b, _ := json.Marshal(p)
	f.calls = append(f.calls, method+" "+string(b))
	key := method
	if method == rpc.MethodTicketList {
		var lp rpc.ListParams
		_ = json.Unmarshal(b, &lp)
		key += ":" + string(lp.Status)
	}
	if method == rpc.MethodRunList {
		var rp rpc.RunListParams
		_ = json.Unmarshal(b, &rp)
		key += ":" + rp.State
	}
	v, ok := f.results[key]
	if !ok {
		v, ok = f.results[method]
	}
	if !ok {
		return nil
	}
	if err, isErr := v.(error); isErr {
		return err
	}
	raw, _ := json.Marshal(v)
	return json.Unmarshal(raw, out)
}

func (f *fakeCtl) Events(_ context.Context, p rpc.EventsSubscribeParams, fn func(rpc.Event)) (rpc.EventsSubscribeResult, error) {
	f.since = p.SinceSeq
	for _, ev := range f.events {
		fn(ev)
	}
	return rpc.EventsSubscribeResult{Lagged: true}, nil
}

func (f *fakeCtl) Subscribe(_ context.Context, p rpc.SessionSubscribeParams, fn func(rpc.SessionEvent)) (rpc.SessionSubscribeResult, error) {
	f.since = p.SinceSeq
	for _, ev := range f.session {
		fn(ev)
	}
	return rpc.SessionSubscribeResult{Seq: 9, Closed: true}, nil
}

func get(t *testing.T, h http.Handler, path string, out any) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Host = "127.0.0.1:7480"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if out != nil {
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), out))
	}
	return w
}

func TestAttention(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "deployed"), []byte("aaaa\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "failed"), []byte("bbbbbbbbbb\n"), 0o600))
	old := time.Now().Add(-3 * time.Hour)
	mid := time.Now().Add(-time.Hour)
	ctl := &fakeCtl{results: map[string]any{
		rpc.MethodRunList + ":" + store.RunNeedsHuman: []store.Ticket{{ID: "r1", Title: "run", UpdatedAt: mid}},
		rpc.MethodTicketList + ":open":                []store.Ticket{{ID: "t1", Title: "escalation", CreatedAt: old}},
		rpc.MethodPoolList:                            rpc.PoolStatusResult{Quarantined: map[string]string{"dev": "agent not found"}},
		rpc.MethodWaitList: []rpc.WaitItem{
			{ID: "w1", Wait: store.Wait{Reason: "approval", Owner: "bob", TimedOut: true, Timeout: mid.Add(time.Minute)}},
			{ID: "w2", Wait: store.Wait{Reason: "not yet"}},
		},
	}}
	h := New(Config{DeployState: dir}, ctl, nil).Handler()
	var items []AttentionItem
	get(t, h, "/api/attention", &items)
	var got []string
	for _, it := range items {
		got = append(got, it.Kind+":"+it.ID)
	}
	// Oldest first; items without a time last.
	assert.Equal(t, []string{"ticket:t1", "run:r1", "wait:w1", "deploy:bbbbbbbbbb", "pool:dev"}, got)
	assert.Contains(t, strings.Join(ctl.calls, "\n"), `"label":"needs-human"`)
}

func TestAttentionEmpty(t *testing.T) {
	h := New(Config{}, &fakeCtl{results: map[string]any{}}, nil).Handler()
	w := get(t, h, "/api/attention", nil)
	assert.Equal(t, "[]\n", w.Body.String())
}

func TestAPIError(t *testing.T) {
	ctl := &fakeCtl{results: map[string]any{rpc.MethodTicketShow: fmt.Errorf("x: %w", store.ErrNotFound)}}
	h := New(Config{}, ctl, nil).Handler()
	w := get(t, h, "/api/ticket/js-1", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
	ctl.results[rpc.MethodTicketShow] = errors.New("boom")
	w = get(t, h, "/api/ticket/js-1", nil)
	assert.Equal(t, http.StatusBadGateway, w.Code)
	assert.Contains(t, w.Body.String(), "boom")
}

func TestViews(t *testing.T) {
	ctl := &fakeCtl{results: map[string]any{
		rpc.MethodRunList:     []store.Ticket{{ID: "r1"}},
		rpc.MethodBatchList:   []store.Ticket{{ID: "b1"}},
		rpc.MethodTriggerList: []rpc.TriggerInfo{{Name: "z"}, {Name: "a"}},
		rpc.MethodVersion:     rpc.VersionResult{Version: "v", Commit: "c"},
		rpc.MethodTicketShow:  rpc.ShowResult{Ticket: store.Ticket{ID: "r1", Kind: store.KindRun}},
		rpc.MethodRunShow:     rpc.RunResult{Steps: []store.Ticket{{ID: "s1"}}},
		rpc.MethodTicketTrace: rpc.TraceNode{ID: "r1"},
	}}
	h := New(Config{}, ctl, nil).Handler()

	var w Work
	get(t, h, "/api/work?state=running&limit=5", &w)
	assert.Equal(t, "r1", w.Runs[0].ID)
	assert.Equal(t, "b1", w.Batches[0].ID)
	assert.Contains(t, ctl.calls, rpc.MethodRunList+` {"state":"running","limit":5}`)

	var p Pools
	get(t, h, "/api/pools", &p)
	assert.Equal(t, "a", p.Triggers[0].Name)

	var d Deploys
	get(t, h, "/api/deploys", &d)
	assert.Equal(t, "c", d.Server.Commit)

	var tv TicketView
	get(t, h, "/api/ticket/r1", &tv)
	require.NotNil(t, tv.Run)
	assert.Equal(t, "s1", tv.Run.Steps[0].ID)
	require.NotNil(t, tv.Trace)

	var ss []rpc.SessionInfo
	get(t, h, "/api/sessions", &ss)
	assert.NotNil(t, ss)
}

func TestGuard(t *testing.T) {
	h := New(Config{Hosts: []string{"desk.tail"}}, &fakeCtl{results: map[string]any{}}, nil).Handler()
	for host, want := range map[string]int{
		"127.0.0.1:7480":  http.StatusOK,
		"localhost:7480":  http.StatusOK,
		"[::1]:7480":      http.StatusOK,
		"desk.tail":       http.StatusOK,
		"evil.example:80": http.StatusMisdirectedRequest,
		"10.0.0.5:7480":   http.StatusMisdirectedRequest,
		"desk.tail.evil":  http.StatusMisdirectedRequest,
	} {
		req := httptest.NewRequest(http.MethodGet, "/api/attention", nil)
		req.Host = host
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, want, w.Code, host)
		assert.Contains(t, w.Header().Get("Content-Security-Policy"), "script-src 'self'", host)
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	}
	// Writes need the CSRF header (none are routed yet; the guard runs first).
	req := httptest.NewRequest(http.MethodPost, "/api/attention", nil)
	req.Host = "127.0.0.1"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
	req.Header.Set(CSRFHeader, "1")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestLoopback(t *testing.T) {
	assert.True(t, Loopback("127.0.0.1:7480"))
	assert.True(t, Loopback("[::1]:1"))
	assert.True(t, Loopback("localhost:1"))
	assert.False(t, Loopback("0.0.0.0:7480"))
	assert.False(t, Loopback(":7480"))
	assert.False(t, Loopback("100.64.0.1:7480"))
}

func TestStatic(t *testing.T) {
	h := New(Config{}, &fakeCtl{}, nil).Handler()
	for _, p := range []string{"/", "/?view=work", "/no/such/path"} {
		w := get(t, h, p, nil)
		assert.Equal(t, http.StatusOK, w.Code, p)
		assert.Contains(t, w.Body.String(), `<div id="app">`, p)
		// The CSP forbids inline scripts, so the built page must have none.
		assert.NotContains(t, w.Body.String(), "<script>", p)
	}
}

func readSSE(t *testing.T, body string) []string {
	t.Helper()
	var out []string
	sc := bufio.NewScanner(strings.NewReader(body))
	for sc.Scan() {
		if line, ok := strings.CutPrefix(sc.Text(), "event: "); ok {
			out = append(out, line)
		}
	}
	return out
}

func TestEventsSSE(t *testing.T) {
	ctl := &fakeCtl{events: []rpc.Event{{Seq: 4, Kind: rpc.EventKindTicket, ID: "t1"}, {Kind: rpc.EventKindReady}}}
	h := New(Config{}, ctl, nil).Handler()
	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	req.Host = "localhost"
	req.Header.Set("Last-Event-ID", "3")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
	assert.Equal(t, int64(3), ctl.since)
	assert.Equal(t, []string{"ticket", "ready", "lagged"}, readSSE(t, w.Body.String()))
	assert.Contains(t, w.Body.String(), "id: 4\n")
}

func TestSessionStream(t *testing.T) {
	ctl := &fakeCtl{
		results: map[string]any{rpc.MethodSessionList: []rpc.SessionInfo{{ID: "s1", Seq: 500}}},
		session: []rpc.SessionEvent{{Seq: 450, Type: rpc.EventText, Text: "<b>hi</b>"}},
	}
	h := New(Config{}, ctl, nil).Handler()
	w := get(t, h, "/api/session/s1/stream?tail=100", nil)
	assert.Equal(t, int64(400), ctl.since)
	assert.Equal(t, []string{"session", "end"}, readSSE(t, w.Body.String()))
}

func TestActivity(t *testing.T) {
	ctl := &fakeCtl{results: map[string]any{rpc.MethodActivityList: rpc.ActivityListResult{
		Items:   []rpc.ActivityItem{{ID: "gh:o/r:1", Source: rpc.ActivityGitHub, Repo: "o/r", Kind: "comment", Actor: "elee1766", Tickets: []string{"js-1"}}},
		Sources: []string{"github:o/r"},
	}}}
	h := New(Config{}, ctl, nil).Handler()
	var res rpc.ActivityListResult
	get(t, h, "/api/activity?since=1h&repo=o/r&bots=humans&ticket=js-1&limit=5", &res)
	require.Len(t, res.Items, 1)
	assert.Equal(t, []string{"js-1"}, res.Items[0].Tickets)
	assert.Contains(t, ctl.calls[0], `"since":3600000000000`)
	assert.Contains(t, ctl.calls[0], `"bots":"humans"`)
	assert.Contains(t, ctl.calls[0], `"limit":5`)
	w := get(t, h, "/api/activity?since=soon", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
