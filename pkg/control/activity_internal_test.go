package control

import (
	"context"
	"encoding/json"
	"fmt"
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

	"github.com/elee1766/jobshop/pkg/apicall"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/github"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

const activityYAML = `projects:
  js: {}
agents:
  fake: {command: /bin/false}
pools:
  dev: {agent: fake, dir: %q}
  plan: {agent: fake, dir: %q, admin: [activity]}
workflows:
  pr-work:
    vars: [issue]
    issue: issue
    repo: elee1766/jobshop
    steps:
      - {id: implement, pool: dev, title: "Implement #{{.issue}}", prompt: "do it"}
activity:
  window: 24h
  max_items: 3
  repos: [elee1766/acphost]
`

type fakeActivity struct {
	mu    sync.Mutex
	gh    map[string][]rpc.ActivityItem
	calls map[string]int
	err   error
}

func (f *fakeActivity) GitHub(_ context.Context, repo string, st *repoPoll) ([]rpc.ActivityItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls[repo]++
	if f.err != nil {
		return nil, f.err
	}
	st.EventsETag = "etag-" + repo
	return f.gh[repo], nil
}

func (f *fakeActivity) Linear(context.Context, string, time.Time) ([]rpc.ActivityItem, error) {
	return nil, nil
}

func activityController(t *testing.T) (*Controller, *scanClock, *fakeActivity) {
	t.Helper()
	c, clk, _ := subController(t)
	require.NoError(t, os.WriteFile(filepath.Join(c.root.Dir, config.FileName), fmt.Appendf(nil, activityYAML, c.root.Dir, c.root.Dir), 0o600))
	cfg, err := config.Load(c.root)
	require.NoError(t, err)
	lc, err := newLiveConfig(cfg)
	require.NoError(t, err)
	c.live.Store(lc)
	c.activity = c.loadActivity()
	return c, clk, &fakeActivity{gh: map[string][]rpc.ActivityItem{}, calls: map[string]int{}}
}

func ghItem(repo, id, kind, actor, ref string, at time.Time) rpc.ActivityItem {
	return rpc.ActivityItem{ID: "gh:" + repo + ":" + id, Source: rpc.ActivityGitHub, Repo: repo, Kind: kind, Actor: actor, Bot: strings.HasSuffix(actor, "[bot]"), At: at, Title: kind, Ref: ref}
}

func listActivity(t *testing.T, c *Controller, p rpc.ActivityListParams) rpc.ActivityListResult {
	t.Helper()
	v, err := c.activityList(t.Context(), scope{admin: true}, p)
	require.NoError(t, err)
	return v.(rpc.ActivityListResult)
}

func TestActivityPollListAndLinks(t *testing.T) {
	t.Parallel()
	c, clk, src := activityController(t)
	now := clk.now()
	run := mkTicket(t, c, store.NewTicket{Kind: store.KindRun, Title: "run", Fields: json.RawMessage(`{"workflow":"pr-work","state":"running"}`), Meta: json.RawMessage(`{"github":{"repo":"elee1766/jobshop","issue":255}}`)})
	mkTicket(t, c, store.NewTicket{Kind: store.KindStep, Title: "step", Fields: json.RawMessage(`{"step":"implement","run":"` + run.ID + `"}`), ParentID: run.ID, Meta: json.RawMessage(`{"github":{"repo":"elee1766/jobshop","pr":"260"}}`)})
	src.gh["elee1766/jobshop"] = []rpc.ActivityItem{
		ghItem("elee1766/jobshop", "1", "comment", "elee1766", "255", now.Add(-time.Hour)),
		ghItem("elee1766/jobshop", "2", "review", "elee-agent[bot]", "260", now.Add(-2*time.Hour)),
		ghItem("elee1766/jobshop", "old", "push", "elee1766", "", now.Add(-48*time.Hour)),
	}
	src.gh["elee1766/acphost"] = []rpc.ActivityItem{ghItem("elee1766/acphost", "3", "ci", "elee1766", "", now.Add(-30*time.Minute))}

	assert.Equal(t, 3, c.pollActivity(t.Context(), c.activity, src))
	assert.Equal(t, 0, c.pollActivity(t.Context(), c.activity, src), "items are deduped by id")

	res := listActivity(t, c, rpc.ActivityListParams{})
	require.Len(t, res.Items, 3, "the 48h-old push is outside the 24h window")
	assert.Equal(t, []string{"github:elee1766/acphost", "github:elee1766/jobshop"}, res.Sources)
	assert.Equal(t, "ci", res.Items[0].Kind, "newest first")
	assert.Equal(t, []string{run.ID}, res.Items[1].Tickets, "issue links to its run")
	assert.Equal(t, []string{run.ID}, res.Items[2].Tickets, "a step's PR links to the step's run")

	assert.Len(t, listActivity(t, c, rpc.ActivityListParams{Bots: "humans"}).Items, 2)
	assert.Len(t, listActivity(t, c, rpc.ActivityListParams{Bots: "bots"}).Items, 1)
	assert.Len(t, listActivity(t, c, rpc.ActivityListParams{Repo: "elee1766/acphost", Since: time.Hour}).Items, 1)
	assert.Len(t, listActivity(t, c, rpc.ActivityListParams{Since: 45 * time.Minute}).Items, 1)
	assert.Len(t, listActivity(t, c, rpc.ActivityListParams{Ticket: run.ID}).Items, 2)
	assert.Len(t, listActivity(t, c, rpc.ActivityListParams{Kind: "review", Actor: "ELEE-AGENT[bot]"}).Items, 1)

	// The cache survives a restart: a fresh load has the items and ETags.
	re := c.loadActivity()
	assert.Len(t, re.Items, 3)
	assert.Equal(t, "etag-elee1766/jobshop", re.Repos["elee1766/jobshop"].EventsETag)

	// The per-source bound (max_items 3) keeps the newest.
	for i := range 5 {
		src.gh["elee1766/acphost"] = append(src.gh["elee1766/acphost"], ghItem("elee1766/acphost", fmt.Sprint("n", i), "push", "x", "", now.Add(time.Duration(-i)*time.Minute)))
	}
	c.pollActivity(t.Context(), c.activity, src)
	got := listActivity(t, c, rpc.ActivityListParams{Repo: "elee1766/acphost"}).Items
	require.Len(t, got, 3)
	assert.Equal(t, "gh:elee1766/acphost:n0", got[0].ID)
}

func TestActivityGrantAndOff(t *testing.T) {
	t.Parallel()
	c, _, _ := activityController(t)
	_, err := c.activityList(t.Context(), scope{pool: "dev"}, rpc.ActivityListParams{})
	require.ErrorContains(t, err, "activity")
	_, err = c.activityList(t.Context(), scope{pool: "plan"}, rpc.ActivityListParams{})
	require.NoError(t, err)
	_, err = c.activityList(t.Context(), scope{admin: true}, rpc.ActivityListParams{Bots: "robots"})
	require.ErrorIs(t, err, store.ErrInvalid)
}

func TestActivityPollInterval(t *testing.T) {
	t.Parallel()
	c, clk, src := activityController(t)
	c.activity.Repos["elee1766/jobshop"] = &repoPoll{Next: clk.now().Add(time.Minute)}
	c.pollActivity(t.Context(), c.activity, src)
	assert.Equal(t, 0, src.calls["elee1766/jobshop"], "X-Poll-Interval not yet passed")
	assert.Equal(t, 1, src.calls["elee1766/acphost"])
	clk.advance(2 * time.Minute)
	c.pollActivity(t.Context(), c.activity, src)
	assert.Equal(t, 1, src.calls["elee1766/jobshop"])
}

func TestActivityErrorLoggedOnce(t *testing.T) {
	t.Parallel()
	c, _, src := activityController(t)
	s := c.activity
	src.err = fmt.Errorf("boom")
	c.pollActivity(t.Context(), s, src)
	c.pollActivity(t.Context(), s, src)
	assert.Equal(t, "boom", s.errs["github:elee1766/jobshop"])
	src.err = nil
	c.pollActivity(t.Context(), s, src)
	assert.Empty(t, s.errs)
}

// TestLiveActivityETag runs the HTTP source against a fake GitHub: the
// second poll sends If-None-Match and a 304 yields nothing.
func TestLiveActivityETag(t *testing.T) {
	t.Parallel()
	c, _, _ := activityController(t)
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
	l := &liveActivity{c: c}
	cl := github.New(srv.URL, apicall.Static("t"))
	st := &repoPoll{}
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
