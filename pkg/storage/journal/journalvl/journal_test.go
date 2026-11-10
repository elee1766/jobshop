//go:build victorialogs

package journalvl_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control/controltest"
	"github.com/gfx-labs/jobshop/pkg/storage/journal/vltest"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// shipYAML has two projects, a pool with the journal-search grant and
// VictoriaLogs at %s. Sessions use agent "ext", which is not configured,
// so the controller spawns nothing and the test writes their journals.
const shipYAML = `projects:
  js: {}
  other: {}
default_project: js
agents:
  fake: {command: /bin/false}
pools:
  dev: {agent: fake, dir: $ROOT}
  searcher: {agent: fake, dir: $ROOT, admin: [journal-search]}
  far: {project: other, agent: fake, dir: $ROOT}
journal:
  victorialogs: {url: %q}
  lag_max: 1m
telemetry:
  metrics_listen: 127.0.0.1:0
`

func newSession(t *testing.T, c *controltest.Controller, project, pool string) rpc.SessionNewResult {
	t.Helper()
	var s rpc.SessionNewResult
	c.Call(rpc.MethodSessionNew, rpc.SessionNewParams{Project: project, Pool: pool, Agent: "ext"}, &s)
	return s
}

// appendJournal adds frames from seq from to to, each at clock time.
// Frame from is a prompt for ticket tk; the rest are agent chunks, with a
// tool call every fourth.
func appendJournal(t *testing.T, c *controltest.Controller, id string, from, to int, tk string) {
	t.Helper()
	dir := c.Root.SessionDir(id)
	require.NoError(t, os.MkdirAll(dir, 0o700))
	f, err := os.OpenFile(filepath.Join(dir, "journal.ndjson"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	require.NoError(t, err)
	defer func() { require.NoError(t, f.Close()) }()
	at := c.Clock.Now()
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

func search(t *testing.T, cl *rpc.Client, q string) []map[string]any {
	t.Helper()
	var out rpc.JournalSearchResult
	require.NoError(t, cl.Call(t.Context(), rpc.MethodJournalSearch, rpc.JournalSearchParams{Query: q}, &out))
	var rows []map[string]any
	for _, r := range out.Rows {
		var m map[string]any
		require.NoError(t, json.Unmarshal(r, &m))
		rows = append(rows, m)
	}
	return rows
}

func countRows(t *testing.T, c *controltest.Controller, q string) int {
	t.Helper()
	rows := search(t, c.Admin, q+" | stats count() n")
	require.Len(t, rows, 1)
	var n int
	_, err := fmt.Sscan(rows[0]["n"].(string), &n)
	require.NoError(t, err)
	return n
}

// backfill ships every tracked journal now and saves the cursors.
func backfill(t *testing.T, c *controltest.Controller) (rpc.JournalBackfillResult, error) {
	t.Helper()
	var out rpc.JournalBackfillResult
	err := c.Admin.Call(t.Context(), rpc.MethodJournalBackfill, rpc.JournalBackfillParams{}, &out)
	return out, err
}

func sessionFields(t *testing.T, c *controltest.Controller, id string) map[string]any {
	t.Helper()
	var sh rpc.ShowResult
	c.Call(rpc.MethodTicketShow, rpc.IDParams{ID: id}, &sh)
	f := map[string]any{}
	require.NoError(t, json.Unmarshal(sh.Ticket.Fields, &f))
	return f
}

func shippedSeq(t *testing.T, c *controltest.Controller, id string) int64 {
	t.Helper()
	n, _ := sessionFields(t, c, id)["shipped_seq"].(float64)
	return int64(n)
}

func hasAttention(t *testing.T, c *controltest.Controller, kind string) bool {
	t.Helper()
	var res rpc.AttentionResult
	c.Call(rpc.MethodAttention, struct{}{}, &res)
	for _, it := range res.Items {
		if it.Kind == kind {
			return true
		}
	}
	return false
}

func scrape(t *testing.T, c *controltest.Controller) string {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, c.MetricsURL(), nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(b)
}

// TestJournalShipSearch ships frames to a real VictoriaLogs and finds them:
// row fields, redaction, the project filter for a pool caller, a resend
// after a restart that lost the cursor, and catching up after an outage.
func TestJournalShipSearch(t *testing.T) {
	t.Parallel()
	vl := vltest.Start(t)
	root := controltest.NewRoot(t, fmt.Sprintf(shipYAML, vl.URL))
	c := controltest.Start(t, root, control.Options{})
	sess := newSession(t, c, "js", "dev")
	far := newSession(t, c, "other", "far")
	const work = "js-work"
	appendJournal(t, c, sess.Ticket.ID, 1, 8, work)
	appendJournal(t, c, far.Ticket.ID, 1, 3, "")

	_, err := backfill(t, c)
	require.NoError(t, err)
	id := sess.Ticket.ID
	assert.Equal(t, 8, countRows(t, c, "session:"+id))
	assert.Equal(t, int64(8), shippedSeq(t, c, id))

	rows := search(t, c.Admin, "session:"+id+" seq:4")
	require.Len(t, rows, 1)
	r := rows[0]
	assert.Equal(t, "js", r["project"])
	assert.Equal(t, "dev", r["pool"])
	assert.Equal(t, "tool_call", r["update"])
	assert.Equal(t, "failed", r["status"])
	assert.Equal(t, "go test", r["tool"])
	assert.Equal(t, work, r["ticket"])
	assert.NotContains(t, r["_msg"], "sk-supersecretvalue123", "redacted before shipping")

	// A session in a pool with the grant sees only its project, whatever
	// the query.
	searcher := c.Dial(newSession(t, c, "js", "searcher").Socket)
	for _, q := range []string{"*", "project:other", `{project="other"}`, "* or project:other"} {
		for _, row := range search(t, searcher, q) {
			assert.Equal(t, "js", row["project"], q)
		}
	}
	assert.NotEmpty(t, search(t, c.Admin, "project:other"), "admin sees every project")
	plain := c.Dial(sess.Socket)
	err = plain.Call(t.Context(), rpc.MethodJournalSearch, rpc.JournalSearchParams{Query: "*"}, nil)
	require.ErrorIs(t, err, rpc.ErrForbidden, "no grant")
	err = c.Admin.Call(t.Context(), rpc.MethodJournalSearch, rpc.JournalSearchParams{Query: "bad((("}, nil)
	require.ErrorIs(t, err, domain.ErrInvalid, "LogsQL syntax error")

	// Restart with a cursor behind what was shipped (a crash before the
	// save): frames 9..12 are sent again from the stored cursor, nothing
	// is lost, and doubles collapse by seq.
	appendJournal(t, c, id, 9, 12, work)
	_, err = backfill(t, c)
	require.NoError(t, err)
	f := sessionFields(t, c, id)
	f["shipped_seq"] = 8
	b, err := json.Marshal(f)
	require.NoError(t, err)
	c.Call(rpc.MethodTicketUpdate, rpc.UpdateParams{ID: id, Fields: b}, nil)
	c.Stop()
	c = controltest.Start(t, root, control.Options{Clock: c.Clock})
	_, err = backfill(t, c)
	require.NoError(t, err)
	assert.Equal(t, 12, countRows(t, c, "session:"+id+" | uniq by (session, seq)"))
	assert.Equal(t, 16, countRows(t, c, "session:"+id), "9..12 sent twice")

	// VictoriaLogs down: frames wait, lag grows past lag_max and shows in
	// attention; back up: they ship and the lag falls to 0.
	vl.Down()
	appendJournal(t, c, id, 13, 16, work)
	_, err = backfill(t, c)
	require.Error(t, err)
	c.Clock.Advance(2 * time.Minute)
	_, _ = backfill(t, c)
	assert.True(t, hasAttention(t, c, rpc.AttentionJournalLag), "journal-lag in attention")
	assert.Contains(t, scrape(t, c), `jobshop_journal_ship_errors_total{`)

	vl.Up()
	_, err = backfill(t, c)
	require.NoError(t, err)
	assert.Equal(t, 16, countRows(t, c, "session:"+id+" | uniq by (session, seq)"))
	assert.False(t, hasAttention(t, c, rpc.AttentionJournalLag), "caught up")
	m := scrape(t, c)
	for _, name := range []string{"frames_total", "bytes_total", "lag_seconds"} {
		assert.True(t, strings.Contains(m, "jobshop_journal_ship_"+name+"{"), name)
	}
}

// TestJournalBackfill ships a session closed before the controller
// started, past the startup pickup window, from cursor 0.
func TestJournalBackfill(t *testing.T) {
	t.Parallel()
	vl := vltest.Start(t)
	root := controltest.NewRoot(t, fmt.Sprintf(shipYAML, vl.URL))
	c := controltest.Start(t, root, control.Options{})
	sess := newSession(t, c, "js", "dev")
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: sess.Ticket.ID, Resolution: "done"}, nil)
	appendJournal(t, c, sess.Ticket.ID, 1, 5, "")
	c.Stop()

	// A restart two days later does not pick it up by itself.
	c.Clock.Advance(48 * time.Hour)
	c = controltest.Start(t, root, control.Options{Clock: c.Clock})
	out, err := backfill(t, c)
	require.NoError(t, err)
	assert.Equal(t, rpc.JournalBackfillResult{Sessions: 1, Frames: 5}, out)
	assert.Equal(t, 5, countRows(t, c, "session:"+sess.Ticket.ID))
	assert.Equal(t, int64(5), shippedSeq(t, c, sess.Ticket.ID))

	out, err = backfill(t, c)
	require.NoError(t, err)
	assert.Equal(t, rpc.JournalBackfillResult{}, out, "nothing new")
}
