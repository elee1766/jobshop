package sqldb

import "context"

const bumpSeq = `-- name: BumpSeq :one
UPDATE id_seqs SET last = last + 1 WHERE prefix = $1 RETURNING last
`

// BumpSeq bumps the seq id counter for a prefix (JOB-57) and returns the
// new value: sql.ErrNoRows if the prefix has no counter yet.
func (q *Queries) BumpSeq(ctx context.Context, prefix string) (int64, error) {
	var n int64
	err := q.db.QueryRowContext(ctx, bumpSeq, prefix).Scan(&n)
	return n, err
}

const startSeq = `-- name: StartSeq :exec
INSERT INTO id_seqs (prefix, last) VALUES ($1, $2)
`

// StartSeq creates a prefix's seq counter at last.
func (q *Queries) StartSeq(ctx context.Context, prefix string, last int64) error {
	_, err := q.db.ExecContext(ctx, startSeq, prefix, last)
	return err
}

const ticketExists = `-- name: TicketExists :one
SELECT EXISTS (SELECT 1 FROM tickets WHERE id = $1)
`

// TicketExists reports whether a ticket has id.
func (q *Queries) TicketExists(ctx context.Context, id string) (bool, error) {
	var ok bool
	err := q.db.QueryRowContext(ctx, ticketExists, id).Scan(&ok)
	return ok, err
}

const ticketIDsWithPrefix = `-- name: TicketIDsWithPrefix :many
SELECT id FROM tickets WHERE substr(id, 1, length($1)) = $1
`

// TicketIDsWithPrefix lists ticket ids starting with p. It runs once per
// prefix, when its seq counter is created.
func (q *Queries) TicketIDsWithPrefix(ctx context.Context, p string) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, ticketIDsWithPrefix, p)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
