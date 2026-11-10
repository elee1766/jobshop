package desk

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// fakeCtl answers admin calls from canned results and records them.
type fakeCtl struct {
	results map[string]any
	calls   []string
	events  []rpc.Event
	session []rpc.SessionEvent
	since   int64
	tail    int64
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
	f.since, f.tail = p.SinceSeq, p.Tail
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
	old := time.Now().Add(-3 * time.Hour)
	ctl := &fakeCtl{results: map[string]any{
		rpc.MethodAttention: rpc.AttentionResult{Items: []rpc.AttentionItem{
			{Kind: rpc.AttentionTicket, ID: "t1", Title: "deploy:rollback", Since: old, Repo: "acme/w", Ref: "#7"},
			{Kind: rpc.AttentionWait, ID: "w1", Title: "wait", Owner: "bob"},
			{Kind: rpc.AttentionQuestion, ID: "q1", Title: "Question: which base?", Context: "ran tests\nboth pass", Options: []domain.QuestionOption{{ID: 1, Title: "main"}}},
		}},
	}}
	h := New(Config{}, ctl, nil).Handler()
	var items []AttentionItem
	get(t, h, "/api/attention", &items)
	// The controller's list, in its order, with no items of Desk's own.
	assert.Equal(t, []AttentionItem{
		{Kind: "ticket", ID: "t1", Title: "deploy:rollback", Why: "acme/w #7", Since: items[0].Since, Ref: "#7"},
		{Kind: "wait", ID: "w1", Title: "wait", Why: "owner bob"},
		{Kind: "question", ID: "q1", Title: "Question: which base?", Context: "ran tests\nboth pass", Options: []domain.QuestionOption{{ID: 1, Title: "main"}}},
	}, items)
	assert.Equal(t, []string{rpc.MethodAttention + " {}"}, ctl.calls)
}

func TestAttentionEmpty(t *testing.T) {
	h := New(Config{}, &fakeCtl{results: map[string]any{}}, nil).Handler()
	w := get(t, h, "/api/attention", nil)
	assert.Equal(t, "[]\n", w.Body.String())
}

func TestAPIError(t *testing.T) {
	ctl := &fakeCtl{results: map[string]any{rpc.MethodTicketShow: fmt.Errorf("x: %w", domain.ErrNotFound)}}
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
		rpc.MethodRunList:     []domain.Ticket{{ID: "r1"}},
		rpc.MethodBatchList:   []domain.Ticket{{ID: "b1"}},
		rpc.MethodTriggerList: []rpc.TriggerInfo{{Name: "z"}, {Name: "a"}},
		rpc.MethodVersion:     rpc.VersionResult{Version: "v", Commit: "c"},
		rpc.MethodTicketShow:  rpc.ShowResult{Ticket: domain.Ticket{ID: "r1", Kind: domain.KindRun}},
		rpc.MethodRunShow:     rpc.RunResult{Steps: []domain.Ticket{{ID: "s1"}}},
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

	var tv TicketView
	get(t, h, "/api/ticket/r1", &tv)
	require.NotNil(t, tv.Run)
	assert.Equal(t, "s1", tv.Run.Steps[0].ID)
	require.NotNil(t, tv.Trace)

	var ss []rpc.SessionInfo
	get(t, h, "/api/sessions", &ss)
	assert.NotNil(t, ss)
}

// TestTraceTimeline (#1103): the trace snapshot carries the timeline only
// when asked, since it replays every node's history.
func TestTraceTimeline(t *testing.T) {
	ctl := &fakeCtl{results: map[string]any{
		rpc.MethodTicketGraph:    rpc.GraphResult{Root: "r1"},
		rpc.MethodTicketTimeline: rpc.TimelineResult{Root: "r1", Rows: []rpc.TimelineRow{{ID: "r1"}}},
		rpc.MethodTicketShow:     rpc.ShowResult{Ticket: domain.Ticket{ID: "r1"}},
		rpc.MethodTicketTrace:    rpc.TraceNode{ID: "r1"},
	}}
	h := New(Config{}, ctl, nil).Handler()
	var plain TicketView
	get(t, h, "/api/trace/r1", &plain)
	assert.Nil(t, plain.Timeline)
	assert.NotContains(t, ctl.calls, rpc.MethodTicketTimeline+` {"id":"r1","limit":2000}`)
	var tv TicketView
	get(t, h, "/api/trace/r1?timeline=1", &tv)
	require.NotNil(t, tv.Timeline)
	assert.Equal(t, "r1", tv.Timeline.Rows[0].ID)
}

// TestHolderSessions: api/work?detail=1 carries each run's steps, last
// comment and the holder of its in-progress step, and api/ticket fills
// trace nodes with their holder session (JOB-54).
func TestHolderSessions(t *testing.T) {
	step := domain.Ticket{ID: "s1", Status: domain.StatusInProgress, Holder: "js-s-abc123"}
	ctl := &fakeCtl{results: map[string]any{
		rpc.MethodRunList: []domain.Ticket{{ID: "r1", Kind: domain.KindRun}},
		rpc.MethodRunShow: rpc.RunResult{Steps: []domain.Ticket{step, {ID: "s2", Status: domain.StatusOpen}}},
		rpc.MethodTicketShow: rpc.ShowResult{
			Ticket:   domain.Ticket{ID: "r1", Kind: domain.KindRun},
			Comments: []domain.Comment{{ID: 1, Body: "a"}, {ID: 2, Body: "b"}},
		},
		rpc.MethodTicketTrace: rpc.TraceNode{ID: "r1", Children: []rpc.TraceNode{{ID: "s1", Holder: "js-s-abc123"}}},
		rpc.MethodSessionList: []rpc.SessionInfo{
			{ID: "js-s-abc123", DisplayName: "dev worker", Alive: true, Busy: true, Tool: "go test", Ticket: &domain.Ticket{Pool: "dev"}},
			{ID: "js-s-other"},
		},
	}}
	h := New(Config{}, ctl, nil).Handler()

	var w Work
	get(t, h, "/api/work?state=running&detail=1", &w)
	require.Contains(t, w.Detail, "r1")
	assert.Len(t, w.Detail["r1"].Steps, 2)
	require.NotNil(t, w.Detail["r1"].Comment)
	assert.Equal(t, "b", w.Detail["r1"].Comment.Body)
	want := rpc.HolderInfo{ID: "js-s-abc123", DisplayName: "dev worker", Pool: "dev", Alive: true, Busy: true, Tool: "go test"}
	assert.Equal(t, map[string]rpc.HolderInfo{"js-s-abc123": want}, w.Sessions)

	var plain Work
	get(t, h, "/api/work", &plain)
	assert.Nil(t, plain.Detail)
	assert.Nil(t, plain.Sessions)

	var tv TicketView
	get(t, h, "/api/ticket/r1", &tv)
	assert.Equal(t, want, tv.Sessions["js-s-abc123"])
	require.NotNil(t, tv.Trace)
	require.NotNil(t, tv.Trace.Children[0].Session)
	assert.Equal(t, "dev worker", tv.Trace.Children[0].Session.DisplayName)
	assert.Nil(t, tv.Trace.Session)
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

// needUI skips a test that reads the built UI when the bundle is not
// embedded. With JOBSHOP_REQUIRE_DESK_UI set (CI) a missing bundle fails.
func needUI(t *testing.T) {
	t.Helper()
	if Built() {
		return
	}
	if os.Getenv("JOBSHOP_REQUIRE_DESK_UI") != "" {
		t.Fatal("Desk UI is not built: run npm --prefix ui run build")
	}
	t.Skip("Desk UI is not built (npm --prefix ui run build)")
}

func TestStatic(t *testing.T) {
	needUI(t)
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
		session: []rpc.SessionEvent{{Seq: 450, Type: rpc.EventText, Text: "<b>hi</b>"}},
	}
	h := New(Config{}, ctl, nil).Handler()
	w := get(t, h, "/api/session/s1/stream?tail=100", nil)
	// Tail goes to the controller; no session list (#614).
	assert.Equal(t, int64(0), ctl.since)
	assert.Equal(t, int64(100), ctl.tail)
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

// TestEscalations: only tickets with a controller-opened chat are listed,
// open ones first (oldest first), then the latest closed ones.
func TestEscalations(t *testing.T) {
	meta := func(session, cause string) json.RawMessage {
		return json.RawMessage(`{"escalation":{"key":"escalation:x","session":"` + session + `","cause":"` + cause + `","held":"h1"}}`)
	}
	t0 := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	ctl := &fakeCtl{results: map[string]any{
		rpc.MethodTicketList + ":open": []domain.Ticket{
			{ID: "e2", Title: "newer", Body: "why two\nmore", Meta: meta("s2", "wait"), CreatedAt: t0.Add(time.Hour)},
			{ID: "e1", Title: "older", Meta: meta("s1", "run"), CreatedAt: t0},
			{ID: "nochat", Title: "no chat", CreatedAt: t0},
		},
		rpc.MethodTicketList + ":closed": []domain.Ticket{
			{ID: "e0", Title: "done", Status: domain.StatusClosed, Resolution: "done", Meta: meta("s0", "label"), CreatedAt: t0, ClosedAt: t0.Add(2 * time.Hour)},
		},
	}}
	h := New(Config{}, ctl, nil).Handler()
	var got []Escalation
	get(t, h, "/api/escalations", &got)
	var ids []string
	for _, e := range got {
		ids = append(ids, e.Ticket)
	}
	assert.Equal(t, []string{"e1", "e2", "e0"}, ids)
	assert.True(t, got[0].Open)
	assert.Equal(t, "run", got[0].Cause)
	assert.Equal(t, "s2", got[1].Session)
	assert.Equal(t, "why two", got[1].Why)
	assert.False(t, got[2].Open)
	assert.Equal(t, "done", got[2].Resolution)
	assert.Contains(t, strings.Join(ctl.calls, "\n"), `"newest":true`)

	h = New(Config{}, &fakeCtl{results: map[string]any{}}, nil).Handler()
	assert.Equal(t, "[]\n", get(t, h, "/api/escalations", nil).Body.String())
}

// assetPath is an asset file of the embedded build.
func assetPath(t *testing.T) string {
	t.Helper()
	ents, err := dist.ReadDir("dist/assets")
	require.NoError(t, err)
	require.NotEmpty(t, ents)
	return "assets/" + ents[0].Name()
}

func TestBasePath(t *testing.T) {
	needUI(t)
	ctl := &fakeCtl{results: map[string]any{}}
	h := New(Config{BasePath: "/desk/"}, ctl, nil).Handler()
	asset := assetPath(t)
	// Prefix kept by the proxy, and prefix stripped by it.
	for name, prefix := range map[string]string{"kept": "/desk/", "stripped": "/"} {
		t.Run(name, func(t *testing.T) {
			w := get(t, h, prefix, nil)
			require.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), `<base href="/desk/">`)
			assert.Contains(t, w.Body.String(), `<div id="app">`)
			assert.NotContains(t, w.Body.String(), `"/assets/`, "asset URLs are relative")
			var items []AttentionItem
			get(t, h, prefix+"api/attention", &items)
			w = get(t, h, prefix+asset, nil)
			assert.Equal(t, http.StatusOK, w.Code, asset)
			assert.Contains(t, w.Header().Get("Cache-Control"), "immutable")
			w = get(t, h, prefix+"?view=work", nil)
			assert.Contains(t, w.Body.String(), `<base href="/desk/">`)
		})
	}
	w := get(t, h, "/desk?view=work", nil)
	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Equal(t, "/desk/?view=work", w.Header().Get("Location"))
	assert.Contains(t, w.Header().Get("Content-Security-Policy"), "base-uri 'self'")
}

func TestForwardedPrefix(t *testing.T) {
	proxies, err := ParseTrustedProxies([]string{"192.0.2.0/24", "::1"})
	require.NoError(t, err)
	page := func(h http.Handler, remote, prefix string) string {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Host = "127.0.0.1"
		req.RemoteAddr = remote
		if prefix != "" {
			req.Header.Set(ForwardedPrefixHeader, prefix)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		return w.Body.String()
	}
	h := New(Config{TrustedProxies: proxies}, &fakeCtl{}, nil).Handler()
	assert.Contains(t, page(h, "192.0.2.1:1234", "/ops/desk"), `<base href="/ops/desk/">`)
	assert.Contains(t, page(h, "[::1]:1234", "/x/"), `<base href="/x/">`)
	assert.Contains(t, page(h, "198.51.100.1:1234", "/ops/"), `<base href="/">`, "untrusted peer")
	assert.Contains(t, page(h, "192.0.2.1:1234", "//evil.example/"), `<base href="/">`, "invalid prefix")
	assert.Contains(t, page(h, "192.0.2.1:1234", ""), `<base href="/">`)
	// The flag wins.
	h = New(Config{BasePath: "desk", TrustedProxies: proxies}, &fakeCtl{}, nil).Handler()
	assert.Contains(t, page(h, "192.0.2.1:1234", "/ops/"), `<base href="/desk/">`)
	_, err = ParseTrustedProxies([]string{"nope"})
	assert.Error(t, err)
}

func TestCleanBasePath(t *testing.T) {
	for in, want := range map[string]string{
		"": "/", "/": "/", "desk": "/desk/", "/desk": "/desk/", "/a/b/": "/a/b/",
		"//evil": "", "/a//b/": "", "/a b/": "", `/"x/`: "",
	} {
		assert.Equal(t, want, CleanBasePath(in), in)
	}
}

// --favicon replaces the built-in icon link and is served at /favicon.
func TestFavicon(t *testing.T) {
	needUI(t)
	ctl := &fakeCtl{results: map[string]any{}}
	w := get(t, New(Config{}, ctl, nil).Handler(), "/", nil)
	assert.Contains(t, w.Body.String(), `href="icon.svg"`)

	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1 1"/>`)
	h := New(Config{Favicon: svg}, ctl, nil).Handler()
	w = get(t, h, "/", nil)
	assert.Contains(t, w.Body.String(), `<link rel="icon" href="favicon" type="image/svg+xml" />`)
	assert.NotContains(t, w.Body.String(), `href="icon.svg"`)
	w = get(t, h, "/favicon", nil)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "image/svg+xml", w.Header().Get("Content-Type"))
	assert.Equal(t, svg, w.Body.Bytes())
}

// TestStaticCompression: assets are gzipped for a client that accepts it
// and sent as is otherwise; streams and the API are never compressed (#1088).
func TestStaticCompression(t *testing.T) {
	needUI(t)
	s := New(Config{}, &fakeCtl{events: []rpc.Event{{Kind: rpc.EventKindReady}}}, nil)
	h := s.Handler()
	entries, err := fs.ReadDir(s.ui, "assets")
	require.NoError(t, err)
	var asset string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".js") {
			asset = "/assets/" + e.Name()
		}
	}
	require.NotEmpty(t, asset, "built UI has a script")
	raw, err := fs.ReadFile(s.ui, strings.TrimPrefix(asset, "/"))
	require.NoError(t, err)

	fetch := func(path, encoding string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Host = "127.0.0.1:7480"
		if encoding != "" {
			req.Header.Set("Accept-Encoding", encoding)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		return w
	}

	w := fetch(asset, "gzip")
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
	assert.Contains(t, w.Header().Values("Vary"), "Accept-Encoding")
	assert.Contains(t, w.Header().Get("Cache-Control"), "immutable", "caching headers survive")
	zr, err := gzip.NewReader(w.Body)
	require.NoError(t, err)
	got, err := io.ReadAll(zr)
	require.NoError(t, err)
	assert.Equal(t, raw, got)
	assert.Less(t, w.Body.Len(), len(raw))

	w = fetch(asset, "")
	assert.Empty(t, w.Header().Get("Content-Encoding"))
	assert.Equal(t, raw, w.Body.Bytes())

	for _, p := range []string{"/api/events", "/api/attention"} {
		w = fetch(p, "gzip")
		assert.Empty(t, w.Header().Get("Content-Encoding"), p)
	}
}

// TestTraceTimelineHolders (JOB-184): the holders of open working bars join
// the one session.list call, also when no trace node or run step names them,
// and a controller-driven session whose host is gone is marked so.
func TestTraceTimelineHolders(t *testing.T) {
	open := rpc.TimelineSegment{State: rpc.TimelineWorking, Actor: "js-s-live"}
	closed := rpc.TimelineSegment{State: rpc.TimelineWorking, Actor: "js-s-old", End: time.Now()}
	ctl := &fakeCtl{results: map[string]any{
		rpc.MethodTicketGraph: rpc.GraphResult{Root: "r1"},
		rpc.MethodTicketTimeline: rpc.TimelineResult{Root: "r1", Rows: []rpc.TimelineRow{
			{ID: "r1", Segments: []rpc.TimelineSegment{closed, open}},
			{ID: "r2", Segments: []rpc.TimelineSegment{{State: rpc.TimelineWorking, Actor: "js-s-dead"}, {State: rpc.TimelineWorking, Actor: "js-s-ext"}}},
		}},
		rpc.MethodTicketShow:  rpc.ShowResult{Ticket: domain.Ticket{ID: "r1"}},
		rpc.MethodTicketTrace: rpc.TraceNode{ID: "r1"},
		rpc.MethodSessionList: []rpc.SessionInfo{
			{ID: "js-s-live", Alive: true, Busy: true, Tool: "go test"},
			{ID: "js-s-dead", HostPid: 42, Exit: "crashed"},
			{ID: "js-s-ext"},
			{ID: "js-s-old", Alive: true},
		},
	}}
	h := New(Config{}, ctl, nil).Handler()
	var tv TicketView
	get(t, h, "/api/trace/r1?timeline=1", &tv)
	assert.Equal(t, rpc.HolderInfo{ID: "js-s-live", Alive: true, Busy: true, Tool: "go test"}, tv.Sessions["js-s-live"])
	assert.True(t, tv.Sessions["js-s-dead"].HostLost)
	assert.False(t, tv.Sessions["js-s-ext"].HostLost, "an external session has no host to lose")
	assert.NotContains(t, tv.Sessions, "js-s-old", "a closed bar's holder is not looked up")
	n := 0
	for _, c := range ctl.calls {
		if strings.HasPrefix(c, rpc.MethodSessionList) {
			n++
		}
	}
	assert.Equal(t, 1, n, "one session.list call")
}

// TestTraceWholeGraph (JOB-190): a trace opened on a sub-item returns the
// graph from the top ticket with the requested ticket inspected, and an
// explicit ?inspect= still wins.
func TestTraceWholeGraph(t *testing.T) {
	ctl := &fakeCtl{results: map[string]any{
		rpc.MethodTicketGraph:    rpc.GraphResult{Root: "a", Requested: "b"},
		rpc.MethodTicketTimeline: rpc.TimelineResult{Root: "a", Requested: "b", Rows: []rpc.TimelineRow{{ID: "a"}, {ID: "b"}}},
		rpc.MethodTicketShow:     rpc.ShowResult{Ticket: domain.Ticket{ID: "a"}},
		rpc.MethodTicketTrace:    rpc.TraceNode{ID: "a"},
	}}
	h := New(Config{}, ctl, nil).Handler()
	shown := func() []string {
		var out []string
		for _, call := range ctl.calls {
			if strings.HasPrefix(call, rpc.MethodTicketShow+" ") {
				out = append(out, strings.TrimPrefix(call, rpc.MethodTicketShow+" "))
			}
		}
		return out
	}

	var tv TicketView
	get(t, h, "/api/trace/b?timeline=1", &tv)
	require.NotNil(t, tv.Graph)
	assert.Equal(t, "a", tv.Graph.Root)
	assert.Equal(t, "a", tv.Ticket.ID, "the details of the root stay the base")
	require.NotNil(t, tv.Inspected, "the requested ticket is the selection")
	assert.Equal(t, []string{`{"id":"a"}`, `{"id":"b"}`}, shown())
	require.NotNil(t, tv.Timeline)
	assert.Equal(t, "a", tv.Timeline.Root)
	assert.Contains(t, ctl.calls, rpc.MethodTicketTimeline+` {"id":"b","limit":2000}`)

	ctl.calls = nil
	var other TicketView
	get(t, h, "/api/trace/b?inspect=c", &other)
	assert.Equal(t, []string{`{"id":"a"}`, `{"id":"c"}`}, shown())

	// Asking for the root selects nothing extra.
	ctl.calls = nil
	ctl.results[rpc.MethodTicketGraph] = rpc.GraphResult{Root: "a", Requested: "a"}
	var top TicketView
	get(t, h, "/api/trace/a", &top)
	assert.Nil(t, top.Inspected)
	assert.Equal(t, []string{`{"id":"a"}`}, shown())
}
