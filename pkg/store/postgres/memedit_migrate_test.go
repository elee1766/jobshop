package postgres_test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/memedit"
	"github.com/elee1766/jobshop/pkg/store/postgres"
)

// TestMemoryEditsMigrate: 00006 (#579) copies memory_edits rows into
// the collection with their ids, Down copies them back, and Up again
// keeps them.
func TestMemoryEditsMigrate(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, name := openNamed(t, postgres.Options{})
	db, err := sql.Open("pgx", withDB(os.Getenv(dsnEnv), name))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	require.NoError(t, postgres.MigrateDownTo(ctx, db, 5))
	for _, q := range []string{
		`INSERT INTO memory_edits (id, pr, learnings, pools, merged_at, baseline, state, after, reason, evaluated_at, created_at)
		 VALUES (3, 'https://x/pull/1', '[7,8]', '["dev"]', 1000, '{"runs":4}', 'helped', '{"runs":5}', 'fewer retries', 5000, 900)`,
		`INSERT INTO memory_edits (id, learnings, pools, merged_at, created_at, path, before, existed)
		 VALUES (5, '[]', '[]', 2000, 2000, 'memory/global.md', 'old text', 1)`,
	} {
		_, err = db.ExecContext(ctx, q)
		require.NoError(t, err, q)
	}
	require.NoError(t, postgres.Migrate(ctx, db))
	check := func() {
		t.Helper()
		e, err := memedit.Get(ctx, s, 3)
		require.NoError(t, err)
		assert.Equal(t, "https://x/pull/1", e.PR)
		assert.Equal(t, []int64{7, 8}, e.Learnings)
		assert.Equal(t, 5, e.After.Runs)
		assert.Equal(t, domain.EditHelped, e.State)
		assert.Equal(t, time.UnixMilli(5000).UTC(), e.EvaluatedAt)
		w, err := memedit.Get(ctx, s, 5)
		require.NoError(t, err)
		assert.True(t, w.Existed)
		assert.Equal(t, "old text", w.Before)
	}
	check()

	require.NoError(t, postgres.MigrateDownTo(ctx, db, 5))
	var n int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM memory_edits").Scan(&n))
	assert.Equal(t, 2, n)
	require.NoError(t, postgres.Migrate(ctx, db))
	check()

	again, err := memedit.AddPR(ctx, s, domain.MemoryEdit{PR: "https://x/pull/1", MergedAt: time.UnixMilli(3000)})
	require.NoError(t, err)
	assert.Equal(t, int64(3), again.ID)
}
