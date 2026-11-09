package journalread_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/control/journalread"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/journal"
	"github.com/elee1766/jobshop/pkg/redact"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// writeJournal writes n frames: odd seqs are agent message chunks, even
// seqs controller prompts. Frame 3 carries a secret-named key.
func writeJournal(t *testing.T, dir string, n int) {
	t.Helper()
	var b strings.Builder
	for i := 1; i <= n; i++ {
		from, msg := "controller", `{"jsonrpc":"2.0","id":1,"method":"session/prompt","params":{}}`
		if i%2 == 1 {
			from, msg = "agent", fmt.Sprintf(`{"jsonrpc":"2.0","method":"session/update","params":{"update":{"sessionUpdate":"agent_message_chunk","content":{"text":"hi %d"}}}}`, i)
		}
		if i == 3 {
			msg = `{"jsonrpc":"2.0","method":"session/update","params":{"update":{"sessionUpdate":"agent_message_chunk","api_key":"sk-supersecretvalue123"}}}`
		}
		fmt.Fprintf(&b, `{"seq":%d,"t":%d,"from":%q,"msg":%s}`+"\n", i, 1000+i, from, msg)
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, host.JournalFile), []byte(b.String()), 0o600))
}

func seqs(r rpc.SessionJournalResult) []int64 {
	var s []int64
	for _, f := range r.Frames {
		s = append(s, f.Seq)
	}
	return s
}

func access(dir string) journalread.Access {
	return journalread.Access{Dir: dir, Redact: redact.New(nil)}
}

func TestPage(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeJournal(t, dir, 10)
	a := access(dir)
	page := func(p rpc.SessionJournalParams) rpc.SessionJournalResult {
		t.Helper()
		r, err := journalread.Page(a, p)
		require.NoError(t, err)
		return r
	}

	// Paging by seq: the cursor resumes where the page ended.
	r := page(rpc.SessionJournalParams{Limit: 4})
	assert.Equal(t, []int64{1, 2, 3, 4}, seqs(r))
	assert.False(t, r.Done)
	assert.Equal(t, int64(4), r.NextSeq)
	r = page(rpc.SessionJournalParams{SinceSeq: r.NextSeq, Limit: 4})
	assert.Equal(t, []int64{5, 6, 7, 8}, seqs(r))
	r = page(rpc.SessionJournalParams{SinceSeq: r.NextSeq, Limit: 4})
	assert.Equal(t, []int64{9, 10}, seqs(r))
	assert.True(t, r.Done)
	assert.Equal(t, int64(10), r.NextSeq)

	// Filters still move the cursor past frames they skip.
	r = page(rpc.SessionJournalParams{Kinds: []string{"session/prompt"}})
	assert.Equal(t, []int64{2, 4, 6, 8, 10}, seqs(r))
	r = page(rpc.SessionJournalParams{From: "agent", Limit: 2})
	assert.Equal(t, []int64{1, 3}, seqs(r))
	r = page(rpc.SessionJournalParams{Kinds: []string{"nothing"}})
	assert.Empty(t, r.Frames)
	assert.True(t, r.Done)
	assert.Equal(t, int64(10), r.NextSeq)

	// Redaction is the Access's.
	r = page(rpc.SessionJournalParams{SinceSeq: 2, Limit: 1})
	require.Len(t, r.Frames, 1)
	assert.NotContains(t, string(r.Frames[0].Msg), "supersecretvalue")
	assert.True(t, json.Valid(r.Frames[0].Msg))

	// No journal file reads as empty.
	r, err := journalread.Page(access(t.TempDir()), rpc.SessionJournalParams{SinceSeq: 3})
	require.NoError(t, err)
	assert.Empty(t, r.Frames)
	assert.True(t, r.Done)
	assert.Equal(t, int64(3), r.NextSeq)
}

// TestPageCompressed: a compressed journal pages the same as the plain
// one, including from the middle.
func TestPageCompressed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeJournal(t, dir, 10)
	require.NoError(t, host.CompressJournal(dir, 3))
	require.NoFileExists(t, filepath.Join(dir, host.JournalFile))

	a := access(dir)
	r, err := journalread.Page(a, rpc.SessionJournalParams{SinceSeq: 6})
	require.NoError(t, err)
	assert.Equal(t, []int64{7, 8, 9, 10}, seqs(r))
	assert.NotContains(t, string(r.Frames[0].Msg), "sk-supersecret")

	r, err = journalread.Page(a, rpc.SessionJournalParams{SinceSeq: 1, Limit: 2})
	require.NoError(t, err)
	assert.Equal(t, []int64{2, 3}, seqs(r))
	assert.NotContains(t, string(r.Frames[1].Msg), "supersecretvalue", "frame 3 holds the secret key")
	assert.False(t, r.Done)

	text, at, ok := journalread.Summary(dir)
	require.True(t, ok)
	assert.Contains(t, text, "hi 5")
	assert.Equal(t, time.UnixMilli(1010), at)
}

func TestValidatePage(t *testing.T) {
	t.Parallel()
	require.NoError(t, journalread.ValidatePage(rpc.SessionJournalParams{From: "agent", SinceSeq: 0}))
	require.NoError(t, journalread.ValidatePage(rpc.SessionJournalParams{From: "controller"}))
	require.ErrorIs(t, journalread.ValidatePage(rpc.SessionJournalParams{From: "nobody"}), domain.ErrInvalid)
	require.ErrorIs(t, journalread.ValidatePage(rpc.SessionJournalParams{SinceSeq: -1}), domain.ErrInvalid)
}

func TestSummaryEmpty(t *testing.T) {
	t.Parallel()
	_, _, ok := journalread.Summary(t.TempDir())
	assert.False(t, ok, "no journal")
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, host.JournalFile), nil, 0o600))
	_, _, ok = journalread.Summary(dir)
	assert.False(t, ok, "no frames")
}

// vlServer is a VictoriaLogs query endpoint that records the form it got
// and answers with n JSON rows, or status when it is not 200.
func vlServer(t *testing.T, n, status int, form *map[string]string) *journal.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/select/logsql/query", r.URL.Path)
		require.NoError(t, r.ParseForm())
		got := map[string]string{}
		for k := range r.PostForm {
			got[k] = r.PostForm.Get(k)
		}
		*form = got
		if status != http.StatusOK {
			http.Error(w, "syntax error", status)
			return
		}
		for i := range n {
			_, _ = fmt.Fprintf(w, `{"_msg":"row %d"}`+"\n", i)
		}
	}))
	t.Cleanup(srv.Close)
	return &journal.Client{URL: srv.URL}
}

func TestRun(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)
	var form map[string]string
	cl := vlServer(t, 5, http.StatusOK, &form)

	// Truncation: one more row than asked tells the caller there is more.
	res, err := journalread.Run(t.Context(), journalread.Search{
		Client: cl, Project: "js", Now: now,
		JournalSearchParams: rpc.JournalSearchParams{Query: "*", Limit: 3, Since: time.Hour},
	})
	require.NoError(t, err)
	assert.Len(t, res.Rows, 3)
	assert.True(t, res.Truncated)
	assert.Equal(t, "4", form["limit"])
	assert.Equal(t, `{project="js"}`, form["extra_stream_filters"], "scope is outside the query text")
	assert.Equal(t, now.Add(-time.Hour).Format(time.RFC3339Nano), form["start"])

	res, err = journalread.Run(t.Context(), journalread.Search{Client: cl, Now: now, JournalSearchParams: rpc.JournalSearchParams{Query: "*", Limit: 5}})
	require.NoError(t, err)
	assert.Len(t, res.Rows, 5)
	assert.False(t, res.Truncated)
	assert.NotContains(t, form, "extra_stream_filters", "empty project is every project")

	// Limits default and cap.
	_, err = journalread.Run(t.Context(), journalread.Search{Client: cl, Now: now, JournalSearchParams: rpc.JournalSearchParams{Query: "*"}})
	require.NoError(t, err)
	assert.Equal(t, "1001", form["limit"])
	_, err = journalread.Run(t.Context(), journalread.Search{Client: cl, Now: now, JournalSearchParams: rpc.JournalSearchParams{Query: "*", Limit: 1 << 20}})
	require.NoError(t, err)
	assert.Equal(t, "10001", form["limit"])

	// An explicit start beats since.
	start := now.Add(-48 * time.Hour)
	_, err = journalread.Run(t.Context(), journalread.Search{Client: cl, Now: now, JournalSearchParams: rpc.JournalSearchParams{Query: "*", Start: start, Since: time.Hour}})
	require.NoError(t, err)
	assert.Equal(t, start.Format(time.RFC3339Nano), form["start"])

	// A blank query is the caller's error.
	_, err = journalread.Run(t.Context(), journalread.Search{Client: cl, Now: now, JournalSearchParams: rpc.JournalSearchParams{Query: "  "}})
	require.ErrorIs(t, err, domain.ErrInvalid)
}

func TestRunErrors(t *testing.T) {
	t.Parallel()
	var form map[string]string
	// A 400 is a LogsQL syntax error the caller made.
	_, err := journalread.Run(t.Context(), journalread.Search{
		Client: vlServer(t, 0, http.StatusBadRequest, &form), JournalSearchParams: rpc.JournalSearchParams{Query: "bad("},
	})
	require.ErrorIs(t, err, domain.ErrInvalid)
	assert.Contains(t, err.Error(), "syntax error")
	// Any other status is not.
	_, err = journalread.Run(t.Context(), journalread.Search{
		Client: vlServer(t, 0, http.StatusBadGateway, &form), JournalSearchParams: rpc.JournalSearchParams{Query: "*"},
	})
	require.Error(t, err)
	assert.NotErrorIs(t, err, domain.ErrInvalid)

	// A cancelled context stops the query.
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = journalread.Run(ctx, journalread.Search{Client: vlServer(t, 1, http.StatusOK, &form), JournalSearchParams: rpc.JournalSearchParams{Query: "*"}})
	require.Error(t, err)
}
