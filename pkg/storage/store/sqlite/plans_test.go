package sqlite_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/storage/store/sqlite"
)

// hookLike is a spec shaped like hook_invocations: a trigger field index.
var hookLike = store.CollectionSpec{
	Name:  "hook_like",
	Index: []store.Field{{Name: "trigger", Path: "trigger"}},
}

// explain returns the query plan of q on the database file at path,
// read the way the product reads it: a fresh connection to a database
// that only Open has prepared.
func explain(t *testing.T, db *sql.DB, q string, args ...any) string {
	t.Helper()
	rows, err := db.QueryContext(t.Context(), "EXPLAIN QUERY PLAN "+q, args...)
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

// planDB fills a collection through one store, then reopens the file with
// Open, which is all the statistics step the product runs, and returns
// a plain connection to it. fill gets the collection and the store.
func planDB(t *testing.T, spec store.CollectionSpec, fill func(col store.Collection)) (*sqlite.Store, *sql.DB) {
	t.Helper()
	ctx := t.Context()
	path := testPath(t)
	s, err := sqlite.Open(ctx, path, sqlite.Options{})
	require.NoError(t, err)
	col, err := s.Collection(ctx, spec)
	require.NoError(t, err)
	fill(col)
	require.NoError(t, s.Close())
	s, err = sqlite.Open(ctx, path, sqlite.Options{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=query_only(1)")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return s, db
}

// TestClaimPlan checks the Claim candidate queries, for a collection that
// is mostly terminal rows with none claimable, do not walk records_created
// (JOB-191: 200ms per kick on a 1 GB database).
func TestClaimPlan(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, db := planDB(t, hookLike, func(col store.Collection) {
		for i := range 1500 {
			st := "done"
			if i%7 == 0 {
				st = "failed"
			}
			_, _, err := col.Put(ctx, store.Record{Key: fmt.Sprint(i), State: st, Value: json.RawMessage(fmt.Sprintf(`{"trigger":"t%d"}`, i%20))})
			require.NoError(t, err)
		}
	})
	texts, args, err := sqlite.ClaimSQL(ctx, s, hookLike, store.RecordFilter{States: []string{"pending", "planned"}})
	require.NoError(t, err)
	require.Len(t, texts, 2, "one query per state")
	for i, q := range texts {
		p := explain(t, db, q, args[i]...)
		require.NotContains(t, p, "records_created", p)
		require.Contains(t, p, "records_state", p)
	}
}

// TestLastFirePlan checks the query LastFire runs, for a trigger with no
// accepted invocation, uses the trigger field index on a database that
// only Open prepared. Without statistics SQLite chose records_created and
// read the whole collection (JOB-191).
func TestLastFirePlan(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, db := planDB(t, hookLike, func(col store.Collection) {
		for i := range 1500 {
			_, _, err := col.Put(ctx, store.Record{Key: fmt.Sprint(i), State: "done", Value: json.RawMessage(fmt.Sprintf(`{"trigger":"t%d"}`, i%20))})
			require.NoError(t, err)
		}
	})
	q, args, err := sqlite.ListSQL(ctx, s, hookLike, store.RecordFilter{
		Where: map[string]string{"trigger": "never-fired"}, States: []string{"done", "planned", "failed"},
		Order: store.OrderCreatedDesc, Limit: 1,
	})
	require.NoError(t, err)
	p := explain(t, db, q, args...)
	require.Contains(t, p, "ri_hook_like_trigger", p)
}
