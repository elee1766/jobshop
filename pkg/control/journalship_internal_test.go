package control

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/journal"
	"github.com/elee1766/jobshop/pkg/journal/vltest"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// shipYAML is journalYAML plus a second project, its pool and a
// VictoriaLogs url (%[2]s).
var shipYAML = strings.Replace(strings.Replace(journalYAML, "projects:\n  js: {}\n", "projects:\n  js: {}\n  other: {}\ndefault_project: js\njournal:\n  victorialogs: {url: %[2]q}\n  lag_max: 1m\n", 1),
	"pools:\n", "pools:\n  searcher: {agent: fake, dir: %[1]q, admin: [journal-search]}\n  far: {project: other, agent: fake, dir: %[1]q}\n", 1)

// shipController is a controller without serving, with metrics, whose
// config points at url.
func shipController(t *testing.T, url string) (*Controller, *stallClock) {
	t.Helper()
	y := strings.ReplaceAll(shipYAML, "%[2]q", fmt.Sprintf("%q", url))
	c, clk, _ := stallControllerYAML(t, y)
	m, err := telemetry.NewMetrics()
	require.NoError(t, err)
	c.metrics = m
	return c, clk
}

// appendJournal adds frames from seq from to to, each at clock time.
// Frame from is a prompt for ticket tk; the rest are agent chunks, with a
// tool call every fourth.
func appendJournal(t *testing.T, c *Controller, id string, from, to int, at time.Time, tk string) {
	t.Helper()
	dir := c.root.SessionDir(id)
	require.NoError(t, os.MkdirAll(dir, 0o700))
	f, err := os.OpenFile(filepath.Join(dir, "journal.ndjson"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	require.NoError(t, err)
	defer func() { require.NoError(t, f.Close()) }()
	for i := from; i <= to; i++ {
		src, msg := "agent", fmt.Sprintf(`{"jsonrpc":"2.0","method":"session/update","params":{"update":{"sessionUpdate":"agent_message_chunk","content":{"text":"chunk %d permission denied"}}}}`, i)
		switch {
		case i == from:
			src, msg = "controller", fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"method":"session/prompt","params":{"_meta":{"jobshop/ticket":%q},"prompt":[]}}`, i, tk)
		case i%4 == 0:
			msg = fmt.Sprintf(`{"jsonrpc":"2.0","method":"session/update","params":{"update":{"sessionUpdate":"tool_call","toolCallId":"c%d","title":"go test","status":"failed","api_key":"sk-supersecretvalue123"}}}`, i)
		}
		_, err := fmt.Fprintf(f, `{"seq":%d,"t":%d,"from":%q,"msg":%s}`+"\n", i, at.UnixMilli(), src, msg)
		require.NoError(t, err)
	}
}

func search(t *testing.T, c *Controller, sc scope, q string) []map[string]any {
	t.Helper()
	out, err := c.journalSearch(t.Context(), sc, rpc.JournalSearchParams{Query: q})
	require.NoError(t, err)
	var rows []map[string]any
	for _, r := range out.(rpc.JournalSearchResult).Rows {
		var m map[string]any
		require.NoError(t, json.Unmarshal(r, &m))
		rows = append(rows, m)
	}
	return rows
}

func countRows(t *testing.T, c *Controller, q string) int {
	t.Helper()
	rows := search(t, c, scope{admin: true}, q+" | stats count() n")
	require.Len(t, rows, 1)
	var n int
	_, err := fmt.Sscan(rows[0]["n"].(string), &n)
	require.NoError(t, err)
	return n
}

func pass(t *testing.T, c *Controller, cl *journal.Client) {
	t.Helper()
	c.ship.mu.Lock()
	_, err := c.shipPass(t.Context(), cl, true)
	c.ship.mu.Unlock()
	require.NoError(t, err)
	require.NoError(t, cl.Flush(t.Context()))
}

func shippedSeq(t *testing.T, c *Controller, id string) int64 {
	t.Helper()
	tk, err := c.store.Get(t.Context(), id)
	require.NoError(t, err)
	f, err := store.DecodeFields[store.SessionFields](tk)
	require.NoError(t, err)
	return f.ShippedSeq
}

// TestJournalShipSearch ships frames to a real VictoriaLogs and finds them:
// row fields, redaction, the project filter for a pool caller, a resend
// after a restart that lost the cursor, and catching up after an outage.
func TestJournalShipSearch(t *testing.T) {
	t.Parallel()
	vl := vltest.Start(t)
	c, clk := shipController(t, vl.URL)
	cl := c.vlClient(c.config())
	ctx := t.Context()
	sess := mkTicket(t, c, store.NewTicket{Kind: store.KindSession, Pool: "dev", Fields: json.RawMessage(`{"agent":"fake"}`)})
	far := mkTicket(t, c, store.NewTicket{Kind: store.KindSession, Project: "other", Pool: "far", Fields: json.RawMessage(`{"agent":"fake"}`)})
	work := mkTicket(t, c, store.NewTicket{Title: "w", Pool: "dev"})
	appendJournal(t, c, sess.ID, 1, 8, clk.now(), work.ID)
	appendJournal(t, c, far.ID, 1, 3, clk.now(), "")

	pass(t, c, cl)
	assert.Equal(t, 8, countRows(t, c, "session:"+sess.ID))
	assert.Zero(t, shippedSeq(t, c, sess.ID), "the cursor is saved every few seconds, not per pass")
	clk.advance(shipSaveEvery)
	pass(t, c, cl)
	assert.Equal(t, int64(8), shippedSeq(t, c, sess.ID))

	rows := search(t, c, scope{admin: true}, "session:"+sess.ID+" seq:4")
	require.Len(t, rows, 1)
	r := rows[0]
	assert.Equal(t, "js", r["project"])
	assert.Equal(t, "dev", r["pool"])
	assert.Equal(t, "tool_call", r["update"])
	assert.Equal(t, "failed", r["status"])
	assert.Equal(t, "go test", r["tool"])
	assert.Equal(t, work.ID, r["ticket"])
	assert.NotContains(t, r["_msg"], "sk-supersecretvalue123", "redacted before shipping")

	// A pool caller with the grant sees only its project, whatever the query.
	pool := scope{actor: "s", session: "s", project: "js", pool: "searcher"}
	for _, q := range []string{"*", "project:other", `{project="other"}`, "* or project:other"} {
		for _, row := range search(t, c, pool, q) {
			assert.Equal(t, "js", row["project"], q)
		}
	}
	assert.NotEmpty(t, search(t, c, scope{admin: true}, "project:other"), "admin sees every project")
	_, err := c.journalSearch(ctx, scope{actor: "s", session: "s", project: "js", pool: "dev"}, rpc.JournalSearchParams{Query: "*"})
	require.ErrorIs(t, err, rpc.ErrForbidden, "no grant")
	_, err = c.journalSearch(ctx, scope{admin: true}, rpc.JournalSearchParams{Query: "bad((("})
	require.ErrorIs(t, err, store.ErrInvalid, "LogsQL syntax error")

	// Restart: a new controller state that never saved frames 9..12 resends
	// from the stored cursor; nothing is lost, doubles collapse by seq.
	appendJournal(t, c, sess.ID, 9, 12, clk.now(), work.ID)
	pass(t, c, cl)
	c.ship = shipper{}
	require.NoError(t, c.updateSessionFields(ctx, sess.ID, func(f *store.SessionFields) { f.ShippedSeq = 8 }))
	pass(t, c, cl)
	assert.Equal(t, 12, countRows(t, c, "session:"+sess.ID+" | uniq by (session, seq)"))

	// VictoriaLogs down: frames wait, lag grows past lag_max and shows in
	// attention; back up: they ship and the lag falls to 0.
	vl.Down()
	appendJournal(t, c, sess.ID, 13, 16, clk.now(), work.ID)
	c.ship.mu.Lock()
	_, err = c.shipPass(ctx, cl, true)
	c.ship.mu.Unlock()
	require.Error(t, err)
	clk.advance(2 * time.Minute)
	c.ship.mu.Lock()
	_, _ = c.shipPass(ctx, cl, true)
	c.ship.mu.Unlock()
	assert.Equal(t, 2*time.Minute, c.ship.lag["dev"])
	items, err := c.attentionItems(ctx)
	require.NoError(t, err)
	assert.True(t, hasAttention(items, rpc.AttentionJournalLag), "journal-lag in attention")
	assert.Contains(t, scrape(t, c), `jobshop_journal_ship_errors_total{`)

	vl.Up()
	pass(t, c, cl)
	assert.Zero(t, c.ship.lag["dev"])
	assert.Equal(t, 16, countRows(t, c, "session:"+sess.ID+" | uniq by (session, seq)"))
	items, err = c.attentionItems(ctx)
	require.NoError(t, err)
	assert.False(t, hasAttention(items, rpc.AttentionJournalLag))
	m := scrape(t, c)
	assert.Contains(t, m, `jobshop_journal_ship_frames_total{`)
	assert.Contains(t, m, `jobshop_journal_ship_bytes_total{`)
	assert.Contains(t, m, `jobshop_journal_ship_lag_seconds{`)
}

// TestJournalBackfill ships an old closed session from cursor 0.
func TestJournalBackfill(t *testing.T) {
	t.Parallel()
	vl := vltest.Start(t)
	c, clk := shipController(t, vl.URL)
	sess := mkTicket(t, c, store.NewTicket{Kind: store.KindSession, Pool: "dev", Fields: json.RawMessage(`{"agent":"fake"}`)})
	appendJournal(t, c, sess.ID, 1, 5, clk.now(), "")
	_, err := store.Close(t.Context(), c.store, "admin", sess.ID, nil, "done")
	require.NoError(t, err)
	clk.advance(48 * time.Hour) // past the startup pickup window

	out, err := c.journalBackfill(t.Context(), scope{admin: true}, rpc.JournalBackfillParams{})
	require.NoError(t, err)
	assert.Equal(t, rpc.JournalBackfillResult{Sessions: 1, Frames: 5}, out)
	assert.Equal(t, 5, countRows(t, c, "session:"+sess.ID))
	assert.Equal(t, int64(5), shippedSeq(t, c, sess.ID))

	out, err = c.journalBackfill(t.Context(), scope{admin: true}, rpc.JournalBackfillParams{})
	require.NoError(t, err)
	assert.Equal(t, rpc.JournalBackfillResult{}, out, "nothing new")
}

// TestJournalSearchNotConfigured: without journal.victorialogs, search
// and backfill give the clear error.
func TestJournalSearchNotConfigured(t *testing.T) {
	t.Parallel()
	c, _, _ := stallControllerYAML(t, journalYAML)
	_, err := c.journalSearch(t.Context(), scope{admin: true}, rpc.JournalSearchParams{Query: "*"})
	require.ErrorIs(t, err, store.ErrInvalid)
	assert.Contains(t, err.Error(), "journal search needs journal.victorialogs")
	_, err = c.journalBackfill(t.Context(), scope{admin: true}, rpc.JournalBackfillParams{})
	assert.Contains(t, err.Error(), "journal search needs journal.victorialogs")
}

func hasAttention(items []rpc.AttentionItem, kind string) bool {
	for _, it := range items {
		if it.Kind == kind {
			return true
		}
	}
	return false
}

func scrape(t *testing.T, c *Controller) string {
	t.Helper()
	rec := httptest.NewRecorder()
	c.metrics.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))
	b, err := io.ReadAll(rec.Body)
	require.NoError(t, err)
	return string(b)
}
