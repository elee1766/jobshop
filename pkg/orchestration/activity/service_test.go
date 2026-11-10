package activity

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/integrations/apicall"
	"github.com/gfx-labs/jobshop/pkg/integrations/github"
	"github.com/gfx-labs/jobshop/pkg/integrations/linear"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// fakeHost is a Host with a settable clock and recorded callbacks.
type fakeHost struct {
	mu       sync.Mutex
	cfg      *config.Config
	now      time.Time
	inbox    []rpc.ActivityItem
	ingests  int
	admitted []rpc.ActivityItem
	changed  int
	svc      *Service
	// admit makes Inbox admit what it receives, as the controller's
	// ingest pass does with no hooks.
	admit bool
}

func (h *fakeHost) Config() *config.Config {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.cfg
}

func (h *fakeHost) Now() time.Time {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.now
}

func (h *fakeHost) advance(d time.Duration) {
	h.mu.Lock()
	h.now = h.now.Add(d)
	h.mu.Unlock()
}

func (h *fakeHost) After(time.Duration) <-chan time.Time { return make(chan time.Time) }

func (h *fakeHost) Inbox(_ context.Context, items []rpc.ActivityItem) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.inbox = append(h.inbox, items...)
}

func (h *fakeHost) Ingest(ctx context.Context) {
	h.mu.Lock()
	h.ingests++
	items := h.inbox
	h.inbox = nil
	admit := h.admit
	h.mu.Unlock()
	if admit {
		h.svc.Admit(ctx, items)
	}
}

func (h *fakeHost) Admitted(_ context.Context, items []rpc.ActivityItem) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.admitted = append(h.admitted, items...)
}

func (h *fakeHost) Changed(time.Time) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.changed++
}

// fakeClients serves no identities: the host's token path.
type fakeClients struct{}

func (fakeClients) Config() *config.Config { return &config.Config{} }
func (fakeClients) Now() time.Time         { return time.Now() }
func (fakeClients) GitHub(string) *github.Client {
	panic("fakeClients: tests pass the client to fetchGitHub")
}
func (fakeClients) GitHubSelf(context.Context, string) string { return "" }
func (fakeClients) Linear(string) (*linear.Client, error) {
	return nil, fmt.Errorf("no linear")
}

type fakeSource struct {
	mu    sync.Mutex
	gh    map[string][]rpc.ActivityItem
	calls map[string]int
	err   error
}

func (f *fakeSource) GitHub(_ context.Context, repo string, st *RepoPoll) ([]rpc.ActivityItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls[repo]++
	if f.err != nil {
		return nil, f.err
	}
	st.EventsETag = "etag-" + repo
	return f.gh[repo], nil
}

func (f *fakeSource) Linear(context.Context, string, time.Time) ([]rpc.ActivityItem, error) {
	return nil, nil
}

func newService(t *testing.T) (*Service, *fakeHost, *fakeSource) {
	t.Helper()
	cfg := &config.Config{
		Activity: &config.Activity{Window: "24h", MaxItems: 3, Repos: []string{"o/a", "o/b"}},
	}
	h := &fakeHost{cfg: cfg, now: time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC), admit: true}
	s := New(h, slog.New(slog.DiscardHandler), t.TempDir())
	h.svc = s
	return s, h, &fakeSource{gh: map[string][]rpc.ActivityItem{}, calls: map[string]int{}}
}

func item(repo, id, kind, actor, ref string, at time.Time) rpc.ActivityItem {
	return rpc.ActivityItem{ID: "gh:" + repo + ":" + id, Source: rpc.ActivityGitHub, Repo: repo, Kind: kind, Actor: actor, Bot: strings.HasSuffix(actor, "[bot]"), At: at, Title: kind, Ref: ref}
}

// TestPollWindowBoundAndRestart: items are deduplicated by id, the window
// and per-source bound trim them, and a new Service on the same directory
// reads the cache back with its ETags.
func TestPollWindowBoundAndRestart(t *testing.T) {
	t.Parallel()
	s, h, src := newService(t)
	now := h.Now()
	src.gh["o/a"] = []rpc.ActivityItem{
		item("o/a", "1", "comment", "u", "255", now.Add(-time.Hour)),
		item("o/a", "2", "review", "bot[bot]", "260", now.Add(-2*time.Hour)),
		item("o/a", "old", "push", "u", "", now.Add(-48*time.Hour)),
	}
	src.gh["o/b"] = []rpc.ActivityItem{item("o/b", "3", "ci", "u", "", now.Add(-30*time.Minute))}

	assert.Equal(t, 3, s.Poll(t.Context(), src))
	assert.Equal(t, 0, s.Poll(t.Context(), src), "items are deduped by id")
	assert.Positive(t, h.changed, "subscribers are told")
	changed := h.changed
	s.Poll(t.Context(), src)
	assert.Equal(t, changed, h.changed, "no event for unchanged items")
	assert.Equal(t, []rpc.ActivityItem{src.gh["o/b"][0]}, s.List(rpc.ActivityListParams{Repo: "o/b"}, nil, now))
	assert.Len(t, s.List(rpc.ActivityListParams{}, nil, now), 3, "the 48h-old push is outside the window")

	re := New(h, slog.New(slog.DiscardHandler), filepath.Dir(s.path))
	assert.Len(t, re.List(rpc.ActivityListParams{}, nil, now), 3)
	assert.Equal(t, "etag-o/a", re.state().Repos["o/a"].EventsETag)

	for i := range 5 {
		src.gh["o/b"] = append(src.gh["o/b"], item("o/b", fmt.Sprint("n", i), "push", "x", "", now.Add(time.Duration(-i)*time.Minute)))
	}
	s.Poll(t.Context(), src)
	got := s.List(rpc.ActivityListParams{Repo: "o/b"}, nil, now)
	require.Len(t, got, 3)
	assert.Equal(t, "gh:o/b:n0", got[0].ID, "the bound keeps the newest")
}

func TestListFiltersAndLinks(t *testing.T) {
	t.Parallel()
	s, h, src := newService(t)
	now := h.Now()
	src.gh["o/a"] = []rpc.ActivityItem{
		item("o/a", "1", "comment", "human", "255", now.Add(-time.Hour)),
		item("o/a", "2", "review", "Bot[bot]", "260", now.Add(-2*time.Hour)),
	}
	s.Poll(t.Context(), src)
	links := map[string][]string{"o/a#255": {"run1"}}
	got := s.List(rpc.ActivityListParams{Ticket: "run1"}, links, now)
	require.Len(t, got, 1)
	assert.Equal(t, []string{"run1"}, got[0].Tickets)
	assert.Len(t, s.List(rpc.ActivityListParams{Bots: "humans"}, links, now), 1)
	assert.Len(t, s.List(rpc.ActivityListParams{Bots: "bots"}, links, now), 1)
	assert.Len(t, s.List(rpc.ActivityListParams{Kind: "review", Actor: "BOT[BOT]"}, links, now), 1)
	assert.Len(t, s.List(rpc.ActivityListParams{Since: 90 * time.Minute}, links, now), 1)
	assert.Len(t, s.List(rpc.ActivityListParams{Limit: 1}, links, now), 1)
	assert.False(t, ValidBots("robots"))
	assert.True(t, ValidBots("humans"))
}

func TestPollInterval(t *testing.T) {
	t.Parallel()
	s, h, src := newService(t)
	s.state().Repos["o/a"] = &RepoPoll{Next: h.Now().Add(time.Minute)}
	s.Poll(t.Context(), src)
	assert.Equal(t, 0, src.calls["o/a"], "X-Poll-Interval not yet passed")
	assert.Equal(t, 1, src.calls["o/b"])
	h.advance(2 * time.Minute)
	s.Poll(t.Context(), src)
	assert.Equal(t, 1, src.calls["o/a"])
}

func TestErrorLoggedOnce(t *testing.T) {
	t.Parallel()
	s, _, src := newService(t)
	src.err = fmt.Errorf("boom")
	s.Poll(t.Context(), src)
	s.Poll(t.Context(), src)
	assert.Equal(t, "boom", s.state().errs["github:o/a"])
	src.err = nil
	s.Poll(t.Context(), src)
	assert.Empty(t, s.state().errs)
}

// TestRecordAndAdmit: controller-made items and admitted items reach the
// cache once; an unconfigured feed records nothing.
func TestRecordAndAdmit(t *testing.T) {
	t.Parallel()
	s, h, _ := newService(t)
	it := rpc.ActivityItem{ID: "gh-call:1", Source: rpc.ActivityGitHub, Kind: "gh-merge", At: h.Now(), Title: "t"}
	s.Record(it)
	s.Record(it)
	assert.Equal(t, 1, h.changed, "a repeated item is not a change")
	assert.Equal(t, 1, s.Admit(t.Context(), []rpc.ActivityItem{it, {ID: "x", Source: "jobshop", At: h.Now()}}))
	assert.Len(t, h.admitted, 2, "the host sees everything admitted")
	assert.Len(t, s.List(rpc.ActivityListParams{}, nil, h.Now()), 2)

	h.cfg = &config.Config{}
	s.Record(rpc.ActivityItem{ID: "y", At: h.Now()})
	assert.Zero(t, s.Admit(t.Context(), []rpc.ActivityItem{{ID: "z", At: h.Now()}}))
	h.cfg = &config.Config{Activity: &config.Activity{}}
	assert.Len(t, s.List(rpc.ActivityListParams{}, nil, h.Now()), 2)
	_, err := os.Stat(s.path)
	require.NoError(t, err)
}

// TestLiveActivityETag runs the HTTP source against a fake GitHub: the
// second poll sends If-None-Match and a 304 yields nothing.
func TestLiveActivityETag(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var inm []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch {
		case strings.HasSuffix(r.URL.Path, "/events"):
			inm = append(inm, r.Header.Get("If-None-Match"))
			w.Header().Set("X-Poll-Interval", "60")
			if r.Header.Get("If-None-Match") == `W/"e1"` {
				w.WriteHeader(http.StatusNotModified)
				return
			}
			w.Header().Set("ETag", `W/"e1"`)
			_, _ = w.Write([]byte(`[{"id":"9","type":"IssueCommentEvent","actor":{"login":"elee1766"},"created_at":"2026-11-03T07:00:00Z",
			  "payload":{"action":"created","issue":{"number":255,"title":"Desk","html_url":"https://github.com/o/r/issues/255"},"comment":{"body":"hi","html_url":"https://github.com/o/r/issues/255#c"}}}]`))
		case strings.HasSuffix(r.URL.Path, "/actions/runs"):
			assert.Equal(t, "main", r.URL.Query().Get("branch"))
			_, _ = w.Write([]byte(`{"workflow_runs":[{"id":5,"run_attempt":1,"name":"ci","display_title":"x","conclusion":"failure","head_branch":"main","html_url":"u","updated_at":"2026-11-03T07:30:00Z","actor":{"login":"elee1766"}}]}`))
		default:
			_, _ = w.Write([]byte(`{"default_branch":"main"}`))
		}
	}))
	defer srv.Close()
	l := &Live{C: fakeClients{}}
	cl := github.New(srv.URL, apicall.Static("t"))
	st := &RepoPoll{}
	// Point the host-token path at the fake.
	items, err := l.fetchGitHub(t.Context(), cl, "o/r", st)
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "comment", items[0].Kind)
	assert.Equal(t, "255", items[0].Ref)
	assert.Equal(t, "hi", items[0].Body)
	assert.Equal(t, "ci", items[1].Kind)
	assert.False(t, st.Next.IsZero())
	items, err = l.fetchGitHub(t.Context(), cl, "o/r", st)
	require.NoError(t, err)
	assert.Equal(t, "ci", items[0].Kind, "runs have no ETag here, events 304")
	assert.Len(t, items, 1)
	assert.Equal(t, []string{"", `W/"e1"`}, inm)
}

func TestParseRepoEvents(t *testing.T) {
	t.Parallel()
	body := `[
	 {"id":"1","type":"PushEvent","actor":{"login":"elee-agent[bot]"},"created_at":"2026-11-03T07:00:00Z","payload":{"ref":"refs/heads/issue-1","head":"bbbbbbbbbbbbbbbb","before":"aaaaaaaaaaaaaaaa"}},
	 {"id":"2","type":"PullRequestEvent","actor":{"login":"elee-agent[bot]"},"created_at":"2026-11-03T07:00:00Z","payload":{"action":"merged","number":254,"pull_request":{"number":254}}}, 
	 {"id":"3","type":"PullRequestReviewEvent","actor":{"login":"elee1766"},"created_at":"2026-11-03T07:00:00Z","payload":{"action":"created","pull_request":{"number":7,"title":"T","html_url":"https://github.com/o/r/pull/7"},"review":{"state":"changes_requested","body":"no","html_url":"https://github.com/o/r/pull/7#r"}}},
	 {"id":"4","type":"DeleteEvent","actor":{"login":"x"},"created_at":"2026-11-03T07:00:00Z","payload":{"ref":"issue-1","ref_type":"branch"}},
	 {"id":"5","type":"IssueCommentEvent","actor":{"login":"x"},"created_at":"2026-11-03T07:00:00Z","payload":{"action":"edited","issue":{"number":1}}},
	 {"id":"6","type":"WatchEvent","actor":{"login":"x"},"created_at":"2026-11-03T07:00:00Z","payload":{}}
	]`
	items, err := parseRepoEvents("o/r", []byte(body))
	require.NoError(t, err)
	require.Len(t, items, 4)
	assert.Equal(t, "push", items[0].Kind)
	assert.True(t, items[0].Bot)
	assert.Equal(t, "https://github.com/o/r/compare/aaaaaaaaaaaa...bbbbbbbbbbbb", items[0].URL)
	assert.Equal(t, "merge", items[1].Kind)
	assert.Equal(t, "254", items[1].Ref)
	assert.True(t, items[1].PR)
	assert.Equal(t, "review changes requested on PR #7: T", items[2].Title)
	assert.Equal(t, "branch", items[3].Kind)
}
