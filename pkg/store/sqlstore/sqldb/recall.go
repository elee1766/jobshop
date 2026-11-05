package sqldb

import (
	"context"
)

const recallSearch = `-- name: RecallSearch :many
SELECT kind, ref, ticket_id, project, CAST(at AS INTEGER) AS at_ms,
       snippet(recall, 0, '[', ']', '...', 16) AS snippet,
       CAST(bm25(recall) AS REAL) AS score
FROM recall
WHERE body MATCH ?1
  AND kind = 'comment'
  AND (?2 = '' OR project = ?2)
  AND (?3 = '' OR ticket_id = ?3)
ORDER BY score
LIMIT CAST(?4 AS INTEGER)
`

// RecallSearchParams are the arguments of RecallSearch.
type RecallSearchParams struct {
	Query    string
	Project  interface{}
	TicketID interface{}
	Lim      int64
}

// RecallSearchRow is one row of RecallSearch.
type RecallSearchRow struct {
	Kind     string
	Ref      string
	TicketID string
	Project  string
	AtMs     int64
	Snippet  string
	Score    float64
}

// RecallSearch searches comments only: journals and learnings are collections (#579).
func (q *Queries) RecallSearch(ctx context.Context, arg RecallSearchParams) ([]RecallSearchRow, error) {
	rows, err := q.db.QueryContext(ctx, recallSearch,
		arg.Query,
		arg.Project,
		arg.TicketID,
		arg.Lim,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := []RecallSearchRow{}
	for rows.Next() {
		var i RecallSearchRow
		if err := rows.Scan(
			&i.Kind,
			&i.Ref,
			&i.TicketID,
			&i.Project,
			&i.AtMs,
			&i.Snippet,
			&i.Score,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
