package activity

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/integrations/apicall"
	"github.com/elee1766/jobshop/pkg/integrations/github"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// TestLiveActivityApp (JOB-108): with an installation token the events API
// answers [], so the feed comes from the per-resource endpoints, each with
// its own ETag, and a refused endpoint does not hide the others.
func TestLiveActivityApp(t *testing.T) {
	t.Parallel()
	var mu sync.Mutex
	inm := map[string][]string{}
	var events int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		p := strings.TrimPrefix(r.URL.Path, "/repos/o/r")
		if p == "/events" {
			events++
			_, _ = w.Write([]byte(`[]`))
			return
		}
		if p == "/pulls/comments" {
			http.Error(w, `{"message":"Resource not accessible by integration"}`, http.StatusForbidden)
			return
		}
		inm[p] = append(inm[p], r.Header.Get("If-None-Match"))
		if r.Header.Get("If-None-Match") != "" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", `W/"`+p+`"`)
		switch p {
		case "/pulls":
			_, _ = w.Write([]byte(`[
			 {"number":7,"title":"T","html_url":"https://github.com/o/r/pull/7","user":{"login":"elee-agent[bot]"},"created_at":"2026-11-06T10:00:00Z","closed_at":"2026-11-06T11:00:00Z","merged_at":"2026-11-06T11:00:00Z"},
			 {"number":8,"title":"U","html_url":"https://github.com/o/r/pull/8","user":{"login":"alice"},"created_at":"2026-11-06T12:00:00Z","closed_at":null,"merged_at":null}]`))
		case "/pulls/7/reviews", "/pulls/8/reviews":
			_, _ = w.Write([]byte(`[{"id":31,"user":{"login":"bob"},"body":"no","state":"CHANGES_REQUESTED","html_url":"https://github.com/o/r/pull/7#r","submitted_at":"2026-11-06T10:30:00Z"},
			 {"id":32,"user":{"login":"bob"},"body":"","state":"PENDING","html_url":"x","submitted_at":null}]`))
		case "/issues":
			_, _ = w.Write([]byte(`[{"number":5,"title":"I","html_url":"https://github.com/o/r/issues/5","user":{"login":"alice"},"created_at":"2026-11-06T09:00:00Z","closed_at":null},
			 {"number":7,"title":"T","html_url":"https://github.com/o/r/pull/7","user":{"login":"elee-agent[bot]"},"created_at":"2026-11-06T10:00:00Z","closed_at":null,"pull_request":{}}]`))
		case "/issues/comments":
			_, _ = w.Write([]byte(`[{"id":41,"html_url":"https://github.com/o/r/pull/7#issuecomment-41","body":"hi","user":{"login":"alice"},"created_at":"2026-11-06T10:15:00Z","issue_url":"https://api.github.com/repos/o/r/issues/7"}]`))
		case "/commits":
			_, _ = w.Write([]byte(`[{"sha":"abcdef0123456","html_url":"https://github.com/o/r/commit/abcdef0123456","author":{"login":"elee-agent[bot]"},"commit":{"message":"fix: x\n\nbody","author":{"name":"n"},"committer":{"date":"2026-11-06T11:00:00Z"}}}]`))
		case "/actions/runs":
			_, _ = w.Write([]byte(`{"workflow_runs":[]}`))
		default:
			_, _ = w.Write([]byte(`{"default_branch":"main"}`))
		}
	}))
	defer srv.Close()
	l := &Live{C: fakeClients{}}
	cl := github.New(srv.URL, apicall.Static("ghs_x"))
	st := &RepoPoll{}

	items, err := l.fetchGitHub(t.Context(), cl, "o/r", st)
	require.ErrorContains(t, err, "review_comments", "the refused endpoint is reported")
	assert.Zero(t, events, "events are not read for an installation token")
	byID := map[string]rpc.ActivityItem{}
	for _, it := range items {
		byID[it.ID] = it
	}
	for _, id := range []string{
		"gh:o/r:pr:7:opened", "gh:o/r:pr:7:merged", "gh:o/r:pr:8:opened", "gh:o/r:review:31",
		"gh:o/r:issue:5:opened", "gh:o/r:comment:41", "gh:o/r:commit:abcdef0123456",
	} {
		assert.Contains(t, byID, id)
	}
	assert.NotContains(t, byID, "gh:o/r:issue:7:opened", "a PR listed as an issue is skipped")
	assert.NotContains(t, byID, "gh:o/r:review:32", "pending reviews are skipped")
	assert.Equal(t, "merge", byID["gh:o/r:pr:7:merged"].Kind)
	assert.True(t, byID["gh:o/r:pr:7:merged"].Bot)
	assert.Equal(t, "7", byID["gh:o/r:comment:41"].Ref)
	assert.True(t, byID["gh:o/r:comment:41"].PR)
	assert.Equal(t, "commit abcdef0: fix: x", byID["gh:o/r:commit:abcdef0123456"].Title)

	// Second poll: every endpoint with an ETag answers 304 and adds nothing.
	items, err = l.fetchGitHub(t.Context(), cl, "o/r", st)
	require.ErrorContains(t, err, "review_comments")
	assert.Empty(t, items)
	assert.Equal(t, []string{"", `W/"/pulls"`}, inm["/pulls"])
	assert.Equal(t, []string{"", `W/"/commits"`}, inm["/commits"])
}
