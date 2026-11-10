package postgres

import (
	"context"

	"github.com/nullism/bqb"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// CountRuns counts the distinct runs that own a step created in r,
// in pools (any pool when empty). The memory-edit gate asks this every
// time it is polled, and listing the steps to count their parents read
// thousands of rows (JOB-191).
func (s *Store) CountRuns(ctx context.Context, pools []string, r store.TimeRange) (int, error) {
	w := bqb.New("kind = ?", string(domain.KindStep))
	w.And("parent_id <> ''")
	if !r.Since.IsZero() {
		w.And("created_at >= ?", ns(r.Since))
	}
	if !r.Until.IsZero() {
		w.And("created_at < ?", ns(r.Until))
	}
	if len(pools) > 0 {
		w.And("pool IN (?)", pools)
	}
	text, args, err := bqb.New("SELECT COUNT(DISTINCT parent_id) FROM tickets WHERE ?", w).ToPgsql()
	if err != nil {
		return 0, err
	}
	var n int
	err = s.rdb.QueryRowContext(ctx, text, args...).Scan(&n)
	return n, err
}

// Maintain does nothing: autovacuum keeps Postgres statistics.
func (s *Store) Maintain(context.Context) error { return nil }
