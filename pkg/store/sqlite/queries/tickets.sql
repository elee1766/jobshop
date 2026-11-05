-- name: GetTicket :one
SELECT * FROM tickets WHERE id = ?;

-- name: InsertTicket :exec
INSERT INTO tickets (
  id, project, kind, title, body, status, resolution, priority, pool, holder,
  lease_until, fence, parent_id, defer_until, ephemeral, fields, meta,
  revision, created_by, created_at, updated_at, closed_at,
  f_key, f_to, f_state, f_workflow, f_index, f_waiting
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
  ?, ?, ?, ?, ?, ?, ?,
  ?, ?, ?, ?, ?,
  ?, ?, ?, ?, ?, ?
);

-- ClearTicketMeta and AddTicketMeta keep ticket_meta in step with meta.
-- name: ClearTicketMeta :exec
DELETE FROM ticket_meta WHERE ticket_id = ?;

-- name: AddTicketMeta :exec
INSERT INTO ticket_meta (ticket_id, pack, mkey, value) VALUES (?, ?, ?, ?);

-- WriteTicket stores every mutable column, guarded by the previous revision.
-- name: WriteTicket :execrows
UPDATE tickets SET
  title       = sqlc.arg(title),
  body        = sqlc.arg(body),
  status      = sqlc.arg(status),
  resolution  = sqlc.arg(resolution),
  priority    = sqlc.arg(priority),
  pool        = sqlc.arg(pool),
  holder      = sqlc.arg(holder),
  lease_until = sqlc.arg(lease_until),
  fence       = sqlc.arg(fence),
  defer_until = sqlc.arg(defer_until),
  fields      = sqlc.arg(fields),
  f_key       = sqlc.narg(f_key),
  f_to        = sqlc.narg(f_to),
  f_state     = sqlc.narg(f_state),
  f_workflow  = sqlc.narg(f_workflow),
  f_index     = sqlc.narg(f_index),
  f_waiting   = sqlc.arg(f_waiting),
  meta        = sqlc.arg(meta),
  revision    = sqlc.arg(revision),
  updated_at  = sqlc.arg(updated_at),
  closed_at   = sqlc.arg(closed_at)
WHERE id = sqlc.arg(id) AND revision = sqlc.arg(expected_revision);

-- ClaimWriteTicket is WriteTicket for a claim (#406): the not-blocked
-- check runs in the same statement as the write, so a blocker that
-- reopens between the caller's read and this write fails the claim
-- instead of leaving a blocked ticket claimed. Blockers are other rows,
-- which the revision check alone does not cover.
-- name: ClaimWriteTicket :execrows
UPDATE tickets SET
  status      = sqlc.arg(status),
  holder      = sqlc.arg(holder),
  lease_until = sqlc.arg(lease_until),
  fence       = sqlc.arg(fence),
  revision    = sqlc.arg(revision),
  updated_at  = sqlc.arg(updated_at)
WHERE tickets.id = sqlc.arg(id) AND tickets.revision = sqlc.arg(expected_revision)
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
      AND NOT EXISTS (SELECT 1 FROM resource_slots s WHERE s.pool = p.name AND s.state = 'ready')));

-- ExpireWriteTicket reopens a ticket only if its lease is still expired
-- at write time (#406). Heartbeat renews a lease without a revision
-- bump, so the revision check alone would let a sweep reopen a ticket
-- whose holder just renewed.
-- name: ExpireWriteTicket :execrows
UPDATE tickets SET
  status      = sqlc.arg(status),
  resolution  = sqlc.arg(resolution),
  holder      = sqlc.arg(holder),
  lease_until = sqlc.arg(lease_until),
  fields      = sqlc.arg(fields),
  revision    = sqlc.arg(revision),
  updated_at  = sqlc.arg(updated_at),
  closed_at   = sqlc.arg(closed_at)
WHERE id = sqlc.arg(id) AND revision = sqlc.arg(expected_revision)
  AND status = 'in_progress' AND lease_until > 0 AND lease_until < sqlc.arg(now);

-- Heartbeat renews a lease without a revision bump or history row.
-- name: Heartbeat :execrows
UPDATE tickets SET lease_until = sqlc.arg(lease_until), updated_at = sqlc.arg(now)
WHERE id = sqlc.arg(id)
  AND holder = sqlc.arg(holder)
  AND fence = sqlc.arg(fence)
  AND status = 'in_progress';

-- name: ExpiredLeases :many
SELECT id FROM tickets
WHERE status = 'in_progress' AND lease_until > 0 AND lease_until < sqlc.arg(now)
ORDER BY id;

-- HeldBy covers claimed and pre-assigned tickets, so a dead session strands neither.
-- name: HeldBy :many
SELECT id FROM tickets WHERE holder = ? AND status <> 'closed' ORDER BY id;

-- SetParent re-parents a ticket (batch adoption, #169). The caller bumps
-- the revision and writes history through mutate.
-- name: SetParent :exec
UPDATE tickets SET parent_id = sqlc.narg(parent_id) WHERE id = sqlc.arg(id);
