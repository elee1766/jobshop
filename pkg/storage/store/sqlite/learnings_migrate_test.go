package sqlite_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/knowledge"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/storage/store/sqlite"
)

// TestLearningsMigrate: rows in the old learnings table are in the
// learnings collection after 00022 (#579), with their ids, status, the
// ticket's project and a search index; Down puts them back.
func TestLearningsMigrate(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	path := testPath(t)
	s, err := sqlite.Open(ctx, path, sqlite.Options{})
	require.NoError(t, err)
	tk, err := store.Create(ctx, s, "admin", store.NewTicket{Project: "p", Kind: domain.KindTask, Title: "t"})
	require.NoError(t, err)
	require.NoError(t, s.Close())

	db, err := openAt21(t, path)
	require.NoError(t, err)
	for _, q := range []string{
		`INSERT INTO learnings (id, kind, summary, fix, scope, ticket, session, source, status, created_at, pr, gate)
		 VALUES (7, 'workaround', 'nix vendorHash mismatch', 'paste got', 'repo:a/b', '` + tk.ID + `', 'sess', 'agent', 'applied', 1000, 'https://x/pull/1', 'helped: x')`,
		`INSERT INTO learnings (id, kind, summary, source, created_at) VALUES (12, 'failure', 'flaky dial timeout', 'controller', 2000)`,
	} {
		_, err = db.ExecContext(ctx, q)
		require.NoError(t, err, q)
	}
	require.NoError(t, db.Close())

	check := func() {
		t.Helper()
		s, err := sqlite.Open(ctx, path, sqlite.Options{})
		require.NoError(t, err)
		defer func() { _ = s.Close() }()
		l, err := knowledge.GetLearning(ctx, s, 7)
		require.NoError(t, err)
		assert.Equal(t, "workaround", l.Kind)
		assert.Equal(t, "paste got", l.Fix)
		assert.Equal(t, "repo:a/b", l.Scope)
		assert.Equal(t, tk.ID, l.Ticket)
		assert.Equal(t, "sess", l.Session)
		assert.Equal(t, domain.LearningApplied, l.Status)
		assert.Equal(t, "https://x/pull/1", l.PR)
		assert.Equal(t, "helped: x", l.Gate)
		assert.Equal(t, time.UnixMilli(1000).UTC(), l.CreatedAt)
		hits, err := knowledge.RecallLearnings(ctx, s, "vendorHash", "other", "", 0)
		require.NoError(t, err)
		assert.Empty(t, hits, "project of the ticket kept")
		hits, err = knowledge.RecallLearnings(ctx, s, "flaky", "other", "", 0)
		require.NoError(t, err)
		assert.Len(t, hits, 1, "global learning in every project")
		idx, err := s.Recall(ctx, store.RecallQuery{Text: "flaky"})
		require.NoError(t, err)
		assert.Empty(t, idx, "learning rows left the shared recall index")
		n, err := knowledge.AddLearning(ctx, s, domain.Learning{Kind: "failure", Summary: "new", Source: "agent"}, "")
		require.NoError(t, err)
		assert.Equal(t, int64(13), n.ID)
	}
	check()

	// Down and up again: the rows survive the round trip.
	db, err = openAt21(t, path)
	require.NoError(t, err)
	var cnt int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM learnings").Scan(&cnt))
	assert.Equal(t, 3, cnt)
	require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM recall WHERE kind = 'learning'").Scan(&cnt))
	assert.Equal(t, 3, cnt)
	_, err = db.ExecContext(ctx, "DELETE FROM learnings WHERE id = 13")
	require.NoError(t, err)
	require.NoError(t, db.Close())
	check()
}
