package control

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

var journalYAML = strings.Replace(scanYAML, "pools:\n", "pools:\n  reader: {agent: fake, dir: %[1]q, admin: [session-journal]}\n", 1)

// writeJournal writes n frames: odd seqs are agent message chunks, even
// seqs controller prompts. Frame 3 carries a secret-named key.
func writeJournal(t *testing.T, c *Controller, id string, n int) {
	t.Helper()
	dir := c.root.SessionDir(id)
	require.NoError(t, os.MkdirAll(dir, 0o700))
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
	require.NoError(t, os.WriteFile(filepath.Join(dir, "journal.ndjson"), []byte(b.String()), 0o600))
}

func TestSessionJournal(t *testing.T) {
	t.Parallel()
	c, _, _ := scanControllerYAML(t, journalYAML)
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	sess := mkTicket(t, c, store.NewTicket{Kind: domain.KindSession, Pool: "dev", Fields: json.RawMessage(`{"agent":"fake"}`)})
	writeJournal(t, c, sess.ID, 10)

	page := func(sc scope, p rpc.SessionJournalParams) rpc.SessionJournalResult {
		t.Helper()
		out, err := c.sessionJournal(ctx, sc, p)
		require.NoError(t, err)
		return out.(rpc.SessionJournalResult)
	}
	seqs := func(r rpc.SessionJournalResult) []int64 {
		var s []int64
		for _, f := range r.Frames {
			s = append(s, f.Seq)
		}
		return s
	}

	// Paging by seq.
	r := page(admin, rpc.SessionJournalParams{ID: sess.ID, Limit: 4})
	assert.Equal(t, []int64{1, 2, 3, 4}, seqs(r))
	assert.False(t, r.Done)
	assert.Equal(t, int64(4), r.NextSeq)
	r = page(admin, rpc.SessionJournalParams{ID: sess.ID, SinceSeq: r.NextSeq, Limit: 4})
	assert.Equal(t, []int64{5, 6, 7, 8}, seqs(r))
	r = page(admin, rpc.SessionJournalParams{ID: sess.ID, SinceSeq: r.NextSeq, Limit: 4})
	assert.Equal(t, []int64{9, 10}, seqs(r))
	assert.True(t, r.Done)
	assert.Equal(t, int64(10), r.NextSeq)

	// Kinds and from filters.
	r = page(admin, rpc.SessionJournalParams{ID: sess.ID, Kinds: []string{"session/prompt"}})
	assert.Equal(t, []int64{2, 4, 6, 8, 10}, seqs(r))
	r = page(admin, rpc.SessionJournalParams{ID: sess.ID, Kinds: []string{"agent_message_chunk"}, Limit: 2})
	assert.Equal(t, []int64{1, 3}, seqs(r))
	r = page(admin, rpc.SessionJournalParams{ID: sess.ID, From: "agent"})
	assert.Equal(t, []int64{1, 3, 5, 7, 9}, seqs(r))

	// Redaction.
	r = page(admin, rpc.SessionJournalParams{ID: sess.ID, SinceSeq: 2, Limit: 1})
	require.Len(t, r.Frames, 1)
	assert.NotContains(t, string(r.Frames[0].Msg), "supersecretvalue")
	assert.True(t, json.Valid(r.Frames[0].Msg))

	// A session reads its own; another session without the grant is refused.
	own := scope{actor: sess.ID, session: sess.ID, project: "js", pool: "dev"}
	assert.Len(t, page(own, rpc.SessionJournalParams{ID: sess.ID}).Frames, 10)
	other := mkTicket(t, c, store.NewTicket{Kind: domain.KindSession, Pool: "dev", Fields: json.RawMessage(`{"agent":"fake"}`)})
	_, err := c.sessionJournal(ctx, scope{actor: other.ID, session: other.ID, project: "js", pool: "dev"}, rpc.SessionJournalParams{ID: sess.ID})
	require.ErrorIs(t, err, rpc.ErrForbidden)

	// A granted pool reads sessions in its project, not another project's.
	reader := scope{actor: "r", session: "r", project: "js", pool: "reader"}
	assert.Len(t, page(reader, rpc.SessionJournalParams{ID: sess.ID}).Frames, 10)
	foreign := mkTicket(t, c, store.NewTicket{Project: "zz", Kind: domain.KindSession, Pool: "dev", Fields: json.RawMessage(`{"agent":"fake"}`)})
	writeJournal(t, c, foreign.ID, 2)
	_, err = c.sessionJournal(ctx, reader, rpc.SessionJournalParams{ID: foreign.ID})
	require.ErrorIs(t, err, rpc.ErrForbidden)

	// An ended session's journal is still readable.
	_, err = store.Close(ctx, c.store, "admin", sess.ID, nil, "done")
	require.NoError(t, err)
	assert.Len(t, page(admin, rpc.SessionJournalParams{ID: sess.ID}).Frames, 10)

	// Not a session.
	task := mkTicket(t, c, store.NewTicket{})
	_, err = c.sessionJournal(ctx, admin, rpc.SessionJournalParams{ID: task.ID})
	require.ErrorIs(t, err, domain.ErrInvalid)
}
