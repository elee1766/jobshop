package sqlite_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/storage/store/sqlite"
	"github.com/elee1766/jobshop/pkg/storage/store/storetest"
)

// TestCollectionIndexPlans checks declared-field and unique lookups use
// the collection's partial indexes, not a scan of records (#579).
func TestCollectionIndexPlans(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	path := testPath(t)
	s, err := sqlite.Open(ctx, path, sqlite.Options{})
	require.NoError(t, err)
	col, err := s.Collection(ctx, storetest.TestSpec)
	require.NoError(t, err)
	// Plans depend on statistics: seed rows spread over many values and
	// analyze, as a database in use would have.
	for i := range 300 {
		v := fmt.Sprintf(`{"source":"s%d","meta":{"repo":"r%d"},"natural_key":"n%d"}`, i%50, i%60, i)
		_, _, err := col.Put(ctx, store.Record{Key: fmt.Sprint(i), State: fmt.Sprint("st", i%5), Value: json.RawMessage(v)})
		require.NoError(t, err)
	}
	require.NoError(t, s.Close())
	// A second store builds the statement text; it does not touch the
	// database the plans run on.
	ps, err := sqlite.Open(ctx, ":memory:", sqlite.Options{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = ps.Close() })

	db, err := sql.Open("sqlite", "file:"+path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.ExecContext(ctx, "ANALYZE")
	require.NoError(t, err)

	plan := func(f store.RecordFilter) string {
		t.Helper()
		q, args, err := sqlite.ListSQL(ctx, ps, storetest.TestSpec, f)
		require.NoError(t, err)
		rows, err := db.QueryContext(ctx, "EXPLAIN QUERY PLAN "+q, args...)
		require.NoError(t, err)
		defer func() { _ = rows.Close() }()
		var out []string
		for rows.Next() {
			var id, parent, unused int
			var detail string
			require.NoError(t, rows.Scan(&id, &parent, &unused, &detail))
			out = append(out, detail)
		}
		require.NoError(t, rows.Err())
		return strings.Join(out, "; ")
	}
	name := storetest.TestSpec.Name
	require.Contains(t, plan(store.RecordFilter{Where: map[string]string{"source": "github"}}), "ri_"+name+"_source")
	require.Contains(t, plan(store.RecordFilter{Where: map[string]string{"repo": "a/b"}}), "ri_"+name+"_repo")
	require.Contains(t, plan(store.RecordFilter{Where: map[string]string{"natural_key": "k"}}), "ru_"+name+"_natural_key")
}

// TestCollectionCountPlan checks a state count over an age range (the
// inbox backlog) uses records_state_created, not a walk of
// records_created that reads every old row (JOB-61).
func TestCollectionCountPlan(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	path := testPath(t)
	s, err := sqlite.Open(ctx, path, sqlite.Options{})
	require.NoError(t, err)
	col, err := s.Collection(ctx, storetest.TestSpec)
	require.NoError(t, err)
	// Mostly terminal rows, a few pending, as a long-running inbox.
	for i := range 300 {
		st := "done"
		if i%30 == 0 {
			st = "pending"
		}
		_, _, err := col.Put(ctx, store.Record{Key: fmt.Sprint(i), State: st, Value: json.RawMessage(`{}`)})
		require.NoError(t, err)
	}
	require.NoError(t, s.Close())
	ps, err := sqlite.Open(ctx, ":memory:", sqlite.Options{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = ps.Close() })
	db, err := sql.Open("sqlite", "file:"+path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.ExecContext(ctx, "ANALYZE")
	require.NoError(t, err)

	q, args, err := sqlite.CountSQL(ctx, ps, storetest.TestSpec, store.RecordFilter{States: []string{"pending"}, Until: time.Now()})
	require.NoError(t, err)
	rows, err := db.QueryContext(ctx, "EXPLAIN QUERY PLAN "+q, args...)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var id, parent, unused int
		var detail string
		require.NoError(t, rows.Scan(&id, &parent, &unused, &detail))
		out = append(out, detail)
	}
	require.NoError(t, rows.Err())
	require.Contains(t, strings.Join(out, "; "), "records_state_created")
}

// TestCollectionRegisterDDL checks registration creates the named
// indexes and the FTS table.
func TestCollectionRegisterDDL(t *testing.T) {
	t.Parallel()
	ddl := strings.Join(sqlite.RegisterDDL(storetest.TestSpec), "\n")
	for _, want := range []string{
		"ri_test_items_source", "ri_test_items_repo", "ru_test_items_natural_key",
		"WHERE collection = 'test_items'", "fts_test_items USING fts5",
	} {
		require.Contains(t, ddl, want)
	}
}

// TestQuoteLit checks the one literal written into collection SQL.
func TestQuoteLit(t *testing.T) {
	t.Parallel()
	require.Equal(t, "'test_items'", sqlite.QuoteLit("test_items"))
	require.Equal(t, "'a''b'''", sqlite.QuoteLit("a'b'"))
}
