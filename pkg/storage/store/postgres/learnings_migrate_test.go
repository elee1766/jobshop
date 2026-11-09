package postgres_test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/learning"
	"github.com/elee1766/jobshop/pkg/storage/store/postgres"
)

// TestLearningsMigrate: 00007 (#579) copies learnings into the
// collection with their ids and search, Down copies them back with the
// recall rows, and Up again keeps them.
func TestLearningsMigrate(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, name := openNamed(t, postgres.Options{})
	db, err := sql.Open("pgx", withDB(os.Getenv(dsnEnv), name))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	require.NoError(t, postgres.MigrateDownTo(ctx, db, 6))
	_, err = db.ExecContext(ctx, `INSERT INTO learnings (id, kind, summary, fix, scope, source, status, created_at, pr, gate)
		VALUES (7, 'workaround', 'nix vendorHash mismatch', 'paste got', 'repo:a/b', 'agent', 'applied', 1000, 'https://x/pull/1', 'helped: x')`)
	require.NoError(t, err)
	require.NoError(t, postgres.Migrate(ctx, db))
	check := func() {
		t.Helper()
		l, err := learning.Get(ctx, s, 7)
		require.NoError(t, err)
		assert.Equal(t, "paste got", l.Fix)
		assert.Equal(t, domain.LearningApplied, l.Status)
		assert.Equal(t, "helped: x", l.Gate)
		assert.Equal(t, time.UnixMilli(1000).UTC(), l.CreatedAt)
		hits, err := learning.Recall(ctx, s, "vendorHash mismatch", "", "", 0)
		require.NoError(t, err)
		assert.Len(t, hits, 1)
	}
	check()

	require.NoError(t, postgres.MigrateDownTo(ctx, db, 6))
	var n int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM learnings").Scan(&n))
	assert.Equal(t, 1, n)
	require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM recall WHERE kind = 'learning'").Scan(&n))
	assert.Equal(t, 1, n)
	require.NoError(t, postgres.Migrate(ctx, db))
	check()
}
