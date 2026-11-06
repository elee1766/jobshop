package pgdb

import (
	"context"
	"database/sql"
)

const addTicketMeta = `-- name: AddTicketMeta :exec
INSERT INTO ticket_meta (ticket_id, pack, mkey, value) VALUES ($1, $2, $3, $4)
`

// AddTicketMetaParams are the arguments of AddTicketMeta.
type AddTicketMetaParams struct {
	TicketID string
	Pack     string
	Mkey     string
	Value    string
}

// AddTicketMeta runs the AddTicketMeta query.
func (q *Queries) AddTicketMeta(ctx context.Context, arg AddTicketMetaParams) error {
	_, err := q.db.ExecContext(ctx, addTicketMeta,
		arg.TicketID,
		arg.Pack,
		arg.Mkey,
		arg.Value,
	)
	return err
}

const claimWriteTicket = `-- name: ClaimWriteTicket :execrows
UPDATE tickets SET
  status      = $1,
  holder      = $2,
  lease_until = $3,
  fence       = $4,
  revision    = $5,
  updated_at  = $6
WHERE tickets.id = $7 AND tickets.revision = $8
  AND tickets.status IN ('open', 'in_progress')
  AND NOT EXISTS (
    SELECT 1 FROM edges e JOIN tickets b ON b.id = e.to_id
    WHERE e.from_id = tickets.id AND e.kind IN ('blocks','waits_for')
      AND (b.status <> 'closed' OR (b.kind = 'step' AND b.resolution NOT IN ('done','skipped'))))
  -- An open ticket also needs a free slot for each resource (#400), so
  -- two claimers cannot both take the last slot. An in-progress resume
  -- already holds its slots.
  AND (tickets.status = 'in_progress' OR NOT EXISTS (
    SELECT 1 FROM ticket_resources r
    JOIN resource_caps c ON c.name = r.rkey OR c.name = r.rname
    WHERE r.ticket_id = tickets.id AND c.capacity <= (
      SELECT count(*) FROM ticket_resources hr JOIN tickets h ON h.id = hr.ticket_id
      WHERE h.status = 'in_progress' AND h.id <> tickets.id AND hr.rkey = r.rkey)))
  -- And a ready slot of each provider resource (#407).
  AND (tickets.status = 'in_progress' OR NOT EXISTS (
    SELECT 1 FROM ticket_resources r
    JOIN resource_providers p ON p.name = r.rkey
    WHERE r.ticket_id = tickets.id
      AND NOT EXISTS (SELECT 1 FROM resource_slots s WHERE s.pool = p.name AND s.state = 'ready')))
`

// ClaimWriteTicketParams are the arguments of ClaimWriteTicket.
type ClaimWriteTicketParams struct {
	Status           string
	Holder           string
	LeaseUntil       int64
	Fence            int64
	Revision         int64
	UpdatedAt        int64
	ID               string
	ExpectedRevision int64
}

// ClaimWriteTicket runs the ClaimWriteTicket query. ClaimWriteTicket is WriteTicket for a claim (#406): the not-blocked
// check runs in the same statement as the write, so a blocker that
// reopens between the caller's read and this write fails the claim
// instead of leaving a blocked ticket claimed. Blockers are other rows,
// which the revision check alone does not cover.
func (q *Queries) ClaimWriteTicket(ctx context.Context, arg ClaimWriteTicketParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, claimWriteTicket,
		arg.Status,
		arg.Holder,
		arg.LeaseUntil,
		arg.Fence,
		arg.Revision,
		arg.UpdatedAt,
		arg.ID,
		arg.ExpectedRevision,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const clearTicketMeta = `-- name: ClearTicketMeta :exec
DELETE FROM ticket_meta WHERE ticket_id = $1
`

// ClearTicketMeta and AddTicketMeta keep ticket_meta in step with meta.
func (q *Queries) ClearTicketMeta(ctx context.Context, ticketID string) error {
	_, err := q.db.ExecContext(ctx, clearTicketMeta, ticketID)
	return err
}

const expireWriteTicket = `-- name: ExpireWriteTicket :execrows
UPDATE tickets SET
  status      = $1,
  resolution  = $2,
  holder      = $3,
  lease_until = $4,
  fields      = $5,
  revision    = $6,
  updated_at  = $7,
  closed_at   = $8
WHERE id = $9 AND revision = $10
  AND status = 'in_progress' AND lease_until > 0 AND lease_until < $11
`

// ExpireWriteTicketParams are the arguments of ExpireWriteTicket.
type ExpireWriteTicketParams struct {
	Status           string
	Resolution       string
	Holder           string
	LeaseUntil       int64
	Fields           string
	Revision         int64
	UpdatedAt        int64
	ClosedAt         int64
	ID               string
	ExpectedRevision int64
	Now              int64
}

// ExpireWriteTicket runs the ExpireWriteTicket query. ExpireWriteTicket reopens a ticket only if its lease is still expired
// at write time (#406). Heartbeat renews a lease without a revision
// bump, so the revision check alone would let a sweep reopen a ticket
// whose holder just renewed.
func (q *Queries) ExpireWriteTicket(ctx context.Context, arg ExpireWriteTicketParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, expireWriteTicket,
		arg.Status,
		arg.Resolution,
		arg.Holder,
		arg.LeaseUntil,
		arg.Fields,
		arg.Revision,
		arg.UpdatedAt,
		arg.ClosedAt,
		arg.ID,
		arg.ExpectedRevision,
		arg.Now,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const expiredLeases = `-- name: ExpiredLeases :many
SELECT id FROM tickets
WHERE status = 'in_progress' AND lease_until > 0 AND lease_until < $1
ORDER BY id
`

// ExpiredLeases runs the ExpiredLeases query.
func (q *Queries) ExpiredLeases(ctx context.Context, now int64) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, expiredLeases, now)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTicket = `-- name: GetTicket :one
SELECT id, project, kind, title, body, status, resolution, priority, pool, holder, lease_until, fence, parent_id, defer_until, ephemeral, fields, meta, revision, created_by, created_at, updated_at, closed_at, f_key, f_to, f_state, f_workflow, f_index, f_waiting FROM tickets WHERE id = $1
`

// GetTicket runs the GetTicket query.
func (q *Queries) GetTicket(ctx context.Context, id string) (Ticket, error) {
	row := q.db.QueryRowContext(ctx, getTicket, id)
	var i Ticket
	err := row.Scan(
		&i.ID,
		&i.Project,
		&i.Kind,
		&i.Title,
		&i.Body,
		&i.Status,
		&i.Resolution,
		&i.Priority,
		&i.Pool,
		&i.Holder,
		&i.LeaseUntil,
		&i.Fence,
		&i.ParentID,
		&i.DeferUntil,
		&i.Ephemeral,
		&i.Fields,
		&i.Meta,
		&i.Revision,
		&i.CreatedBy,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ClosedAt,
		&i.FKey,
		&i.FTo,
		&i.FState,
		&i.FWorkflow,
		&i.FIndex,
		&i.FWaiting,
	)
	return i, err
}

const heartbeat = `-- name: Heartbeat :execrows
UPDATE tickets SET lease_until = $1, updated_at = $2
WHERE id = $3
  AND holder = $4
  AND fence = $5
  AND status = 'in_progress'
`

// HeartbeatParams are the arguments of Heartbeat.
type HeartbeatParams struct {
	LeaseUntil int64
	Now        int64
	ID         string
	Holder     string
	Fence      int64
}

// Heartbeat renews a lease without a revision bump or history row.
func (q *Queries) Heartbeat(ctx context.Context, arg HeartbeatParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, heartbeat,
		arg.LeaseUntil,
		arg.Now,
		arg.ID,
		arg.Holder,
		arg.Fence,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const heldBy = `-- name: HeldBy :many
SELECT id FROM tickets WHERE holder = $1 AND status <> 'closed' ORDER BY id
`

// HeldBy covers claimed and pre-assigned tickets, so a dead session strands neither.
func (q *Queries) HeldBy(ctx context.Context, holder string) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, heldBy, holder)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertTicket = `-- name: InsertTicket :exec
INSERT INTO tickets (
  id, project, kind, title, body, status, resolution, priority, pool, holder,
  lease_until, fence, parent_id, defer_until, ephemeral, fields, meta,
  revision, created_by, created_at, updated_at, closed_at,
  f_key, f_to, f_state, f_workflow, f_index, f_waiting
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
  $11, $12, $13, $14, $15, $16, $17,
  $18, $19, $20, $21, $22,
  $23, $24, $25, $26, $27, $28
)
`

// InsertTicketParams are the arguments of InsertTicket.
type InsertTicketParams struct {
	ID         string
	Project    string
	Kind       string
	Title      string
	Body       string
	Status     string
	Resolution string
	Priority   int64
	Pool       string
	Holder     string
	LeaseUntil int64
	Fence      int64
	ParentID   sql.NullString
	DeferUntil int64
	Ephemeral  int64
	Fields     string
	Meta       string
	Revision   int64
	CreatedBy  string
	CreatedAt  int64
	UpdatedAt  int64
	ClosedAt   int64
	FKey       sql.NullString
	FTo        sql.NullString
	FState     sql.NullString
	FWorkflow  sql.NullString
	FIndex     sql.NullInt64
	FWaiting   int64
}

// InsertTicket runs the InsertTicket query.
func (q *Queries) InsertTicket(ctx context.Context, arg InsertTicketParams) error {
	_, err := q.db.ExecContext(ctx, insertTicket,
		arg.ID,
		arg.Project,
		arg.Kind,
		arg.Title,
		arg.Body,
		arg.Status,
		arg.Resolution,
		arg.Priority,
		arg.Pool,
		arg.Holder,
		arg.LeaseUntil,
		arg.Fence,
		arg.ParentID,
		arg.DeferUntil,
		arg.Ephemeral,
		arg.Fields,
		arg.Meta,
		arg.Revision,
		arg.CreatedBy,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.ClosedAt,
		arg.FKey,
		arg.FTo,
		arg.FState,
		arg.FWorkflow,
		arg.FIndex,
		arg.FWaiting,
	)
	return err
}

const setParent = `-- name: SetParent :exec
UPDATE tickets SET parent_id = $1 WHERE id = $2
`

// SetParentParams are the arguments of SetParent.
type SetParentParams struct {
	ParentID sql.NullString
	ID       string
}

// SetParent runs the SetParent query. SetParent re-parents a ticket (batch adoption, #169). The caller bumps
// the revision and writes history through mutate.
func (q *Queries) SetParent(ctx context.Context, arg SetParentParams) error {
	_, err := q.db.ExecContext(ctx, setParent, arg.ParentID, arg.ID)
	return err
}

const writeTicket = `-- name: WriteTicket :execrows
UPDATE tickets SET
  title       = $1,
  body        = $2,
  status      = $3,
  resolution  = $4,
  priority    = $5,
  pool        = $6,
  holder      = $7,
  lease_until = $8,
  fence       = $9,
  defer_until = $10,
  fields      = $11,
  f_key       = $12,
  f_to        = $13,
  f_state     = $14,
  f_workflow  = $15,
  f_index     = $16,
  f_waiting   = $17,
  meta        = $18,
  revision    = $19,
  updated_at  = $20,
  closed_at   = $21
WHERE id = $22 AND revision = $23
`

// WriteTicketParams are the arguments of WriteTicket.
type WriteTicketParams struct {
	Title            string
	Body             string
	Status           string
	Resolution       string
	Priority         int64
	Pool             string
	Holder           string
	LeaseUntil       int64
	Fence            int64
	DeferUntil       int64
	Fields           string
	FKey             sql.NullString
	FTo              sql.NullString
	FState           sql.NullString
	FWorkflow        sql.NullString
	FIndex           sql.NullInt64
	FWaiting         int64
	Meta             string
	Revision         int64
	UpdatedAt        int64
	ClosedAt         int64
	ID               string
	ExpectedRevision int64
}

// WriteTicket stores every mutable column, guarded by the previous revision.
func (q *Queries) WriteTicket(ctx context.Context, arg WriteTicketParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, writeTicket,
		arg.Title,
		arg.Body,
		arg.Status,
		arg.Resolution,
		arg.Priority,
		arg.Pool,
		arg.Holder,
		arg.LeaseUntil,
		arg.Fence,
		arg.DeferUntil,
		arg.Fields,
		arg.FKey,
		arg.FTo,
		arg.FState,
		arg.FWorkflow,
		arg.FIndex,
		arg.FWaiting,
		arg.Meta,
		arg.Revision,
		arg.UpdatedAt,
		arg.ClosedAt,
		arg.ID,
		arg.ExpectedRevision,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
