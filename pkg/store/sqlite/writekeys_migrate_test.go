package sqlite_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/store/sqlite"
	"github.com/elee1766/jobshop/pkg/writekey"
)

// TestWriteKeysCursorsMigrate: rows in the old write_keys and
// poll_cursors tables are in their collections after 00020 (#579).
func TestWriteKeysCursorsMigrate(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	path := testPath(t)
	s, err := sqlite.Open(ctx, path, sqlite.Options{})
	require.NoError(t, err)
	require.NoError(t, s.Close())

	since := time.Date(2026, 11, 4, 10, 0, 0, 0, time.UTC)
	db, err := sql.Open("sqlite", "file:"+path)
	require.NoError(t, err)
	for _, q := range []string{
		"CREATE TABLE write_keys (key TEXT PRIMARY KEY, op TEXT NOT NULL, result TEXT NOT NULL DEFAULT '', done INTEGER NOT NULL DEFAULT 0, actor TEXT NOT NULL DEFAULT '', created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL)",
		"CREATE TABLE poll_cursors (poller TEXT NOT NULL, repo TEXT NOT NULL, stream TEXT NOT NULL, since INTEGER NOT NULL DEFAULT 0, etag TEXT NOT NULL DEFAULT '', updated_at INTEGER NOT NULL, PRIMARY KEY (poller, repo, stream))",
		"DELETE FROM goose_db_version WHERE version_id = 20",
		"INSERT INTO write_keys VALUES ('js/done', 'gh.comment', 'url', 1, 'sess', 1000, 2000)",
		"INSERT INTO write_keys VALUES ('js/pending', 'gh.comment', '', 0, 'sess', 1000, 1000)",
	} {
		_, err = db.ExecContext(ctx, q)
		require.NoError(t, err, q)
	}
	_, err = db.ExecContext(ctx, "INSERT INTO poll_cursors VALUES ('gh', 'acme/app', 'issues', ?, 'W/\"e\"', 3000)", since.UnixMilli())
	require.NoError(t, err)
	require.NoError(t, db.Close())

	s, err = sqlite.Open(ctx, path, sqlite.Options{})
	require.NoError(t, err)
	defer func() { _ = s.Close() }()

	k, err := writekey.Claim(ctx, s, "js/done", "gh.comment", "x")
	require.NoError(t, err)
	assert.True(t, k.Done)
	assert.False(t, k.New)
	assert.Equal(t, "url", k.Result)
	assert.Equal(t, "sess", k.Actor)
	assert.Equal(t, time.UnixMilli(1000).UTC(), k.CreatedAt)
	k, err = writekey.Claim(ctx, s, "js/pending", "gh.comment", "x")
	require.NoError(t, err)
	assert.False(t, k.Done)
	assert.False(t, k.New)

	c, err := ingest.GetPollCursor(ctx, s, "gh", "acme/app", "issues")
	require.NoError(t, err)
	assert.True(t, c.Found)
	assert.Equal(t, since, c.Since)
	assert.Equal(t, `W/"e"`, c.ETag)
}
