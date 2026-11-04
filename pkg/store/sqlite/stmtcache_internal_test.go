package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestStmtCacheInTx: a text first run inside a transaction on the only
// connection runs unprepared (preparing would wait for that connection)
// and is prepared by the next call outside it.
func TestStmtCacheInTx(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "db"))
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	c := newStmtCache(db)
	_, err = cachedDB{DB: db, c: c}.ExecContext(ctx, "CREATE TABLE t (v INTEGER)")
	require.NoError(t, err)

	const ins = "INSERT INTO t (v) VALUES (?)"
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	_, err = cachedTx{Tx: tx, c: c}.ExecContext(ctx, ins, 1)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.True(t, c.pending[ins], "queued, not prepared, in the transaction")

	var n int
	require.NoError(t, cachedDB{DB: db, c: c}.QueryRowContext(ctx, "SELECT count(*) FROM t").Scan(&n))
	require.Equal(t, 1, n)
	require.Contains(t, c.m, ins, "prepared by the next call outside")
	require.Empty(t, c.pending)

	tx, err = db.BeginTx(ctx, nil)
	require.NoError(t, err)
	_, err = cachedTx{Tx: tx, c: c}.ExecContext(ctx, ins, 2)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.NoError(t, cachedDB{DB: db, c: c}.QueryRowContext(ctx, "SELECT count(*) FROM t").Scan(&n))
	require.Equal(t, 2, n)
}

// TestSeedMatchesMigrate: a database copied from the template has the
// schema and goose versions that running the migrations gives.
func TestSeedMatchesMigrate(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	schema := func(db *sql.DB) []string {
		rows, err := db.QueryContext(ctx, "SELECT type || ' ' || name || ' ' || coalesce(sql, '') FROM sqlite_schema ORDER BY type, name")
		require.NoError(t, err)
		defer func() { _ = rows.Close() }()
		var out []string
		for rows.Next() {
			var s string
			require.NoError(t, rows.Scan(&s))
			out = append(out, s)
		}
		require.NoError(t, rows.Err())
		var vs []int64
		vr, err := db.QueryContext(ctx, "SELECT version_id FROM goose_db_version ORDER BY id")
		require.NoError(t, err)
		defer func() { _ = vr.Close() }()
		for vr.Next() {
			var v int64
			require.NoError(t, vr.Scan(&v))
			vs = append(vs, v)
		}
		return append(out, fmt.Sprint(vs))
	}
	dir := t.TempDir()
	seeded := filepath.Join(dir, "seeded")
	seed(ctx, seeded)
	require.FileExists(t, seeded)
	a, err := sql.Open("sqlite", "file:"+seeded)
	require.NoError(t, err)
	t.Cleanup(func() { _ = a.Close() })
	b, err := sql.Open("sqlite", "file:"+filepath.Join(dir, "migrated"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = b.Close() })
	require.NoError(t, migrate(ctx, b))
	require.Equal(t, schema(b), schema(a))
	require.NoFileExists(t, seeded+".seed")

	// An existing file is never replaced.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "x"), []byte("keep"), 0o600))
	seed(ctx, filepath.Join(dir, "x"))
	got, err := os.ReadFile(filepath.Join(dir, "x"))
	require.NoError(t, err)
	require.Equal(t, "keep", string(got))
}
