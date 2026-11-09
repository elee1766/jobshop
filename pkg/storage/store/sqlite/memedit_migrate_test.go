package sqlite_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/memedit"
	"github.com/elee1766/jobshop/pkg/storage/store/sqlite"
)

// TestMemoryEditsMigrate: rows in the old memory_edits table are in the
// memory_edits collection after 00021 (#579), with their ids, and new
// edits continue the ids.
func TestMemoryEditsMigrate(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	path := testPath(t)
	s, err := sqlite.Open(ctx, path, sqlite.Options{})
	require.NoError(t, err)
	require.NoError(t, s.Close())

	db, err := openAt21(t, path)
	require.NoError(t, err)
	for _, q := range []string{
		"DROP INDEX ru_memory_edits_pr",
		"CREATE TABLE memory_edits (id INTEGER PRIMARY KEY, pr TEXT NOT NULL DEFAULT '', learnings TEXT NOT NULL DEFAULT '[]', pools TEXT NOT NULL DEFAULT '[]', merged_at INTEGER NOT NULL, baseline TEXT NOT NULL DEFAULT '{}', state TEXT NOT NULL DEFAULT 'pending', after TEXT NOT NULL DEFAULT '{}', reason TEXT NOT NULL DEFAULT '', evaluated_at INTEGER NOT NULL DEFAULT 0, created_at INTEGER NOT NULL, path TEXT NOT NULL DEFAULT '', before TEXT NOT NULL DEFAULT '', existed INTEGER NOT NULL DEFAULT 0)",
		"DELETE FROM goose_db_version WHERE version_id = 21",
		`INSERT INTO memory_edits VALUES (3, 'https://x/pull/1', '[7,8]', '["dev"]', 1000, '{"runs":4,"steps":9}', 'helped', '{"runs":5}', 'fewer retries', 5000, 900, '', '', 0)`,
		`INSERT INTO memory_edits VALUES (5, '', '[]', '[]', 2000, '{}', 'pending', '{}', '', 0, 2000, 'memory/global.md', 'old text', 1)`,
	} {
		_, err = db.ExecContext(ctx, q)
		require.NoError(t, err, q)
	}
	require.NoError(t, db.Close())

	s, err = sqlite.Open(ctx, path, sqlite.Options{})
	require.NoError(t, err)
	defer func() { _ = s.Close() }()

	e, err := memedit.Get(ctx, s, 3)
	require.NoError(t, err)
	assert.Equal(t, "https://x/pull/1", e.PR)
	assert.Equal(t, []int64{7, 8}, e.Learnings)
	assert.Equal(t, []string{"dev"}, e.Pools)
	assert.Equal(t, time.UnixMilli(1000).UTC(), e.MergedAt)
	assert.Equal(t, 4, e.Baseline.Runs)
	assert.Equal(t, 5, e.After.Runs)
	assert.Equal(t, domain.EditHelped, e.State)
	assert.Equal(t, "fewer retries", e.Reason)
	assert.Equal(t, time.UnixMilli(5000).UTC(), e.EvaluatedAt)
	assert.Equal(t, time.UnixMilli(900).UTC(), e.CreatedAt)

	w, err := memedit.Get(ctx, s, 5)
	require.NoError(t, err)
	assert.Equal(t, "memory/global.md", w.Path)
	assert.Equal(t, "old text", w.Before)
	assert.True(t, w.Existed)
	assert.Equal(t, domain.EditPending, w.State)

	// The PR stays unique and ids continue past the migrated ones.
	again, err := memedit.AddPR(ctx, s, domain.MemoryEdit{PR: "https://x/pull/1", MergedAt: time.UnixMilli(3000)})
	require.NoError(t, err)
	assert.Equal(t, int64(3), again.ID)
	n, err := memedit.AddWrite(ctx, s, domain.MemoryEdit{Path: "memory/global.md", MergedAt: time.UnixMilli(3000)})
	require.NoError(t, err)
	assert.Equal(t, int64(6), n.ID)
}
