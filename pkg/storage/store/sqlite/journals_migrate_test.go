package sqlite_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/storage/journal"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/storage/store/sqlite"
)

// TestJournalsMigrate: journal rows in the shared recall index are in
// the journals collection after 00024 (#579), searchable, and Down puts
// them back.
func TestJournalsMigrate(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	path := testPath(t)
	s, err := sqlite.Open(ctx, path, sqlite.Options{})
	require.NoError(t, err)
	require.NoError(t, s.Close())

	db, err := sql.Open("sqlite", "file:"+path)
	require.NoError(t, err)
	require.NoError(t, sqlite.MigrateDownTo(ctx, db, 23))
	_, err = db.ExecContext(ctx, `INSERT INTO recall (body, kind, ref, ticket_id, project, at)
		VALUES ('nix vendorHash mismatch, paste got', 'journal', 'sess1', 'p-1', 'p', '1000')`)
	require.NoError(t, err)
	require.NoError(t, db.Close())

	check := func() {
		t.Helper()
		s, err := sqlite.Open(ctx, path, sqlite.Options{})
		require.NoError(t, err)
		defer func() { _ = s.Close() }()
		hits, err := journal.Recall(ctx, s, "vendorHash", "p", "p-1", 0)
		require.NoError(t, err)
		require.Len(t, hits, 1)
		assert.Equal(t, "sess1", hits[0].Ref)
		assert.Equal(t, time.UnixMilli(1000).UTC(), hits[0].At)
		idx, err := s.Recall(ctx, store.RecallQuery{Text: "vendorHash"})
		require.NoError(t, err)
		assert.Empty(t, idx, "journal rows left the shared recall index")
	}
	check()

	db, err = sql.Open("sqlite", "file:"+path)
	require.NoError(t, err)
	require.NoError(t, sqlite.MigrateDownTo(ctx, db, 23))
	var n int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM recall WHERE kind = 'journal' AND ref = 'sess1'").Scan(&n))
	assert.Equal(t, 1, n)
	require.NoError(t, db.Close())
	check()
}
