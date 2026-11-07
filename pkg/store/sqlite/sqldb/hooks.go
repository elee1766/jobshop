package sqldb

import "context"

const countChildren = `-- name: CountChildren :one
SELECT count(*) FROM tickets WHERE parent_id = $1
`

// CountChildren guards a complete child read set inside a graph transaction.
func (q *Queries) CountChildren(ctx context.Context, parent string) (int64, error) {
	var n int64
	err := q.db.QueryRowContext(ctx, countChildren, parent).Scan(&n)
	return n, err
}
