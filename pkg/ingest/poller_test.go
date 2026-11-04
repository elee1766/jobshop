package ingest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite"
)

// fakeGitHub serves the issues, comments and reviews lists of one repo
// with ETags: a request whose If-None-Match matches gets a 304.
type fakeGitHub struct {
	mu       sync.Mutex
	issues   string
	comments string
	reviews  map[int]string
	hits     map[string]int // path -> requests
	notMod   int
}

func (f *fakeGitHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.hits[r.URL.Path]++
	var body string
	switch r.URL.Path {
	case "/repos/acme/app/issues":
		body = f.issues
	case "/repos/acme/app/issues/comments":
		body = f.comments
	default:
		var n int
		if _, err := fmt.Sscanf(r.URL.Path, "/repos/acme/app/pulls/%d/reviews", &n); err != nil {
			http.NotFound(w, r)
			return
		}
		body = f.reviews[n]
	}
	if body == "" {
		body = "[]"
	}
	etag := fmt.Sprintf(`"%x"`, len(body)*31+len(r.URL.Path))
	if r.Header.Get("If-None-Match") == etag {
		f.notMod++
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("ETag", etag)
	_, _ = w.Write([]byte(body))
}

func pollFixture(t *testing.T) (*Poller, *fakeGitHub, *sqlite.Store, *time.Time) {
	t.Helper()
	st, err := sqlite.Open(t.Context(), filepath.Join(t.TempDir(), "j.db"), sqlite.Options{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })
	fk := &fakeGitHub{hits: map[string]int{}, reviews: map[int]string{}}
	srv := httptest.NewServer(fk)
	t.Cleanup(srv.Close)
	now := time.Date(2026, 11, 4, 10, 0, 0, 0, time.UTC)
	p := &Poller{
		Name: "gh-poll", Events: config.PollerEvents, HTTP: srv.Client(), Store: st,
		Now:  func() time.Time { return now },
		Auth: func(context.Context, string) (string, string, error) { return srv.URL, "tok", nil },
	}
	return p, fk, st, &now
}

func inbox(t *testing.T, st *sqlite.Store) []store.InboxItem {
	t.Helper()
	its, err := st.InboxList(t.Context(), store.InboxFilter{})
	require.NoError(t, err)
	return its
}

// TestPollerRows: the first pass only sets the cursor; a new issue, PR,
// review and comment then become webhook-shaped rows; an unchanged repo
// is all 304s and no rows; a repeat poll of the same items adds none.
func TestPollerRows(t *testing.T) {
	p, fk, st, now := pollFixture(t)
	ctx := t.Context()

	fk.issues = `[{"number":3,"title":"old","created_at":"2026-11-01T00:00:00Z","updated_at":"2026-11-01T00:00:00Z","user":{"login":"x"}}]`
	res, err := p.Poll(ctx, "acme/app")
	require.NoError(t, err)
	assert.Zero(t, res.Items, "first pass starts at now")
	assert.Zero(t, fk.hits["/repos/acme/app/issues"], "no read before a cursor exists")
	cur, err := st.PollCursor(ctx, "gh-poll", "acme/app", "comments")
	require.NoError(t, err)
	require.True(t, cur.Found)
	assert.Equal(t, *now, cur.Since)

	fk.issues = `[
	 {"number":7,"title":"bug","state":"open","created_at":"2026-11-04T10:01:00Z","updated_at":"2026-11-04T10:01:00Z","user":{"login":"alice"}},
	 {"number":8,"title":"fix","state":"open","created_at":"2026-11-04T09:00:00Z","updated_at":"2026-11-04T10:02:00Z","user":{"login":"bob"},"pull_request":{}}]`
	fk.comments = `[{"id":99,"body":"hi","issue_url":"https://api/repos/acme/app/issues/7","user":{"login":"carol"},
	 "created_at":"2026-11-04T10:03:00Z","updated_at":"2026-11-04T10:03:00Z"}]`
	fk.reviews[8] = `[{"id":5,"state":"APPROVED","user":{"login":"dan"},"submitted_at":"2026-11-04T10:04:00Z"},
	 {"id":4,"state":"COMMENTED","user":{"login":"dan"},"submitted_at":"2026-11-03T10:04:00Z"}]`
	res, err = p.Poll(ctx, "acme/app")
	require.NoError(t, err)
	assert.Equal(t, PollResult{Items: 4, Fresh: 4}, res)

	byKind := map[string]store.InboxItem{}
	for _, it := range inbox(t, st) {
		byKind[it.Kind] = it
		assert.Equal(t, "gh-poll", it.Endpoint)
		assert.Equal(t, SourceGitHub, it.Source)
		assert.Equal(t, "acme/app", it.RepoOrTeam)
	}
	require.Len(t, byKind, 4)
	assert.Equal(t, int64(7), byKind["issues"].Number)
	assert.Equal(t, "alice", byKind["issues"].Actor)
	assert.Contains(t, string(byKind["issues"].Payload), `"action":"opened"`)
	assert.Equal(t, int64(8), byKind["pull_request"].Number)
	assert.Contains(t, string(byKind["pull_request"].Payload), `"action":"updated"`)
	cm := byKind["issue_comment"]
	assert.Equal(t, int64(7), cm.Number)
	assert.Equal(t, "99", cm.CommentID)
	assert.Equal(t, GitHubKey("acme/app", "comment", "99", time.Date(2026, 11, 4, 10, 3, 0, 0, time.UTC)), cm.NaturalKey)
	assert.Equal(t, "poll:"+cm.NaturalKey, cm.DeliveryID)
	rv := byKind["pull_request_review"]
	assert.Equal(t, int64(8), rv.Number)
	assert.Equal(t, "5", rv.CommentID, "the review older than the cursor is skipped")

	// Nothing changed: every list answers 304.
	before := fk.notMod
	res, err = p.Poll(ctx, "acme/app")
	require.NoError(t, err)
	assert.Zero(t, res.Fresh)
	assert.Equal(t, 2, fk.notMod-before, "issues and comments are 304s")
	assert.Len(t, inbox(t, st), 4)

	// Same items again with a new ETag (the boundary item GitHub returns
	// for since): no duplicates.
	fk.comments += " "
	res, err = p.Poll(ctx, "acme/app")
	require.NoError(t, err)
	assert.Zero(t, res.Fresh)
	assert.Len(t, inbox(t, st), 4)

	// A restart resumes from the stored cursor.
	p2 := *p
	res, err = p2.Poll(ctx, "acme/app")
	require.NoError(t, err)
	assert.Zero(t, res.Fresh)
}

// TestPollerWebhookDedupe: a comment that arrived by webhook and then by
// poll is one inbox row.
func TestPollerWebhookDedupe(t *testing.T) {
	p, fk, st, _ := pollFixture(t)
	ctx := t.Context()
	_, err := p.Poll(ctx, "acme/app")
	require.NoError(t, err)

	sp, err := OpenSpool(ctx, filepath.Join(t.TempDir(), "spool.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = sp.Close() })
	ok, err := sp.Write(ctx, Delivery{
		Endpoint: "gh", Source: SourceGitHub, DeliveryID: "guid-1", Kind: "issue_comment",
		Payload: []byte(`{"action":"created","repository":{"full_name":"acme/app"},"sender":{"login":"carol"},
		"issue":{"number":7},"comment":{"id":99,"updated_at":"2026-11-04T10:03:00Z"}}`),
	})
	require.NoError(t, err)
	require.True(t, ok)
	_, err = Drain(ctx, sp, st, time.Now, quiet)
	require.NoError(t, err)
	require.Len(t, inbox(t, st), 1)

	fk.comments = `[{"id":99,"body":"hi","issue_url":"https://api/repos/acme/app/issues/7","user":{"login":"carol"},
	 "created_at":"2026-11-04T10:03:00Z","updated_at":"2026-11-04T10:03:00Z"}]`
	res, err := p.Poll(ctx, "acme/app")
	require.NoError(t, err)
	assert.Equal(t, 1, res.Items)
	assert.Zero(t, res.Fresh)
	its := inbox(t, st)
	require.Len(t, its, 1)
	assert.Equal(t, "gh", its[0].Endpoint, "the webhook row is kept")
}

// TestPollerEvents: a poller limited to issue_comment reads no issues.
func TestPollerEvents(t *testing.T) {
	p, fk, _, _ := pollFixture(t)
	p.Events = []string{"issue_comment"}
	for range 2 {
		_, err := p.Poll(t.Context(), "acme/app")
		require.NoError(t, err)
	}
	assert.Zero(t, fk.hits["/repos/acme/app/issues"])
	assert.Equal(t, 1, fk.hits["/repos/acme/app/issues/comments"])
}
