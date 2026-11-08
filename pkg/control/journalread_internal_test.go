package control

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// searchYAML adds a second project and a pool with the journal-search
// grant to journalYAML. vl is the VictoriaLogs url, or empty for none.
func searchYAML(vl string) string {
	y := strings.Replace(journalYAML, "projects:\n  js: {}\n", "default_project: js\nprojects:\n  js: {}\n  zz: {}\n", 1)
	y = strings.Replace(y, "pools:\n", "pools:\n  searcher: {agent: fake, dir: %[1]q, admin: [journal-search]}\n  far: {project: zz, agent: fake, dir: %[1]q}\n", 1)
	if vl != "" {
		y += fmt.Sprintf("journal:\n  victorialogs: {url: %q}\n", vl)
	}
	return y
}

// TestJournalSearchAccess: the grant and the project scope are decided in
// the controller, and journalread gets only the result: a pool's search
// carries its project as the stream filter, admin's carries none, a pool
// without the grant never reaches VictoriaLogs.
func TestJournalSearchAccess(t *testing.T) {
	t.Parallel()
	var mu sync.Mutex
	var forms []map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		got := map[string]string{}
		for k := range r.PostForm {
			got[k] = r.PostForm.Get(k)
		}
		mu.Lock()
		forms = append(forms, got)
		mu.Unlock()
		for i := range 3 {
			_, _ = fmt.Fprintf(w, `{"_msg":"row %d"}`+"\n", i)
		}
	}))
	t.Cleanup(srv.Close)
	c, _, _ := scanControllerYAML(t, searchYAML(srv.URL))
	ctx := t.Context()

	searcher := scope{actor: "s", session: "s", project: "js", pool: "searcher"}
	res, err := c.journalSearch(ctx, searcher, rpc.JournalSearchParams{Query: "*", Limit: 2})
	require.NoError(t, err)
	assert.Len(t, res.(rpc.JournalSearchResult).Rows, 2)
	assert.True(t, res.(rpc.JournalSearchResult).Truncated)

	_, err = c.journalSearch(ctx, scope{admin: true}, rpc.JournalSearchParams{Query: "*"})
	require.NoError(t, err)

	nogrant := scope{actor: "d", session: "d", project: "js", pool: "dev"}
	_, err = c.journalSearch(ctx, nogrant, rpc.JournalSearchParams{Query: "*"})
	require.ErrorIs(t, err, rpc.ErrForbidden)

	_, err = c.journalSearch(ctx, scope{admin: true}, rpc.JournalSearchParams{Query: " "})
	require.ErrorIs(t, err, store.ErrInvalid)

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, forms, 2, "the forbidden and the blank query never reached VictoriaLogs")
	assert.Equal(t, `{project="js"}`, forms[0]["extra_stream_filters"])
	assert.NotContains(t, forms[1], "extra_stream_filters")
}

// TestRecallJournalScope: journal hits come from the summaries indexJournal
// wrote, read through journalread.Summary from a plain and from a
// compressed journal, and a session's recall sees only its project's.
func TestRecallJournalScope(t *testing.T) {
	t.Parallel()
	c, _, _ := scanControllerYAML(t, searchYAML(""))
	ctx := t.Context()
	mk := func(project, pool string) store.Ticket {
		s := mkTicket(t, c, store.NewTicket{Project: project, Kind: store.KindSession, Pool: pool, Fields: json.RawMessage(`{"agent":"fake"}`)})
		writeJournal(t, c, s.ID, 5)
		r := &runner{c: c, id: s.ID, pool: pool, ctx: ctx, dir: c.root.SessionDir(s.ID), log: c.log}
		r.indexJournal(ctx, s.ID)
		return s
	}
	mine, theirs := mk("js", "dev"), mk("zz", "far")

	refs := func(sc scope, p rpc.RecallParams) []string {
		t.Helper()
		out, err := c.recall(ctx, sc, p)
		require.NoError(t, err)
		var r []string
		for _, h := range out.([]store.RecallHit) {
			r = append(r, h.Ref)
		}
		return r
	}
	p := rpc.RecallParams{Query: "hi", Kind: store.RecallJournal}
	assert.ElementsMatch(t, []string{mine.ID, theirs.ID}, refs(scope{admin: true}, p))
	assert.Equal(t, []string{mine.ID}, refs(scope{actor: mine.ID, session: mine.ID, project: "js", pool: "dev"}, p))
	_, err := c.recall(ctx, scope{actor: mine.ID, session: mine.ID, project: "js", pool: "dev"}, rpc.RecallParams{Query: "hi", Project: "zz"})
	require.ErrorIs(t, err, rpc.ErrForbidden)
}
