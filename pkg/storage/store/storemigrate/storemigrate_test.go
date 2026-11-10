package storemigrate_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver "pgx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/storage/store/postgres"
	"github.com/gfx-labs/jobshop/pkg/storage/store/sqlite"
	"github.com/gfx-labs/jobshop/pkg/storage/store/storemigrate"
	"github.com/gfx-labs/jobshop/pkg/storage/store/storetest"
)

// dsnEnv names a Postgres server the test may create a database on, as
// in pkg/storage/store/postgres. Unset, the copy test skips.
const dsnEnv = "JOBSHOP_TEST_POSTGRES"

// TestTablesComplete: every SQLite table is copied, so a new table
// cannot be left behind silently. FTS5 tables (recall's shadow tables,
// a collection's fts_<name>) are indexes Postgres builds itself.
func TestTablesComplete(t *testing.T) {
	ctx := t.Context()
	path := filepath.Join(t.TempDir(), "jobshop.db")
	s, err := sqlite.Open(ctx, path, sqlite.Options{})
	require.NoError(t, err)
	rows, err := s.DB().QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%' AND name NOT LIKE 'recall_%' AND name NOT LIKE 'fts_%' AND name != 'goose_db_version'")
	require.NoError(t, err)
	var got []string
	for rows.Next() {
		var n string
		require.NoError(t, rows.Scan(&n))
		got = append(got, n)
	}
	require.NoError(t, rows.Err())
	require.NoError(t, rows.Close())
	require.NoError(t, s.Close())
	want := slices.Clone(storemigrate.Tables)
	slices.Sort(want)
	slices.Sort(got)
	assert.Equal(t, want, got)
}

func withDB(dsn, name string) string {
	u, err := url.Parse(dsn)
	if err != nil || u.Scheme == "" {
		return dsn + " dbname=" + name
	}
	u.Path = "/" + name
	return u.String()
}

func TestCopy(t *testing.T) {
	dsn := os.Getenv(dsnEnv)
	if dsn == "" {
		t.Skip(dsnEnv + " is not set")
	}
	ctx := t.Context()
	clock := &storetest.Clock{T: time.Date(2026, 11, 5, 10, 0, 0, 0, time.UTC)}

	path := filepath.Join(t.TempDir(), "jobshop.db")
	src, err := sqlite.Open(ctx, path, sqlite.Options{Now: clock.Now})
	require.NoError(t, err)
	filled := storetest.Fill(t, src, clock)

	admin, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	name := fmt.Sprintf("jobshop_migrate_%d", os.Getpid())
	_, err = admin.ExecContext(ctx, "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)")
	require.NoError(t, err)
	_, err = admin.ExecContext(ctx, "CREATE DATABASE "+name)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = admin.ExecContext(context.WithoutCancel(ctx), "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)")
		_ = admin.Close()
	})
	dst, err := postgres.Open(ctx, withDB(dsn, name), postgres.Options{Now: clock.Now})
	require.NoError(t, err)

	res, err := storemigrate.Copy(ctx, src.DB(), dst.DB(), storemigrate.Options{})
	require.NoError(t, err)
	assert.Equal(t, int64(4), res["tickets"])
	assert.Equal(t, int64(1), res["recall"], "the comment; journals are records")

	// A second copy refuses a filled target, and --force replaces it.
	_, err = storemigrate.Copy(ctx, src.DB(), dst.DB(), storemigrate.Options{})
	require.ErrorIs(t, err, storemigrate.ErrNotEmpty)
	_, err = storemigrate.Copy(ctx, src.DB(), dst.DB(), storemigrate.Options{Force: true})
	require.NoError(t, err)
	require.NoError(t, src.Close())
	require.NoError(t, dst.Close())

	// Reopen as the controller would, so collection registration and
	// prepared statements start fresh.
	dst, err = postgres.Open(ctx, withDB(dsn, name), postgres.Options{Now: clock.Now})
	require.NoError(t, err)
	defer func() { _ = dst.Close() }()
	storetest.CheckFilled(t, dst, filled)
}
