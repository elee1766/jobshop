package postgres_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/storage/store/postgres"
	"github.com/gfx-labs/jobshop/pkg/storage/store/storetest"
)

// TestClaimPlan checks the Claim candidate query, for a collection that is
// mostly terminal rows with none claimable, does not walk created_at
// order through every row (JOB-191).
func TestClaimPlan(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, db := openRaw(t)
	col, err := s.Collection(ctx, storetest.TestSpec)
	require.NoError(t, err)
	for i := range 3000 {
		st := "done"
		if i%7 == 0 {
			st = "failed"
		}
		_, _, err := col.Put(ctx, store.Record{Key: fmt.Sprint(i), State: st, Value: json.RawMessage(`{}`)})
		require.NoError(t, err)
	}
	_, err = db.ExecContext(ctx, "ANALYZE")
	require.NoError(t, err)

	texts, args, err := postgres.ClaimSQL(ctx, s, storetest.TestSpec, store.RecordFilter{States: []string{"pending", "planned"}})
	require.NoError(t, err)
	for i, q := range texts {
		rows, err := db.QueryContext(ctx, "EXPLAIN "+q, args[i]...)
		require.NoError(t, err)
		var out []string
		for rows.Next() {
			var line string
			require.NoError(t, rows.Scan(&line))
			out = append(out, line)
		}
		require.NoError(t, rows.Err())
		require.NoError(t, rows.Close())
		p := strings.Join(out, "\n")
		t.Log(p)
		require.NotContains(t, p, "records_test_items_created_at_idx", p)
		require.NotContains(t, p, "Seq Scan", p)
	}
}
