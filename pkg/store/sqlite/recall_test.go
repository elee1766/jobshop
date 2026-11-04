package sqlite_test

import (
	"database/sql"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite"
)

func TestRecallNewCommentSearchable(t *testing.T) {
	ctx := t.Context()
	s, err := sqlite.Open(ctx, filepath.Join(t.TempDir(), "j.db"), sqlite.Options{})
	require.NoError(t, err)
	defer func() { _ = s.Close() }()
	tk, err := s.Create(ctx, "admin", store.NewTicket{Project: "p", Kind: store.KindTask, Title: "t"})
	require.NoError(t, err)
	_, err = s.AddComment(ctx, "a", tk.ID, "CI failed: golangci-lint gofumpt formatting in pkg/cli")
	require.NoError(t, err)
	_, err = s.AddComment(ctx, "a", tk.ID, "unrelated note about the weather")
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

func TestRecallRanking(t *testing.T) {
	ctx := t.Context()
	s, err := sqlite.Open(ctx, filepath.Join(t.TempDir(), "j.db"), sqlite.Options{})
	require.NoError(t, err)
	defer func() { _ = s.Close() }()
	at := time.UnixMilli(1_700_000_000_000)
	require.NoError(t, s.IndexRecall(ctx, store.RecallDoc{
		Kind: store.RecallJournal, Ref: "s1:1", Project: "p", At: at,
		Body: "vendor hash mismatch once, then lots of other text about reviewing and merging and waiting",
	}))
	require.NoError(t, s.IndexRecall(ctx, store.RecallDoc{
		Kind: store.RecallJournal, Ref: "s2:1", Project: "p", At: at,
		Body: "nix vendor hash mismatch: set vendorHash to the got value",
	}))
	hits, err := s.Recall(ctx, store.RecallQuery{Text: "vendor hash mismatch"})
	require.NoError(t, err)
	require.Len(t, hits, 2)
	require.Equal(t, "s2:1", hits[0].Ref)
	require.Equal(t, at.UTC(), hits[0].At)

	// Re-indexing replaces; an empty body removes.
	require.NoError(t, s.IndexRecall(ctx, store.RecallDoc{Kind: store.RecallJournal, Ref: "s2:1", Body: ""}))
	hits, err = s.Recall(ctx, store.RecallQuery{Text: "vendor", Kind: store.RecallJournal})
	require.NoError(t, err)
	require.Len(t, hits, 1)
	require.Equal(t, "s1:1", hits[0].Ref)
}

// TestRecallBackfill writes comments before the recall migration and checks
// the migration indexes them.
func TestRecallBackfill(t *testing.T) {
	ctx := t.Context()
	path := filepath.Join(t.TempDir(), "j.db")
	s, err := sqlite.Open(ctx, path, sqlite.Options{})
	require.NoError(t, err)
	tk, err := s.Create(ctx, "admin", store.NewTicket{Project: "p", Kind: store.KindTask, Title: "t"})
	require.NoError(t, err)
	_, err = s.AddComment(ctx, "a", tk.ID, "flaky TestFollowJournal timeout")
	require.NoError(t, err)
	require.NoError(t, s.Close())

	// Roll the index back to before 00004 (and later migrations), as on a
	// database from an older release, then reopen to migrate forward.
	db, err := sql.Open("sqlite", "file:"+path)
	require.NoError(t, err)
	for _, q := range []string{
		"ALTER TABLE learnings DROP COLUMN pr", "DELETE FROM goose_db_version WHERE version_id = 5",
		"DROP TRIGGER recall_learning_delete", "DROP TRIGGER recall_learning_update",
		"DROP TRIGGER recall_learning_insert",
		"DROP TRIGGER recall_comment_delete", "DROP TRIGGER recall_comment_update",
		"DROP TRIGGER recall_comment_insert", "DROP TABLE recall",
		"DELETE FROM goose_db_version WHERE version_id = 4",
	} {
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

// TestRecallLearnings: learnings are indexed on insert; a global learning
// (no ticket, no project) is visible to every project.
func TestRecallLearnings(t *testing.T) {
	ctx := t.Context()
	s, err := sqlite.Open(ctx, filepath.Join(t.TempDir(), "j.db"), sqlite.Options{})
	require.NoError(t, err)
	defer func() { _ = s.Close() }()
	l, err := s.AddLearning(ctx, store.Learning{
		Kind: store.LearningWorkaround, Summary: "nix vendorHash mismatch", Fix: "use lib.fakeHash then paste got",
		Source: store.LearningSourceAgent,
	})
	require.NoError(t, err)
	hits, err := s.Recall(ctx, store.RecallQuery{Text: "fakeHash", Project: "p"})
	require.NoError(t, err)
	require.Len(t, hits, 1)
	require.Equal(t, store.RecallLearning, hits[0].Kind)
	require.Equal(t, strconv.FormatInt(l.ID, 10), hits[0].Ref)
}
