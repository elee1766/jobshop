-- name: DeleteResourceCaps :exec
DELETE FROM resource_caps;

-- name: InsertResourceCap :exec
INSERT INTO resource_caps (name, capacity) VALUES (?, ?);

-- ResourceFull reports whether ticket id names a resource key with no
-- free slot (#400), for Claim by id. Same rule as ReadyTickets.
-- name: ResourceFull :one
SELECT CAST(EXISTS (
  SELECT 1 FROM ticket_resources r
  JOIN resource_caps c ON c.name = r.rkey OR c.name = r.rname
  WHERE r.ticket_id = sqlc.arg(id) AND c.capacity <= (
    SELECT count(*) FROM ticket_resources hr JOIN tickets h ON h.id = hr.ticket_id
    WHERE h.status = 'in_progress' AND h.id <> r.ticket_id AND hr.rkey = r.rkey)
) AS INTEGER) AS is_full;

-- name: ResourceCaps :many
SELECT name, capacity FROM resource_caps ORDER BY name;

-- ResourceUse counts in-progress tickets per resource key (#400).
-- name: ResourceUse :many
SELECT r.rkey, count(*) AS held
FROM ticket_resources r JOIN tickets t ON t.id = r.ticket_id
WHERE t.status = 'in_progress'
GROUP BY r.rkey
ORDER BY r.rkey;

-- ResourceWaiting counts open tickets per resource key (#400).
-- name: ResourceWaiting :many
SELECT r.rkey, count(*) AS waiting
FROM ticket_resources r JOIN tickets t ON t.id = r.ticket_id
WHERE t.status = 'open'
GROUP BY r.rkey
ORDER BY r.rkey;

-- ResourceWaiters lists open, unblocked, non-deferred tickets per
-- resource key they name, in claim order (#408). The caller keeps the
-- keys that are full: those tickets wait on the resource alone.
-- name: ResourceWaiters :many
SELECT t.id, r.rkey
FROM ticket_resources r JOIN tickets t ON t.id = r.ticket_id
WHERE t.status = 'open'
  AND t.defer_until <= sqlc.arg(now)
  AND t.f_waiting = 0
  AND NOT EXISTS (
    SELECT 1 FROM edges e JOIN tickets b ON b.id = e.to_id
    WHERE e.from_id = t.id AND e.kind IN ('blocks','waits_for')
      AND (b.status <> 'closed' OR (b.kind = 'step' AND b.resolution NOT IN ('done','skipped'))))
ORDER BY r.rkey, t.priority, t.created_at, t.id;

-- ClearTicketResources and AddTicketResource keep ticket_resources in
-- step with fields.resources (#426).
-- name: ClearTicketResources :exec
DELETE FROM ticket_resources WHERE ticket_id = ?;

-- name: AddTicketResource :exec
INSERT INTO ticket_resources (ticket_id, rkey, rname) VALUES (?, ?, ?)
ON CONFLICT DO NOTHING;

-- name: DeleteResourceProviders :exec
DELETE FROM resource_providers;

-- name: InsertResourceProvider :exec
INSERT INTO resource_providers (name) VALUES (?);

-- ProviderEmpty reports whether ticket id needs a provider resource with
-- no ready slot (#407), for Claim by id. Same rule as ReadyTickets.
-- name: ProviderEmpty :one
SELECT CAST(EXISTS (
  SELECT 1 FROM ticket_resources r
  JOIN resource_providers p ON p.name = r.rkey
  WHERE r.ticket_id = sqlc.arg(id) AND NOT EXISTS (
    SELECT 1 FROM resource_slots s WHERE s.pool = p.name AND s.state = 'ready')
) AS INTEGER) AS is_empty;

-- TicketProviderResources lists the provider resources ticket id needs.
-- name: TicketProviderResources :many
SELECT r.rkey AS name
FROM ticket_resources r
JOIN resource_providers p ON p.name = r.rkey
WHERE r.ticket_id = sqlc.arg(id)
ORDER BY r.rkey;

-- name: GetSlot :one
SELECT * FROM resource_slots WHERE pool = ? AND id = ?;

-- name: InsertSlot :exec
INSERT INTO resource_slots (pool, id, state, env, updated_at) VALUES (?, ?, 'ready', ?, ?);

-- name: UpdateSlot :exec
UPDATE resource_slots SET
  state = sqlc.arg(state), env = sqlc.arg(env), ticket = sqlc.arg(ticket),
  fence = sqlc.arg(fence), removing = sqlc.arg(removing), updated_at = sqlc.arg(updated_at)
WHERE pool = sqlc.arg(pool) AND id = sqlc.arg(id);

-- name: DeleteSlot :exec
DELETE FROM resource_slots WHERE pool = ? AND id = ?;

-- TakeSlot holds pool's longest-idle ready slot for a claim (#407).
-- name: TakeSlot :one
UPDATE resource_slots SET state = 'held', ticket = sqlc.arg(ticket), fence = sqlc.arg(fence),
  updated_at = sqlc.arg(updated_at)
WHERE resource_slots.pool = sqlc.arg(pool) AND resource_slots.id = (
  SELECT s.id FROM resource_slots s WHERE s.pool = sqlc.arg(pool) AND s.state = 'ready'
  ORDER BY s.updated_at, s.id LIMIT 1)
RETURNING *;

-- name: ListSlots :many
SELECT * FROM resource_slots ORDER BY pool, id;

-- MarkSlotsReleasing moves held slots whose claim ended (the ticket is
-- not in progress, or was claimed again with a new fence) to releasing.
-- name: MarkSlotsReleasing :exec
UPDATE resource_slots SET state = 'releasing', updated_at = sqlc.arg(updated_at)
WHERE state = 'held' AND NOT EXISTS (
  SELECT 1 FROM tickets t WHERE t.id = resource_slots.ticket
    AND t.status = 'in_progress' AND t.fence = resource_slots.fence);

-- ProviderDemand is the waiting side of a provider's demand (#407): open
-- tickets naming it and the oldest one's created_at.
-- name: ProviderDemand :one
SELECT count(*) AS waiting, CAST(coalesce(min(t.created_at), 0) AS INTEGER) AS oldest
FROM tickets t JOIN ticket_resources r ON r.ticket_id = t.id
WHERE t.status = 'open' AND r.rkey = CAST(sqlc.arg(pool) AS TEXT);
