-- name: GetTicket :one
SELECT * FROM tickets WHERE id = ?;

-- name: InsertTicket :exec
INSERT INTO tickets (
  id, project, kind, title, body, status, resolution, priority, pool, holder,
  lease_until, fence, parent_id, defer_until, ephemeral, fields, meta,
  revision, created_by, created_at, updated_at, closed_at
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
  ?, ?, ?, ?, ?, ?, ?,
  ?, ?, ?, ?, ?
);

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
  meta        = sqlc.arg(meta),
  revision    = sqlc.arg(revision),
  updated_at  = sqlc.arg(updated_at),
  closed_at   = sqlc.arg(closed_at)
WHERE id = sqlc.arg(id) AND revision = sqlc.arg(expected_revision);

-- Heartbeat renews a lease without a revision bump or history row.
-- name: Heartbeat :execrows
UPDATE tickets SET lease_until = sqlc.arg(lease_until), updated_at = sqlc.arg(now)
WHERE id = sqlc.arg(id)
  AND holder = sqlc.arg(holder)
  AND fence = sqlc.arg(fence)
  AND status = 'in_progress';

-- name: ListTickets :many
SELECT * FROM tickets
WHERE (sqlc.narg(project) IS NULL OR project = sqlc.narg(project))
  AND (sqlc.narg(kind) IS NULL OR kind = sqlc.narg(kind))
  AND (sqlc.narg(status) IS NULL OR status = sqlc.narg(status))
  AND (sqlc.narg(pool) IS NULL OR pool = sqlc.narg(pool))
  AND (sqlc.narg(holder) IS NULL OR holder = sqlc.narg(holder))
  AND (sqlc.narg(parent_id) IS NULL OR parent_id = sqlc.narg(parent_id))
  AND (sqlc.narg(label) IS NULL OR EXISTS (
    SELECT 1 FROM labels l WHERE l.ticket_id = tickets.id AND l.label = sqlc.narg(label)))
ORDER BY priority, created_at, id
LIMIT sqlc.arg(max_rows);

-- name: ReadyTickets :many
SELECT t.* FROM tickets t
WHERE t.project = sqlc.arg(project)
  AND t.status = 'open'
  AND t.kind IN (SELECT value FROM json_each(sqlc.arg(kinds_json)))
  AND (sqlc.narg(pool) IS NULL OR t.pool = sqlc.narg(pool))
  AND t.defer_until <= sqlc.arg(now)
  AND NOT EXISTS (
    SELECT 1 FROM edges e JOIN tickets b ON b.id = e.to_id
    WHERE e.from_id = t.id AND e.kind IN ('blocks','waits_for') AND b.status <> 'closed')
ORDER BY t.priority, t.created_at, t.id
LIMIT sqlc.arg(max_rows);

-- Three tiers: my in-progress work, tickets pre-assigned to me, the pool queue.
-- name: NextClaimCandidate :one
SELECT t.* FROM tickets t
WHERE t.project = sqlc.arg(project)
  AND t.kind IN (SELECT value FROM json_each(sqlc.arg(kinds_json)))
  AND (
    (t.status = 'in_progress' AND t.holder = sqlc.arg(holder))
    OR (t.status = 'open' AND t.holder = sqlc.arg(holder))
    OR (t.status = 'open' AND t.holder = '' AND t.pool = sqlc.arg(pool))
  )
  AND t.defer_until <= sqlc.arg(now)
  AND NOT EXISTS (
    SELECT 1 FROM edges e JOIN tickets b ON b.id = e.to_id
    WHERE e.from_id = t.id AND e.kind IN ('blocks','waits_for') AND b.status <> 'closed')
ORDER BY
  CASE WHEN t.status = 'in_progress' THEN 0
       WHEN t.holder <> '' THEN 1 -- open with a holder can only be mine here
       ELSE 2 END,
  t.priority, t.created_at, t.id
LIMIT 1;

-- name: ExpiredLeases :many
SELECT id FROM tickets
WHERE status = 'in_progress' AND lease_until > 0 AND lease_until < sqlc.arg(now)
ORDER BY id;

-- HeldBy covers claimed and pre-assigned tickets, so a dead session strands neither.
-- name: HeldBy :many
SELECT id FROM tickets WHERE holder = ? AND status <> 'closed' ORDER BY id;
