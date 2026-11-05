package postgres_test

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/postgres"
	"github.com/elee1766/jobshop/pkg/store/storetest"
)

// openRaw opens a fresh migrated database and returns the store and a
// plain connection to the same database.
func openRaw(t *testing.T) (*postgres.Store, *sql.DB) {
	t.Helper()
	dsn := os.Getenv(dsnEnv)
	if dsn == "" {
		t.Skip(dsnEnv + " is not set")
	}
	s, name := openNamed(t, postgres.Options{})
	db, err := sql.Open("pgx", withDB(dsn, name))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return s, db
}

// TestCollectionPartitions checks registration creates the collection's
// partition and that a write to an unregistered collection fails (#579).
func TestCollectionPartitions(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, db := openRaw(t)
	_, err := s.Collection(ctx, storetest.TestSpec)
	require.NoError(t, err)

	var parent string
	require.NoError(t, db.QueryRowContext(ctx, `SELECT inhparent::regclass::text FROM pg_inherits
		WHERE inhrelid = 'records_test_items'::regclass`).Scan(&parent))
	require.Equal(t, "records", parent)

	_, err = db.ExecContext(ctx, `INSERT INTO records (collection, key, created_at, updated_at)
		VALUES ('never_registered', 'k', 1, 1)`)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no partition")
}

// TestCollectionIndexPlans checks declared-field lookups use their
// indexes once the planner is told to avoid scans.
func TestCollectionIndexPlans(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, db := openRaw(t)
	_, err := s.Collection(ctx, storetest.TestSpec)
	require.NoError(t, err)
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	_, err = conn.ExecContext(ctx, "SET enable_seqscan = off")
	require.NoError(t, err)

	plan := func(f store.RecordFilter) string {
		t.Helper()
		q, args, err := postgres.ListSQL(ctx, s, storetest.TestSpec, f)
		require.NoError(t, err)
		rows, err := conn.QueryContext(ctx, "EXPLAIN "+q, args...)
		require.NoError(t, err)
		defer func() { _ = rows.Close() }()
		var out []string
		for rows.Next() {
			var line string
			require.NoError(t, rows.Scan(&line))
			out = append(out, line)
		}
		require.NoError(t, rows.Err())
		return strings.Join(out, "\n")
	}
	p := plan(store.RecordFilter{Where: map[string]string{"source": "github"}})
	require.Contains(t, p, "records_test_items", p)
	require.NotContains(t, p, "Seq Scan", p)
	p = plan(store.RecordFilter{Where: map[string]string{"natural_key": "k"}})
	require.Contains(t, p, fmt.Sprintf("ru_%s_natural_key", storetest.TestSpec.Name), p)
}

// TestQuoteLit checks the literal used in partition DDL.
func TestQuoteLit(t *testing.T) {
	t.Parallel()
	require.Equal(t, "'a''b'", postgres.QuoteLit("a'b"))
}
