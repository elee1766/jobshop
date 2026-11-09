package sqldb

import (
	"context"
	"strconv"
	"strings"
)

const blockingPathExists = `-- name: BlockingPathExists :one
WITH RECURSIVE reach(id) AS (
  SELECT CAST($2 AS TEXT)
  UNION
  SELECT e.to_id FROM edges e JOIN reach r ON e.from_id = r.id
  WHERE e.kind IN ('blocks','waits_for')
)
SELECT EXISTS (SELECT 1 FROM reach WHERE reach.id = CAST($1 AS TEXT))
`

// BlockingPathExistsParams are the arguments of BlockingPathExists.
type BlockingPathExistsParams struct {
	Target string
	Start  string
}

// BlockingPathExists runs the BlockingPathExists query. BlockingPathExists reports whether start reaches target through blocking
// edges. Adding (A blocked on B) would close a cycle if B already reaches A.
func (q *Queries) BlockingPathExists(ctx context.Context, arg BlockingPathExistsParams) (bool, error) {
	row := q.db.QueryRowContext(ctx, blockingPathExists, arg.Target, arg.Start)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const deleteEdge = `-- name: DeleteEdge :execrows
DELETE FROM edges WHERE from_id = $1 AND to_id = $2 AND kind = $3
`

// DeleteEdgeParams are the arguments of DeleteEdge.
type DeleteEdgeParams struct {
	FromID string
	ToID   string
	Kind   string
}

// DeleteEdge runs the DeleteEdge query.
func (q *Queries) DeleteEdge(ctx context.Context, arg DeleteEdgeParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteEdge, arg.FromID, arg.ToID, arg.Kind)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deleteLabel = `-- name: DeleteLabel :execrows
DELETE FROM labels WHERE ticket_id = $1 AND label = $2
`

// DeleteLabelParams are the arguments of DeleteLabel.
type DeleteLabelParams struct {
	TicketID string
	Label    string
}

// DeleteLabel runs the DeleteLabel query.
func (q *Queries) DeleteLabel(ctx context.Context, arg DeleteLabelParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteLabel, arg.TicketID, arg.Label)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const insertComment = `-- name: InsertComment :one
INSERT INTO comments (ticket_id, author, body, created_at) VALUES ($1, $2, $3, $4)
RETURNING id, ticket_id, author, body, created_at, edited_at, deleted
`

// InsertCommentParams are the arguments of InsertComment.
type InsertCommentParams struct {
	TicketID  string
	Author    string
	Body      string
	CreatedAt int64
}

// InsertComment runs the InsertComment query.
func (q *Queries) InsertComment(ctx context.Context, arg InsertCommentParams) (Comment, error) {
	row := q.db.QueryRowContext(ctx, insertComment,
		arg.TicketID,
		arg.Author,
		arg.Body,
		arg.CreatedAt,
	)
	var i Comment
	err := row.Scan(
		&i.ID,
		&i.TicketID,
		&i.Author,
		&i.Body,
		&i.CreatedAt,
		&i.EditedAt,
		&i.Deleted,
	)
	return i, err
}

const insertEdge = `-- name: InsertEdge :exec
INSERT INTO edges (from_id, to_id, kind, created_at) VALUES ($1, $2, $3, $4)
`

// InsertEdgeParams are the arguments of InsertEdge.
type InsertEdgeParams struct {
	FromID    string
	ToID      string
	Kind      string
	CreatedAt int64
}

// InsertEdge runs the InsertEdge query.
func (q *Queries) InsertEdge(ctx context.Context, arg InsertEdgeParams) error {
	_, err := q.db.ExecContext(ctx, insertEdge,
		arg.FromID,
		arg.ToID,
		arg.Kind,
		arg.CreatedAt,
	)
	return err
}

const insertHistory = `-- name: InsertHistory :exec
INSERT INTO history (ticket_id, at, actor, op, revision, diff) VALUES ($1, $2, $3, $4, $5, $6)
`

// InsertHistoryParams are the arguments of InsertHistory.
type InsertHistoryParams struct {
	TicketID string
	At       int64
	Actor    string
	Op       string
	Revision int64
	Diff     string
}

// InsertHistory runs the InsertHistory query.
func (q *Queries) InsertHistory(ctx context.Context, arg InsertHistoryParams) error {
	_, err := q.db.ExecContext(ctx, insertHistory,
		arg.TicketID,
		arg.At,
		arg.Actor,
		arg.Op,
		arg.Revision,
		arg.Diff,
	)
	return err
}

const insertLabel = `-- name: InsertLabel :execrows
INSERT INTO labels (ticket_id, label) VALUES ($1, $2)
ON CONFLICT DO NOTHING
`

// InsertLabelParams are the arguments of InsertLabel.
type InsertLabelParams struct {
	TicketID string
	Label    string
}

// InsertLabel runs the InsertLabel query.
func (q *Queries) InsertLabel(ctx context.Context, arg InsertLabelParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, insertLabel, arg.TicketID, arg.Label)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const labelsFor = `-- name: LabelsFor :many
SELECT ticket_id, label FROM labels
WHERE ticket_id IN (/*SLICE:ids*/)
ORDER BY ticket_id, label
`

// LabelsFor runs the LabelsFor query.
func (q *Queries) LabelsFor(ctx context.Context, ids []string) ([]Label, error) {
	query := labelsFor
	var queryParams []interface{}
	if len(ids) > 0 {
		ph := make([]string, len(ids))
		for i, v := range ids {
			queryParams = append(queryParams, v)
			ph[i] = "$" + strconv.Itoa(i+1)
		}
		query = strings.Replace(query, "/*SLICE:ids*/", strings.Join(ph, ","), 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/", "NULL", 1)
	}
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := []Label{}
	for rows.Next() {
		var i Label
		if err := rows.Scan(&i.TicketID, &i.Label); err != nil {
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

const listEdges = `-- name: ListEdges :many
SELECT from_id, to_id, kind, created_at FROM edges
WHERE from_id = $1 OR to_id = $1
ORDER BY created_at, from_id, to_id, kind
`

// ListEdges runs the ListEdges query.
func (q *Queries) ListEdges(ctx context.Context, id string) ([]Edge, error) {
	rows, err := q.db.QueryContext(ctx, listEdges, id)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := []Edge{}
	for rows.Next() {
		var i Edge
		if err := rows.Scan(
			&i.FromID,
			&i.ToID,
			&i.Kind,
			&i.CreatedAt,
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
