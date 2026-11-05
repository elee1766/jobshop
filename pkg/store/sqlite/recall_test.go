package sqlite_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite"
)

func TestRecallNewCommentSearchable(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, err := sqlite.Open(ctx, testPath(t), sqlite.Options{})
	require.NoError(t, err)
	defer func() { _ = s.Close() }()
	tk, err := store.Create(ctx, s, "admin", store.NewTicket{Project: "p", Kind: store.KindTask, Title: "t"})
	require.NoError(t, err)
	_, err = store.AddComment(ctx, s, "a", tk.ID, "CI failed: golangci-lint gofumpt formatting in pkg/cli")
	require.NoError(t, err)
	_, err = store.AddComment(ctx, s, "a", tk.ID, "unrelated note about the weather")
	require.NoError(t, err)

	hits, err := s.Recall(ctx, store.RecallQuery{Text: "gofumpt formatting"})
	require.NoError(t, err)
	require.Len(t, hits, 1)
	require.Equal(t, store.RecallComment, hits[0].Kind)
	require.Equal(t, tk.ID, hits[0].TicketID)
	require.Equal(t, "p", hits[0].Project)
	require.Contains(t, hits[0].Snippet, "[gofumpt]")
	require.False(t, hits[0].At.IsZero())

	// Any-word fallback and FTS syntax in the query do not fail.
	hits, err = s.Recall(ctx, store.RecallQuery{Text: `weather "nothing-matches" AND:`})
	require.NoError(t, err)
	require.Len(t, hits, 1)

	// Project filter.
	hits, err = s.Recall(ctx, store.RecallQuery{Text: "gofumpt", Project: "other"})
	require.NoError(t, err)
	require.Empty(t, hits)
}

// TestRecallBackfill writes comments before the recall migration and checks
// the migration indexes them.
func TestRecallBackfill(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	path := testPath(t)
	s, err := sqlite.Open(ctx, path, sqlite.Options{})
	require.NoError(t, err)
	tk, err := store.Create(ctx, s, "admin", store.NewTicket{Project: "p", Kind: store.KindTask, Title: "t"})
	require.NoError(t, err)
	_, err = store.AddComment(ctx, s, "a", tk.ID, "flaky TestFollowJournal timeout")
	require.NoError(t, err)
	require.NoError(t, s.Close())

	// Roll the index back to before 00004 (and later migrations), as on a
	// database from an older release, then reopen to migrate forward.
	db, err := openAt21(t, path)
	require.NoError(t, err)
	for _, q := range append(undo00013(), []string{
		"DROP TABLE resource_caps", "DELETE FROM goose_db_version WHERE version_id = 12",
		"DELETE FROM goose_db_version WHERE version_id = 11",
		"DROP INDEX inbox_natural_key", "DELETE FROM goose_db_version WHERE version_id = 10",
		"DELETE FROM goose_db_version WHERE version_id = 9",
		"DROP TABLE inbox", "DELETE FROM goose_db_version WHERE version_id = 8",
		"DROP TABLE write_keys", "DELETE FROM goose_db_version WHERE version_id = 7",
		"DROP TABLE memory_edits", "ALTER TABLE learnings DROP COLUMN gate",
		"DELETE FROM goose_db_version WHERE version_id = 6",
		"ALTER TABLE learnings DROP COLUMN pr", "DELETE FROM goose_db_version WHERE version_id = 5",
		"DROP TRIGGER recall_learning_delete", "DROP TRIGGER recall_learning_update",
		"DROP TRIGGER recall_learning_insert",
		"DROP TRIGGER recall_comment_delete", "DROP TRIGGER recall_comment_update",
		"DROP TRIGGER recall_comment_insert", "DROP TABLE recall",
		"DELETE FROM goose_db_version WHERE version_id = 4",
	}...) {
		_, err = db.ExecContext(ctx, q)
		require.NoError(t, err, q)
	}
	require.NoError(t, db.Close())

	s, err = sqlite.Open(ctx, path, sqlite.Options{})
	require.NoError(t, err)
	defer func() { _ = s.Close() }()
	hits, err := s.Recall(ctx, store.RecallQuery{Text: "flaky timeout"})
	require.NoError(t, err)
	require.Len(t, hits, 1)
	require.Equal(t, tk.ID, hits[0].TicketID)
}
