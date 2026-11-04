package ingest

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite"
)

func drainFixture(t *testing.T) (*Spool, *sqlite.Store) {
	t.Helper()
	dir := t.TempDir()
	sp, err := OpenSpool(t.Context(), filepath.Join(dir, "spool.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = sp.Close() })
	st, err := sqlite.Open(t.Context(), filepath.Join(dir, "j.db"), sqlite.Options{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })
	return sp, st
}

const ghComment = `{"repository":{"full_name":"Acme/App"},"sender":{"login":"alice"},"issue":{"number":7},
"comment":{"id":99,"updated_at":"2026-11-04T10:00:00Z"}}`

func spoolGH(t *testing.T, sp *Spool, id, payload string) {
	t.Helper()
	ok, err := sp.Write(t.Context(), Delivery{Endpoint: "gh", Source: SourceGitHub, DeliveryID: id, Kind: "issue_comment", Payload: []byte(payload)})
	require.NoError(t, err)
	require.True(t, ok)
}

var quiet = slog.New(slog.DiscardHandler)

// TestDrainInsertsThenDeletes: rows land in the inbox with their subject
// and natural key, then leave the spool; an empty spool drains to zero.
func TestDrainInsertsThenDeletes(t *testing.T) {
	sp, st := drainFixture(t)
	ctx := t.Context()
	spoolGH(t, sp, "d1", ghComment)
	spoolGH(t, sp, "d2", `{"repository":{"full_name":"acme/app"}}`)

	res, err := Drain(ctx, sp, st, time.Now, quiet)
	require.NoError(t, err)
	assert.Equal(t, DrainResult{Moved: 2, Fresh: 2}, res)
	rows, err := sp.Rows(ctx, 0)
	require.NoError(t, err)
	assert.Empty(t, rows)

	its, err := st.InboxList(ctx, store.InboxFilter{})
	require.NoError(t, err)
	require.Len(t, its, 2)
	var c store.InboxItem
	for _, it := range its {
		if it.DeliveryID == "github:d1" {
			c = it
		}
	}
	assert.Equal(t, "Acme/App", c.RepoOrTeam)
	assert.Equal(t, int64(7), c.Number)
	assert.Equal(t, "99", c.CommentID)
	assert.Equal(t, "alice", c.Actor)
	assert.Equal(t, "gh", c.Endpoint)
	assert.Equal(t, "github:comment:acme/app:99:2026-11-04T10:00:00Z", c.NaturalKey)

	res, err = Drain(ctx, sp, st, time.Now, quiet)
	require.NoError(t, err)
	assert.Equal(t, DrainResult{}, res)
}

// crashAfterInsert commits the inbox batch and then fails, as a crash
// between the inbox commit and the spool delete would.
type crashAfterInsert struct {
	Inbox
	err error
}

func (c crashAfterInsert) InboxPutBatch(ctx context.Context, its []store.InboxItem) (int, error) {
	n, err := c.Inbox.InboxPutBatch(ctx, its)
	if err != nil {
		return n, err
	}
	return n, c.err
}

// TestDrainCrashBeforeDelete: the inbox commit happened but the spool
// rows stayed; the next drain moves them again without a second row.
func TestDrainCrashBeforeDelete(t *testing.T) {
	sp, st := drainFixture(t)
	ctx := t.Context()
	spoolGH(t, sp, "d1", ghComment)
	_, err := Drain(ctx, sp, crashAfterInsert{st, errors.New("crash")}, time.Now, quiet)
	require.Error(t, err)
	rows, err := sp.Rows(ctx, 0)
	require.NoError(t, err)
	require.Len(t, rows, 1, "spool row kept")

	res, err := Drain(ctx, sp, st, time.Now, quiet)
	require.NoError(t, err)
	assert.Equal(t, DrainResult{Moved: 1, Fresh: 0}, res)
	its, err := st.InboxList(ctx, store.InboxFilter{})
	require.NoError(t, err)
	assert.Len(t, its, 1)
}

// TestDrainBadRow: a row that does not parse moves to spool_bad and the
// rest of the batch still drains.
func TestDrainBadRow(t *testing.T) {
	sp, st := drainFixture(t)
	ctx := t.Context()
	spoolGH(t, sp, "good", ghComment)
	spoolGH(t, sp, "bad", `{"repository":"not an object"}`)
	res, err := Drain(ctx, sp, st, time.Now, quiet)
	require.NoError(t, err)
	assert.Equal(t, DrainResult{Moved: 1, Fresh: 1, Bad: 1}, res)
	n, err := sp.BadCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	rows, err := sp.Rows(ctx, 0)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

// TestWebhookAndPollDedupe: the webhook copy and a poller find of the
// same comment version share a natural key and give one inbox row.
func TestWebhookAndPollDedupe(t *testing.T) {
	sp, st := drainFixture(t)
	ctx := t.Context()
	spoolGH(t, sp, "d1", ghComment)
	_, err := Drain(ctx, sp, st, time.Now, quiet)
	require.NoError(t, err)

	at := time.Date(2026, 11, 4, 10, 0, 0, 0, time.UTC)
	key := GitHubKey("acme/app", "comment", "99", at)
	_, dup, err := st.InboxPut(ctx, store.InboxItem{Source: "github", DeliveryID: "poll:" + key, NaturalKey: key})
	require.NoError(t, err)
	assert.True(t, dup)
	its, err := st.InboxList(ctx, store.InboxFilter{})
	require.NoError(t, err)
	assert.Len(t, its, 1)

	// An edit is a new version, so a new row.
	key2 := GitHubKey("acme/app", "comment", "99", at.Add(time.Minute))
	_, dup, err = st.InboxPut(ctx, store.InboxItem{Source: "github", DeliveryID: "poll:" + key2, NaturalKey: key2})
	require.NoError(t, err)
	assert.False(t, dup)
}

// fakeDeliveries serves GET /app/hook/deliveries in two pages and
// records redeliver requests.
func fakeDeliveries(t *testing.T, pages [][]string) (*httptest.Server, *[]string) {
	t.Helper()
	var mu sync.Mutex
	var redelivered []string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer jwt" {
			http.Error(w, "no", http.StatusUnauthorized)
			return
		}
		if r.Method == http.MethodPost {
			mu.Lock()
			redelivered = append(redelivered, strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/app/hook/deliveries/"), "/attempts"))
			mu.Unlock()
			w.WriteHeader(http.StatusAccepted)
			return
		}
		page := 0
		if r.URL.Query().Get("cursor") == "2" {
			page = 1
		}
		if page == 0 && len(pages) > 1 {
			w.Header().Set("Link", `<`+srv.URL+`/app/hook/deliveries?per_page=100&cursor=2>; rel="next"`)
		}
		_, _ = w.Write([]byte("[" + strings.Join(pages[page], ",") + "]"))
	}))
	t.Cleanup(srv.Close)
	return srv, &redelivered
}

func hd(id int, guid string, at string, status int) string {
	return `{"id":` + strconv.Itoa(id) + `,"guid":"` + guid + `","delivered_at":"` + at + `","status_code":` + strconv.Itoa(status) + `}`
}

// TestReconcile: only failed GUIDs with no later success are
// redelivered, oldest first, bounded per pass; deliveries before the
// cutoff are not read.
func TestReconcile(t *testing.T) {
	srv, got := fakeDeliveries(t, [][]string{
		{ // newest first
			hd(6, "e", "2026-11-04T10:05:00Z", 500),
			hd(5, "b", "2026-11-04T10:04:00Z", 202), // b redelivered OK
			hd(4, "d", "2026-11-04T10:03:00Z", 502),
		},
		{
			hd(3, "c", "2026-11-04T10:02:00Z", 0), // timeout
			hd(2, "b", "2026-11-04T10:01:00Z", 500),
			hd(1, "a", "2026-11-04T09:00:00Z", 500), // before the cutoff
		},
	})
	r := &Reconcile{APIURL: srv.URL, JWT: func() (string, error) { return "jwt", nil }, Max: 2}
	since := time.Date(2026, 11, 4, 9, 30, 0, 0, time.UTC)
	res, err := r.Pass(t.Context(), since)
	require.NoError(t, err)
	assert.Equal(t, ReconcileResult{Failed: 3, Redelivered: 2}, res)
	assert.Equal(t, []string{"3", "4"}, *got, "c then d; e waits for the next pass")
}
