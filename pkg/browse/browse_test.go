package browse

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/integrations/apicall"
	"github.com/elee1766/jobshop/pkg/integrations/github"
	"github.com/elee1766/jobshop/pkg/integrations/linear"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// The fixtures in testdata are what Linear and GitHub answered for the four
// reads (see record_test.go). A fake server replays them, so the tests meet
// the providers' real shapes.

type request struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type fakeHost struct {
	cfg *config.Config
	lin *linear.Client
	gh  *github.Client

	mu  sync.Mutex
	now time.Time
}

func (h *fakeHost) Config() *config.Config { return h.cfg }
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
func (h *fakeHost) Linear(string) (*linear.Client, error) { return h.lin, nil }
func (h *fakeHost) GitHub(string) (*github.Client, error) { return h.gh, nil }

// rig is a service over a fake provider. fail, when set, answers instead of
// the fixture.
type rig struct {
	*Service
	host  *fakeHost
	calls atomic.Int64
	last  atomic.Pointer[request]
	// fail is read under mu.
	mu   sync.Mutex
	fail func(w http.ResponseWriter)
	gate chan struct{}
}

func (r *rig) setFail(f func(w http.ResponseWriter)) {
	r.mu.Lock()
	r.fail = f
	r.mu.Unlock()
}

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name + ".json")
	require.NoError(t, err)
	return b
}

func newRig(t *testing.T) *rig {
	t.Helper()
	r := &rig{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var body request
		b, _ := io.ReadAll(req.Body)
		_ = json.Unmarshal(b, &body)
		r.calls.Add(1)
		r.last.Store(&body)
		if r.gate != nil {
			<-r.gate
		}
		r.mu.Lock()
		fail := r.fail
		r.mu.Unlock()
		if fail != nil {
			fail(w)
			return
		}
		name := "github_list"
		switch q := body.Query; {
		case strings.Contains(q, "searchIssues") || strings.Contains(q, "issues("):
			name = "linear_list"
		case strings.Contains(q, "issue(id"):
			name = "linear_show"
		case strings.Contains(q, "pullRequest"):
			name = "github_show"
		}
		_, _ = w.Write(fixture(t, name))
	}))
	t.Cleanup(srv.Close)
	cfg := &config.Config{
		Identities: map[string]config.Identity{
			"lin": {Kind: config.IdentityLinear, Teams: []string{"JOB", "*"}, Workspace: "gfx-labs"},
		},
		Activity: &config.Activity{Repos: []string{"elee1766/jobshop"}},
	}
	r.host = &fakeHost{
		cfg: cfg,
		// The provider clients stamp a Retry-After with the real clock, and the
		// service subtracts its own: a fixed date here makes the wait grow with
		// the real date (it broke master once real time passed it). Start at the
		// real time, then only advance.
		now: time.Now(),
		lin: linear.New(srv.URL, apicall.Static("lin_api_secretvalue")),
		gh:  github.New(srv.URL, apicall.Static("ghs_secretvalue")),
	}
	r.Service = New(r.host, slog.New(slog.DiscardHandler))
	return r
}

var (
	linearJob = rpc.BrowseListParams{Provider: rpc.BrowseLinear, Scope: "JOB", Limit: 8}
	ghRepo    = rpc.BrowseListParams{Provider: rpc.BrowseGitHub, Scope: "elee1766/jobshop", Kind: "all"}
)

func TestScope(t *testing.T) {
	r := newRig(t)
	sc := r.Scope()
	assert.Equal(t, []rpc.BrowseLinearScope{{Team: "JOB", Workspace: "gfx-labs"}}, sc.Linear, `"*" names no team`)
	assert.Equal(t, []rpc.BrowseGitHubScope{{Repo: "elee1766/jobshop"}}, sc.GitHub)
	assert.False(t, sc.Writes.Enabled)
	assert.NotEmpty(t, sc.Writes.Reason)
	assert.Positive(t, sc.Limits.PageSize)
}

func TestScopeRejection(t *testing.T) {
	r := newRig(t)
	ctx := t.Context()
	for _, p := range []rpc.BrowseListParams{
		{Provider: rpc.BrowseLinear, Scope: "OTHER"},
		{Provider: rpc.BrowseLinear, Scope: "*"},
		{Provider: rpc.BrowseGitHub, Scope: "evil/repo"},
		{Provider: rpc.BrowseGitHub, Scope: "elee1766/jobshop", Author: "a b"},
		{Provider: rpc.BrowseGitHub, Scope: "elee1766/jobshop", View: "bogus"},
		{Provider: rpc.BrowseLinear, Scope: "JOB", View: "bogus"},
		{Provider: rpc.BrowseLinear, Scope: "JOB", Project: "not an id!"},
		{Provider: rpc.BrowseLinear, Scope: "JOB", Cursor: strings.Repeat("x", 5000)},
		{Provider: "gitlab", Scope: "x"},
	} {
		_, err := r.List(ctx, p)
		require.ErrorIs(t, err, domain.ErrInvalid, "%+v", p)
	}
	_, err := r.Show(ctx, rpc.BrowseShowParams{Provider: rpc.BrowseGitHub, Ref: "evil/repo#1"})
	require.ErrorIs(t, err, domain.ErrInvalid)
	_, err = r.Show(ctx, rpc.BrowseShowParams{Provider: rpc.BrowseLinear, Ref: "OTHER-1"})
	require.ErrorIs(t, err, domain.ErrInvalid)
	_, err = r.Catalog(ctx, rpc.BrowseCatalogParams{Team: "OTHER"})
	require.ErrorIs(t, err, domain.ErrInvalid)
	assert.Zero(t, r.calls.Load(), "a refused request never reaches a provider")
}

func TestLinearList(t *testing.T) {
	r := newRig(t)
	res, err := r.List(t.Context(), linearJob)
	require.NoError(t, err)
	assert.Equal(t, rpc.BrowseFresh, res.Status.State)
	require.NotEmpty(t, res.Items)
	it := res.Items[0]
	assert.Equal(t, "JOB-165", it.Ref)
	assert.Equal(t, "started", it.StateType)
	assert.Equal(t, "https://linear.app/gfx-labs/issue/JOB-165", it.URL)
	assert.True(t, res.HasMore, "a list is not complete just because a page is loaded")
	assert.NotEmpty(t, res.Next)
	f, _ := r.last.Load().Variables["filter"].(map[string]any)
	assert.Equal(t, map[string]any{"key": map[string]any{"eq": "JOB"}}, f["team"], "scope is always the team")
	assert.EqualValues(t, 8, r.last.Load().Variables["first"])

	_, err = r.List(t.Context(), rpc.BrowseListParams{Provider: rpc.BrowseLinear, Scope: "JOB", View: "closed", Cursor: res.Next})
	require.NoError(t, err)
	assert.Equal(t, res.Next, r.last.Load().Variables["after"])
}

func TestLinearExactKeyLookup(t *testing.T) {
	r := newRig(t)
	p := linearJob
	p.Query = "job-150"
	_, err := r.List(t.Context(), p)
	require.NoError(t, err)
	f := r.last.Load().Variables["filter"].(map[string]any)
	assert.Equal(t, map[string]any{"eq": float64(150)}, f["number"])
	assert.NotContains(t, f, "state", "a key is found at any age")
	assert.NotContains(t, r.last.Load().Query, "searchIssues")

	p.Query = "other-150"
	_, err = r.List(t.Context(), p)
	require.NoError(t, err)
	assert.Contains(t, r.last.Load().Query, "searchIssues", "another team's key is only text")
}

func TestLinearDetailThreads(t *testing.T) {
	r := newRig(t)
	d, err := r.Show(t.Context(), rpc.BrowseShowParams{Provider: rpc.BrowseLinear, Ref: "JOB-150"})
	require.NoError(t, err)
	assert.Equal(t, rpc.BrowseFresh, d.Status.State)
	assert.Equal(t, "JOB-150", d.Item.Ref)
	assert.Equal(t, "completed", d.Item.StateType)
	assert.NotEmpty(t, d.Body)

	var replies, comments int
	for _, th := range d.Threads {
		comments += len(th.Comments)
		for i, c := range th.Comments {
			if c.ReplyTo != "" {
				replies++
				assert.Positive(t, i, "a reply follows the message it answers")
				assert.NotEmpty(t, c.ReplyToAuthor)
			}
		}
		for i := 1; i < len(th.Comments); i++ {
			assert.False(t, th.Comments[i].At.Before(th.Comments[i-1].At), "oldest first inside a thread")
		}
	}
	assert.Equal(t, 10, comments, "every comment appears once")
	assert.Equal(t, 1, replies, "only the reply the provider names")

	var prs []string
	for _, l := range d.Linked {
		if l.Kind == "pr" {
			prs = append(prs, l.URL)
		}
	}
	assert.ElementsMatch(t, []string{"https://github.com/elee1766/jobshop/pull/1097", "https://github.com/elee1766/jobshop/pull/1092"}, prs)
}

func TestGitHubList(t *testing.T) {
	r := newRig(t)
	res, err := r.List(t.Context(), ghRepo)
	require.NoError(t, err)
	require.NotEmpty(t, res.Items)
	it := res.Items[0]
	assert.Equal(t, "elee1766/jobshop#1111", it.Ref)
	assert.Equal(t, "#1111", it.ID)
	assert.Equal(t, "pr", it.Kind)
	assert.Equal(t, "merged", it.State)
	assert.True(t, it.OwnerBot)
	assert.Equal(t, "pass", it.Checks)
	assert.Positive(t, res.Total)
	assert.True(t, res.HasMore)
	assert.NotEmpty(t, res.Next)
}

func TestGitHubSearchQuery(t *testing.T) {
	r := newRig(t)
	q := func(p rpc.BrowseListParams) string {
		t.Helper()
		p.Scope = "elee1766/jobshop"
		p.Provider = rpc.BrowseGitHub
		_, err := r.List(t.Context(), p)
		require.NoError(t, err)
		return r.last.Load().Variables["q"].(string)
	}
	assert.Equal(t, "repo:elee1766/jobshop is:pr is:open sort:updated-desc", q(rpc.BrowseListParams{}))
	assert.Equal(t, "repo:elee1766/jobshop is:issue is:closed is:unmerged sort:updated-desc", q(rpc.BrowseListParams{Kind: "issue", View: "closed"}))
	assert.Contains(t, q(rpc.BrowseListParams{Author: "elee-agent[bot]", Kind: "all", View: "all"}), "author:app/elee-agent")

	got := q(rpc.BrowseListParams{Query: "repo:evil/x org:evil -user:a is:private \"quoted\" fix\nbug", View: "all"})
	assert.True(t, strings.HasPrefix(got, "repo:elee1766/jobshop is:pr "), got)
	for _, w := range strings.Fields(got) {
		if strings.HasPrefix(w, "repo:") {
			assert.Equal(t, "repo:elee1766/jobshop", w, "the typed repo: is not a qualifier")
		}
		for _, bad := range []string{"org:", "user:", "is:private"} {
			assert.NotEqual(t, bad, w[:min(len(w), len(bad))], got)
		}
	}
	assert.Contains(t, got, "evil/x", "the words stay as text")
	assert.Contains(t, got, "in:title,body")
}

func TestGitHubDetail(t *testing.T) {
	r := newRig(t)
	d, err := r.Show(t.Context(), rpc.BrowseShowParams{Provider: rpc.BrowseGitHub, Ref: "elee1766/jobshop#1097", Kind: "pr"})
	require.NoError(t, err)
	assert.Equal(t, "pr", d.Item.Kind)
	assert.Equal(t, "merged", d.Item.State)
	assert.Equal(t, "master", d.Props.Base)
	assert.Equal(t, 75, d.Props.ChangedFiles)
	assert.NotEmpty(t, d.Files)
	var names []string
	for _, c := range d.Checks {
		names = append(names, c.Name)
		assert.Contains(t, []string{"pass", "fail", "run", "skip"}, c.State)
	}
	assert.Contains(t, names, "go")
	assert.Contains(t, names, "ui")
	require.Len(t, d.Linked, 1, "only the Linear key the PR names, in a configured team")
	assert.Equal(t, rpc.BrowseLinked{Kind: "issue", Ref: "JOB-150", Relation: "mentions"}, d.Linked[0])
	require.Len(t, d.Threads, 1)
	assert.True(t, d.Threads[0].Comments[0].Bot)
}

func TestCacheSingleFlightAndRefresh(t *testing.T) {
	r := newRig(t)
	ctx := t.Context()
	res, err := r.List(ctx, linearJob)
	require.NoError(t, err)
	assert.Equal(t, rpc.BrowseFresh, res.Status.State)
	res, err = r.List(ctx, linearJob)
	require.NoError(t, err)
	assert.Equal(t, rpc.BrowseCached, res.Status.State)
	assert.EqualValues(t, 1, r.calls.Load())

	p := linearJob
	p.Refresh = true
	res, err = r.List(ctx, p)
	require.NoError(t, err)
	assert.Equal(t, rpc.BrowseCached, res.Status.State, "a refresh right after a read does not hit the provider")
	r.host.advance(minRefresh + time.Second)
	res, err = r.List(ctx, p)
	require.NoError(t, err)
	assert.Equal(t, rpc.BrowseFresh, res.Status.State)
	assert.EqualValues(t, 2, r.calls.Load())

	r.host.advance(listTTL + time.Second)
	_, err = r.List(ctx, linearJob)
	require.NoError(t, err)
	assert.EqualValues(t, 3, r.calls.Load(), "an expired entry is read again")

	r.Invalidate(rpc.BrowseLinear, "JOB")
	before := r.Revision()
	res, err = r.List(ctx, linearJob)
	require.NoError(t, err)
	assert.Equal(t, rpc.BrowseFresh, res.Status.State, "an invalidated scope reads again")
	r.Invalidate(rpc.BrowseLinear, "JOB")
	assert.Greater(t, r.Revision(), before)
	assert.Equal(t, r.Revision(), r.Scope().Revision)

	// Another scope is untouched by it.
	_, err = r.List(ctx, ghRepo)
	require.NoError(t, err)
	n := r.calls.Load()
	r.Invalidate(rpc.BrowseLinear, "JOB")
	_, err = r.List(ctx, ghRepo)
	require.NoError(t, err)
	assert.Equal(t, n, r.calls.Load())

	// Many callers at once make one provider call.
	r2 := newRig(t)
	r2.gate = make(chan struct{})
	var wg sync.WaitGroup
	results := make([]rpc.BrowseListResult, 6)
	for i := range results {
		wg.Go(func() {
			results[i], _ = r2.List(ctx, linearJob)
		})
	}
	require.Eventually(t, func() bool { return r2.calls.Load() == 1 }, 5*time.Second, time.Millisecond)
	close(r2.gate)
	wg.Wait()
	assert.EqualValues(t, 1, r2.calls.Load())
	for _, x := range results {
		assert.NotEmpty(t, x.Items)
	}
}

func rateLimited(w http.ResponseWriter) {
	w.Header().Set("Retry-After", "120")
	w.WriteHeader(http.StatusTooManyRequests)
	_, _ = io.WriteString(w, `{"message":"rate limit lin_api_secretvalue"}`)
}

func TestRateLimitBlocksAndStaleServes(t *testing.T) {
	r := newRig(t)
	ctx := t.Context()
	_, err := r.List(ctx, linearJob)
	require.NoError(t, err)
	r.host.advance(listTTL + time.Second)

	r.setFail(rateLimited)
	res, err := r.List(ctx, linearJob)
	require.NoError(t, err, "a provider failure is a status, not an error")
	assert.Equal(t, rpc.BrowseStale, res.Status.State)
	assert.Equal(t, rpc.BrowseCodeRateLimited, res.Status.Code)
	assert.NotEmpty(t, res.Items, "the older copy is shown")
	assert.False(t, res.Status.RetryAt.IsZero())
	assert.NotContains(t, res.Status.Message, "secretvalue")
	n := r.calls.Load()

	// Any read of the team waits for the limit, not just this list.
	p := linearJob
	p.View = "backlog"
	res, err = r.List(ctx, p)
	require.NoError(t, err)
	assert.Equal(t, rpc.BrowseError, res.Status.State)
	assert.Equal(t, rpc.BrowseCodeRateLimited, res.Status.Code)
	assert.Empty(t, res.Items)
	assert.NotNil(t, res.Items, "an empty list is still a list")
	_, err = r.Show(ctx, rpc.BrowseShowParams{Provider: rpc.BrowseLinear, Ref: "JOB-150"})
	require.NoError(t, err)
	assert.Equal(t, n, r.calls.Load(), "no call while blocked")

	// GitHub is a different bucket.
	r.setFail(nil)
	res, err = r.List(ctx, ghRepo)
	require.NoError(t, err)
	assert.Equal(t, rpc.BrowseFresh, res.Status.State)

	r.host.advance(3 * time.Minute)
	res, err = r.List(ctx, linearJob)
	require.NoError(t, err)
	assert.Equal(t, rpc.BrowseFresh, res.Status.State, "reads resume after the wait")

	// An old copy is not shown for ever.
	r.host.advance(listTTL + staleMax)
	r.setFail(func(w http.ResponseWriter) { w.WriteHeader(http.StatusBadGateway) })
	res, err = r.List(ctx, linearJob)
	require.NoError(t, err)
	assert.Equal(t, rpc.BrowseError, res.Status.State)
	assert.Equal(t, rpc.BrowseCodeUnavailable, res.Status.Code)
	assert.Empty(t, res.Items)
}

// leakyToken has the shape of a real Linear key, which the redactor knows.
const leakyToken = "lin_api_aB3dE6gH9jK2mN5pQ8sT1vW4yZ7cF0hJ3kL6nP9r"

func TestProviderErrorCodes(t *testing.T) {
	ctx := t.Context()
	graphQLError := func(code string) func(http.ResponseWriter) {
		return func(w http.ResponseWriter) {
			_, _ = io.WriteString(w, `{"errors":[{"message":"nope `+leakyToken+`","extensions":{"code":"`+code+`"}}]}`)
		}
	}
	for name, tc := range map[string]struct {
		fail func(http.ResponseWriter)
		code string
	}{
		"http 401":       {func(w http.ResponseWriter) { w.WriteHeader(401) }, rpc.BrowseCodeAuth},
		"http 429":       {rateLimited, rpc.BrowseCodeRateLimited},
		"http 500":       {func(w http.ResponseWriter) { w.WriteHeader(500) }, rpc.BrowseCodeUnavailable},
		"ratelimited":    {graphQLError("RATELIMITED"), rpc.BrowseCodeRateLimited},
		"authentication": {graphQLError("AUTHENTICATION_ERROR"), rpc.BrowseCodeAuth},
		"other":          {graphQLError("INTERNAL"), rpc.BrowseCodeError},
	} {
		t.Run(name, func(t *testing.T) {
			r := newRig(t)
			r.setFail(tc.fail)
			res, err := r.List(ctx, linearJob)
			require.NoError(t, err)
			assert.Equal(t, rpc.BrowseError, res.Status.State)
			assert.Equal(t, tc.code, res.Status.Code)
			assert.NotContains(t, res.Status.Message, leakyToken, "no credential in a message")
			assert.NotEmpty(t, res.Status.Message)
		})
	}

	t.Run("github rate limit header", func(t *testing.T) {
		r := newRig(t)
		reset := r.host.now.Add(90 * time.Second).Unix()
		r.setFail(func(w http.ResponseWriter) {
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(reset, 10))
			w.WriteHeader(http.StatusForbidden)
		})
		res, err := r.List(ctx, ghRepo)
		require.NoError(t, err)
		assert.Equal(t, rpc.BrowseError, res.Status.State)
		assert.Equal(t, rpc.BrowseCodeRateLimited, res.Status.Code)
		assert.WithinDuration(t, r.host.now.Add(90*time.Second), res.Status.RetryAt, time.Second)
	})

	t.Run("not found", func(t *testing.T) {
		r := newRig(t)
		r.setFail(func(w http.ResponseWriter) {
			_, _ = io.WriteString(w, `{"data":{"repository":null},"errors":[{"type":"NOT_FOUND","message":"Could not resolve to a PullRequest with the number of 9."}]}`)
		})
		d, err := r.Show(ctx, rpc.BrowseShowParams{Provider: rpc.BrowseGitHub, Ref: "elee1766/jobshop#9"})
		require.NoError(t, err)
		assert.Equal(t, rpc.BrowseCodeNotFound, d.Status.Code)
		assert.Equal(t, "elee1766/jobshop#9", d.Item.Ref, "the reference is kept to show the error against")
		assert.NotNil(t, d.Threads)
	})
}

func TestBriefSkipsDiscussion(t *testing.T) {
	r := newRig(t)
	d, err := r.Show(t.Context(), rpc.BrowseShowParams{Provider: rpc.BrowseGitHub, Ref: "elee1766/jobshop#1097", Brief: true})
	require.NoError(t, err)
	assert.EqualValues(t, 1, r.last.Load().Variables["last"], "a brief read asks for the least discussion")
	assert.Equal(t, "elee1766/jobshop#1097", d.Item.Ref)
}

func TestSearchTermAndClip(t *testing.T) {
	assert.Equal(t, "a b c", searchTerm("a\x00\tb\n c "))
	assert.Len(t, searchTerm(strings.Repeat("é", 400)), 200)
	assert.Equal(t, "plain words here", plainWords("plain words here"))
	assert.Equal(t, "repo evil/x org evil user a is private quoted", plainWords("repo:evil/x org:evil -user:a is:private \"quoted\""))
	s, cut := clip(strings.Repeat("é", 100), 11)
	assert.True(t, cut)
	assert.LessOrEqual(t, len(s), 20)
	assert.Equal(t, 50, bound(0))
	assert.Equal(t, 100, bound(5000))
}

func TestCatalogHidesEmail(t *testing.T) {
	r := newRig(t)
	r.setFail(func(w http.ResponseWriter) {
		_, _ = io.WriteString(w, `{"data":{"team":{"states":{"nodes":[{"id":"s1","name":"Todo","type":"unstarted"}]},"projects":{"nodes":[]},"members":{"nodes":[{"id":"u1","name":"Ann","displayName":"ann","email":"ann@example.com"}]},"labels":{"nodes":[]}}}}`)
	})
	res, err := r.Catalog(t.Context(), rpc.BrowseCatalogParams{Team: "JOB"})
	require.NoError(t, err)
	b, _ := json.Marshal(res)
	assert.NotContains(t, string(b), "example.com")
	assert.NotNil(t, res.Labels)
}

func TestCacheBound(t *testing.T) {
	r := newRig(t)
	for i := range maxEntries + 20 {
		p := linearJob
		p.Query = "q" + strings.Repeat("x", i%7) + string(rune('a'+i%26)) + string(rune('a'+i/26%26))
		_, err := r.List(t.Context(), p)
		require.NoError(t, err)
	}
	r.mu2().Lock()
	defer r.mu2().Unlock()
	assert.LessOrEqual(t, len(r.cache), maxEntries)
}

func (r *rig) mu2() *sync.Mutex { return &r.Service.mu }
