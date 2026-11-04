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
  AND (sqlc.narg(created_by) IS NULL OR created_by = sqlc.narg(created_by))
  AND (sqlc.narg(label) IS NULL OR EXISTS (
    SELECT 1 FROM labels l WHERE l.ticket_id = tickets.id AND l.label = sqlc.narg(label)))
  AND (sqlc.narg(meta_path) IS NULL OR json_extract(meta, sqlc.narg(meta_path)) = sqlc.narg(meta_value))
ORDER BY priority, created_at, id
LIMIT sqlc.arg(max_rows);

-- ListTicketsNewest is ListTickets ordered newest first. sqlc's sqlite
-- engine does not bind parameters in ORDER BY, so it is a second query.
-- name: ListTicketsNewest :many
SELECT * FROM tickets
WHERE (sqlc.narg(project) IS NULL OR project = sqlc.narg(project))
  AND (sqlc.narg(kind) IS NULL OR kind = sqlc.narg(kind))
  AND (sqlc.narg(status) IS NULL OR status = sqlc.narg(status))
  AND (sqlc.narg(pool) IS NULL OR pool = sqlc.narg(pool))
  AND (sqlc.narg(holder) IS NULL OR holder = sqlc.narg(holder))
  AND (sqlc.narg(parent_id) IS NULL OR parent_id = sqlc.narg(parent_id))
  AND (sqlc.narg(created_by) IS NULL OR created_by = sqlc.narg(created_by))
  AND (sqlc.narg(label) IS NULL OR EXISTS (
    SELECT 1 FROM labels l WHERE l.ticket_id = tickets.id AND l.label = sqlc.narg(label)))
  AND (sqlc.narg(meta_path) IS NULL OR json_extract(meta, sqlc.narg(meta_path)) = sqlc.narg(meta_value))
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(max_rows);

-- name: ReadyTickets :many
SELECT t.* FROM tickets t
WHERE t.project = sqlc.arg(project)
  AND t.status = 'open'
  AND t.kind IN (SELECT value FROM json_each(sqlc.arg(kinds_json)))
  AND (sqlc.narg(pool) IS NULL OR t.pool = sqlc.narg(pool))
  AND t.defer_until <= sqlc.arg(now)
  -- fields.wait parks a ticket until the controller clears it (#381).
  AND json_extract(t.fields, '$.wait') IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM edges e JOIN tickets b ON b.id = e.to_id
    WHERE e.from_id = t.id AND e.kind IN ('blocks','waits_for')
      -- A step unblocks the next only when it closed done or skipped (#90, #385).
      AND (b.status <> 'closed' OR (b.kind = 'step' AND b.resolution NOT IN ('done','skipped'))))
  -- Every resource the ticket needs has a free slot (#400). Occupancy
  -- is in-progress tickets naming the same key; a key "name:..." takes
  -- the cap of "name"; a key with no cap is unlimited.
  AND NOT EXISTS (
    SELECT 1 FROM json_each(t.fields, '$.resources') r
    JOIN resource_caps c ON c.name = r.value
      OR (instr(r.value, ':') > 0 AND c.name = substr(r.value, 1, instr(r.value, ':') - 1))
    WHERE c.capacity <= (
      SELECT count(*) FROM tickets h, json_each(h.fields, '$.resources') hr
      WHERE h.status = 'in_progress' AND h.id <> t.id AND hr.value = r.value))
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
  -- fields.wait parks a ticket until the controller clears it (#381).
  AND json_extract(t.fields, '$.wait') IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM edges e JOIN tickets b ON b.id = e.to_id
    WHERE e.from_id = t.id AND e.kind IN ('blocks','waits_for')
      -- A step unblocks the next only when it closed done or skipped (#90, #385).
      AND (b.status <> 'closed' OR (b.kind = 'step' AND b.resolution NOT IN ('done','skipped'))))
  -- Resource slots, as in ReadyTickets (#400). My own in-progress
  -- ticket already holds its slots: h.id <> t.id leaves it out.
  AND NOT EXISTS (
    SELECT 1 FROM json_each(t.fields, '$.resources') r
    JOIN resource_caps c ON c.name = r.value
      OR (instr(r.value, ':') > 0 AND c.name = substr(r.value, 1, instr(r.value, ':') - 1))
    WHERE c.capacity <= (
      SELECT count(*) FROM tickets h, json_each(h.fields, '$.resources') hr
      WHERE h.status = 'in_progress' AND h.id <> t.id AND hr.value = r.value))
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

-- OrphanedTickets is open, unclaimed claimable work no pool will pick up:
-- no pool, or a pool name not in known_json (the loaded config).
-- name: OrphanedTickets :many
SELECT t.* FROM tickets t
WHERE (sqlc.narg(project) IS NULL OR t.project = sqlc.narg(project))
  AND t.status = 'open'
  AND t.holder = ''
  AND t.kind IN (SELECT value FROM json_each(sqlc.arg(kinds_json)))
  AND (t.pool = '' OR t.pool NOT IN (SELECT value FROM json_each(sqlc.arg(known_json))))
ORDER BY t.priority, t.created_at, t.id
LIMIT sqlc.arg(max_rows);

-- name: ListChildren :many
SELECT * FROM tickets
WHERE parent_id = sqlc.arg(parent_id)
ORDER BY CAST(json_extract(fields, '$.index') AS INTEGER), created_at, id;

-- FindOpenRunByKey uses tickets_run_key; at most one open run per
-- workflow and key (the controller serializes starts).
-- name: FindOpenRunByKey :one
SELECT * FROM tickets
WHERE kind = 'run' AND status <> 'closed'
  AND json_extract(fields, '$.key') = sqlc.arg(key)
  AND json_extract(fields, '$.workflow') = sqlc.arg(workflow)
ORDER BY created_at, id
LIMIT 1;

-- name: ListRuns :many
SELECT * FROM tickets
WHERE kind = 'run'
  AND (sqlc.narg(project) IS NULL OR project = sqlc.narg(project))
  AND (sqlc.narg(state) IS NULL OR json_extract(fields, '$.state') = sqlc.narg(state))
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(max_rows);

-- SetParent re-parents a ticket (batch adoption, #169). The caller bumps
-- the revision and writes history through mutate.
-- name: SetParent :exec
UPDATE tickets SET parent_id = sqlc.narg(parent_id) WHERE id = sqlc.arg(id);

-- FindOpenBatchByKey: at most one open batch per project and key (the
-- controller serializes starts).
-- name: FindOpenBatchByKey :one
SELECT * FROM tickets
WHERE kind = 'batch' AND status <> 'closed'
  AND project = sqlc.arg(project)
  AND json_extract(fields, '$.key') = sqlc.arg(key)
ORDER BY created_at, id
LIMIT 1;
