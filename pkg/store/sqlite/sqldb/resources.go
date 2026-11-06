package sqldb

import (
	"context"
)

const addTicketResource = `-- name: AddTicketResource :exec
INSERT INTO ticket_resources (ticket_id, rkey, rname) VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING
`

// AddTicketResourceParams are the arguments of AddTicketResource.
type AddTicketResourceParams struct {
	TicketID string
	Rkey     string
	Rname    string
}

// AddTicketResource runs the AddTicketResource query.
func (q *Queries) AddTicketResource(ctx context.Context, arg AddTicketResourceParams) error {
	_, err := q.db.ExecContext(ctx, addTicketResource, arg.TicketID, arg.Rkey, arg.Rname)
	return err
}

const clearTicketResources = `-- name: ClearTicketResources :exec
DELETE FROM ticket_resources WHERE ticket_id = $1
`

// ClearTicketResources runs the ClearTicketResources query. ClearTicketResources and AddTicketResource keep ticket_resources in
// step with fields.resources (#426).
func (q *Queries) ClearTicketResources(ctx context.Context, ticketID string) error {
	_, err := q.db.ExecContext(ctx, clearTicketResources, ticketID)
	return err
}

const deleteResourceCaps = `-- name: DeleteResourceCaps :exec
DELETE FROM resource_caps
`

// DeleteResourceCaps runs the DeleteResourceCaps query.
func (q *Queries) DeleteResourceCaps(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteResourceCaps)
	return err
}

const deleteResourceProviders = `-- name: DeleteResourceProviders :exec
DELETE FROM resource_providers
`

// DeleteResourceProviders runs the DeleteResourceProviders query.
func (q *Queries) DeleteResourceProviders(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteResourceProviders)
	return err
}

const deleteSlot = `-- name: DeleteSlot :exec
DELETE FROM resource_slots WHERE pool = $1 AND id = $2
`

// DeleteSlotParams are the arguments of DeleteSlot.
type DeleteSlotParams struct {
	Pool string
	ID   string
}

// DeleteSlot runs the DeleteSlot query.
func (q *Queries) DeleteSlot(ctx context.Context, arg DeleteSlotParams) error {
	_, err := q.db.ExecContext(ctx, deleteSlot, arg.Pool, arg.ID)
	return err
}

const getSlot = `-- name: GetSlot :one
SELECT pool, id, state, env, ticket, fence, removing, updated_at FROM resource_slots WHERE pool = $1 AND id = $2
`

// GetSlotParams are the arguments of GetSlot.
type GetSlotParams struct {
	Pool string
	ID   string
}

// GetSlot runs the GetSlot query.
func (q *Queries) GetSlot(ctx context.Context, arg GetSlotParams) (ResourceSlot, error) {
	row := q.db.QueryRowContext(ctx, getSlot, arg.Pool, arg.ID)
	var i ResourceSlot
	err := row.Scan(
		&i.Pool,
		&i.ID,
		&i.State,
		&i.Env,
		&i.Ticket,
		&i.Fence,
		&i.Removing,
		&i.UpdatedAt,
	)
	return i, err
}

const insertResourceCap = `-- name: InsertResourceCap :exec
INSERT INTO resource_caps (name, capacity) VALUES ($1, $2)
`

// InsertResourceCapParams are the arguments of InsertResourceCap.
type InsertResourceCapParams struct {
	Name     string
	Capacity int64
}

// InsertResourceCap runs the InsertResourceCap query.
func (q *Queries) InsertResourceCap(ctx context.Context, arg InsertResourceCapParams) error {
	_, err := q.db.ExecContext(ctx, insertResourceCap, arg.Name, arg.Capacity)
	return err
}

const insertResourceProvider = `-- name: InsertResourceProvider :exec
INSERT INTO resource_providers (name) VALUES ($1)
`

// InsertResourceProvider runs the InsertResourceProvider query.
func (q *Queries) InsertResourceProvider(ctx context.Context, name string) error {
	_, err := q.db.ExecContext(ctx, insertResourceProvider, name)
	return err
}

const insertSlot = `-- name: InsertSlot :exec
INSERT INTO resource_slots (pool, id, state, env, updated_at) VALUES ($1, $2, 'ready', $3, $4)
`

// InsertSlotParams are the arguments of InsertSlot.
type InsertSlotParams struct {
	Pool      string
	ID        string
	Env       string
	UpdatedAt int64
}

// InsertSlot runs the InsertSlot query.
func (q *Queries) InsertSlot(ctx context.Context, arg InsertSlotParams) error {
	_, err := q.db.ExecContext(ctx, insertSlot,
		arg.Pool,
		arg.ID,
		arg.Env,
		arg.UpdatedAt,
	)
	return err
}

const listSlots = `-- name: ListSlots :many
SELECT pool, id, state, env, ticket, fence, removing, updated_at FROM resource_slots ORDER BY pool, id
`

// ListSlots runs the ListSlots query.
func (q *Queries) ListSlots(ctx context.Context) ([]ResourceSlot, error) {
	rows, err := q.db.QueryContext(ctx, listSlots)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := []ResourceSlot{}
	for rows.Next() {
		var i ResourceSlot
		if err := rows.Scan(
			&i.Pool,
			&i.ID,
			&i.State,
			&i.Env,
			&i.Ticket,
			&i.Fence,
			&i.Removing,
			&i.UpdatedAt,
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

const markSlotsReleasing = `-- name: MarkSlotsReleasing :exec
UPDATE resource_slots SET state = 'releasing', updated_at = $1
WHERE state = 'held' AND NOT EXISTS (
  SELECT 1 FROM tickets t WHERE t.id = resource_slots.ticket
    AND t.status = 'in_progress' AND t.fence = resource_slots.fence)
`

// MarkSlotsReleasing runs the MarkSlotsReleasing query. MarkSlotsReleasing moves held slots whose claim ended (the ticket is
// not in progress, or was claimed again with a new fence) to releasing.
func (q *Queries) MarkSlotsReleasing(ctx context.Context, updatedAt int64) error {
	_, err := q.db.ExecContext(ctx, markSlotsReleasing, updatedAt)
	return err
}

const providerDemand = `-- name: ProviderDemand :one
SELECT count(*) AS waiting, CAST(coalesce(min(t.created_at), 0) AS BIGINT) AS oldest
FROM tickets t JOIN ticket_resources r ON r.ticket_id = t.id
WHERE t.status = 'open' AND r.rkey = CAST($1 AS TEXT)
`

// ProviderDemandRow is one row of ProviderDemand.
type ProviderDemandRow struct {
	Waiting int64
	Oldest  int64
}

// ProviderDemand runs the ProviderDemand query. ProviderDemand is the waiting side of a provider's demand (#407): open
// tickets naming it and the oldest one's created_at.
func (q *Queries) ProviderDemand(ctx context.Context, pool string) (ProviderDemandRow, error) {
	row := q.db.QueryRowContext(ctx, providerDemand, pool)
	var i ProviderDemandRow
	err := row.Scan(&i.Waiting, &i.Oldest)
	return i, err
}

const providerEmpty = `-- name: ProviderEmpty :one
SELECT CAST(CAST(EXISTS (
  SELECT 1 FROM ticket_resources r
  JOIN resource_providers p ON p.name = r.rkey
  WHERE r.ticket_id = $1 AND NOT EXISTS (
    SELECT 1 FROM resource_slots s WHERE s.pool = p.name AND s.state = 'ready')
) AS INT) AS BIGINT) AS is_empty
`

// ProviderEmpty runs the ProviderEmpty query. ProviderEmpty reports whether ticket id needs a provider resource with
// no ready slot (#407), for Claim by id. Same rule as ReadyTickets.
func (q *Queries) ProviderEmpty(ctx context.Context, id string) (int64, error) {
	row := q.db.QueryRowContext(ctx, providerEmpty, id)
	var isEmpty int64
	err := row.Scan(&isEmpty)
	return isEmpty, err
}

const resourceCaps = `-- name: ResourceCaps :many
SELECT name, capacity FROM resource_caps ORDER BY name
`

// ResourceCaps runs the ResourceCaps query.
func (q *Queries) ResourceCaps(ctx context.Context) ([]ResourceCap, error) {
	rows, err := q.db.QueryContext(ctx, resourceCaps)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := []ResourceCap{}
	for rows.Next() {
		var i ResourceCap
		if err := rows.Scan(&i.Name, &i.Capacity); err != nil {
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

const resourceFull = `-- name: ResourceFull :one
SELECT CAST(CAST(EXISTS (
  SELECT 1 FROM ticket_resources r
  JOIN resource_caps c ON c.name = r.rkey OR c.name = r.rname
  WHERE r.ticket_id = $1 AND c.capacity <= (
    SELECT count(*) FROM ticket_resources hr JOIN tickets h ON h.id = hr.ticket_id
    WHERE h.status = 'in_progress' AND h.id <> r.ticket_id AND hr.rkey = r.rkey)
) AS INT) AS BIGINT) AS is_full
`

// ResourceFull runs the ResourceFull query. ResourceFull reports whether ticket id names a resource key with no
// free slot (#400), for Claim by id. Same rule as ReadyTickets.
func (q *Queries) ResourceFull(ctx context.Context, id string) (int64, error) {
	row := q.db.QueryRowContext(ctx, resourceFull, id)
	var isFull int64
	err := row.Scan(&isFull)
	return isFull, err
}

const resourceUse = `-- name: ResourceUse :many
SELECT r.rkey, count(*) AS held
FROM ticket_resources r JOIN tickets t ON t.id = r.ticket_id
WHERE t.status = 'in_progress'
GROUP BY r.rkey
ORDER BY r.rkey
`

// ResourceUseRow is one row of ResourceUse.
type ResourceUseRow struct {
	Rkey string
	Held int64
}

// ResourceUse counts in-progress tickets per resource key (#400).
func (q *Queries) ResourceUse(ctx context.Context) ([]ResourceUseRow, error) {
	rows, err := q.db.QueryContext(ctx, resourceUse)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := []ResourceUseRow{}
	for rows.Next() {
		var i ResourceUseRow
		if err := rows.Scan(&i.Rkey, &i.Held); err != nil {
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

const resourceWaiters = `-- name: ResourceWaiters :many
SELECT t.id, r.rkey
FROM ticket_resources r JOIN tickets t ON t.id = r.ticket_id
WHERE t.status = 'open'
  AND t.defer_until <= $1
  AND t.f_waiting = 0
  AND NOT EXISTS (
    SELECT 1 FROM edges e JOIN tickets b ON b.id = e.to_id
    WHERE e.from_id = t.id AND e.kind IN ('blocks','waits_for')
      AND (b.status <> 'closed' OR (b.kind = 'step' AND b.resolution NOT IN ('done','skipped'))))
ORDER BY r.rkey, t.priority, t.created_at, t.id
`

// ResourceWaitersRow is one row of ResourceWaiters.
type ResourceWaitersRow struct {
	ID   string
	Rkey string
}

// ResourceWaiters runs the ResourceWaiters query. ResourceWaiters lists open, unblocked, non-deferred tickets per
// resource key they name, in claim order (#408). The caller keeps the
// keys that are full: those tickets wait on the resource alone.
func (q *Queries) ResourceWaiters(ctx context.Context, now int64) ([]ResourceWaitersRow, error) {
	rows, err := q.db.QueryContext(ctx, resourceWaiters, now)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := []ResourceWaitersRow{}
	for rows.Next() {
		var i ResourceWaitersRow
		if err := rows.Scan(&i.ID, &i.Rkey); err != nil {
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

const resourceWaiting = `-- name: ResourceWaiting :many
SELECT r.rkey, count(*) AS waiting
FROM ticket_resources r JOIN tickets t ON t.id = r.ticket_id
WHERE t.status = 'open'
GROUP BY r.rkey
ORDER BY r.rkey
`

// ResourceWaitingRow is one row of ResourceWaiting.
type ResourceWaitingRow struct {
	Rkey    string
	Waiting int64
}

// ResourceWaiting counts open tickets per resource key (#400).
func (q *Queries) ResourceWaiting(ctx context.Context) ([]ResourceWaitingRow, error) {
	rows, err := q.db.QueryContext(ctx, resourceWaiting)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := []ResourceWaitingRow{}
	for rows.Next() {
		var i ResourceWaitingRow
		if err := rows.Scan(&i.Rkey, &i.Waiting); err != nil {
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

const takeSlot = `-- name: TakeSlot :one
UPDATE resource_slots SET state = 'held', ticket = $1, fence = $2,
  updated_at = $3
WHERE resource_slots.pool = $4 AND resource_slots.id = (
  SELECT s.id FROM resource_slots s WHERE s.pool = $4 AND s.state = 'ready'
  ORDER BY s.updated_at, s.id LIMIT 1)
RETURNING pool, id, state, env, ticket, fence, removing, updated_at
`

// TakeSlotParams are the arguments of TakeSlot.
type TakeSlotParams struct {
	Ticket    string
	Fence     int64
	UpdatedAt int64
	Pool      string
}

// TakeSlot holds pool's longest-idle ready slot for a claim (#407).
func (q *Queries) TakeSlot(ctx context.Context, arg TakeSlotParams) (ResourceSlot, error) {
	row := q.db.QueryRowContext(ctx, takeSlot,
		arg.Ticket,
		arg.Fence,
		arg.UpdatedAt,
		arg.Pool,
	)
	var i ResourceSlot
	err := row.Scan(
		&i.Pool,
		&i.ID,
		&i.State,
		&i.Env,
		&i.Ticket,
		&i.Fence,
		&i.Removing,
		&i.UpdatedAt,
	)
	return i, err
}

const ticketProviderResources = `-- name: TicketProviderResources :many
SELECT r.rkey AS name
FROM ticket_resources r
JOIN resource_providers p ON p.name = r.rkey
WHERE r.ticket_id = $1
ORDER BY r.rkey
`

// TicketProviderResources lists the provider resources ticket id needs.
func (q *Queries) TicketProviderResources(ctx context.Context, id string) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, ticketProviderResources, id)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		items = append(items, name)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateSlot = `-- name: UpdateSlot :exec
UPDATE resource_slots SET
  state = $1, env = $2, ticket = $3,
  fence = $4, removing = $5, updated_at = $6
WHERE pool = $7 AND id = $8
`

// UpdateSlotParams are the arguments of UpdateSlot.
type UpdateSlotParams struct {
	State     string
	Env       string
	Ticket    string
	Fence     int64
	Removing  int64
	UpdatedAt int64
	Pool      string
	ID        string
}

// UpdateSlot runs the UpdateSlot query.
func (q *Queries) UpdateSlot(ctx context.Context, arg UpdateSlotParams) error {
	_, err := q.db.ExecContext(ctx, updateSlot,
		arg.State,
		arg.Env,
		arg.Ticket,
		arg.Fence,
		arg.Removing,
		arg.UpdatedAt,
		arg.Pool,
		arg.ID,
	)
	return err
}
