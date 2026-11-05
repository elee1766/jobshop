-- name: InsertEdge :exec
INSERT INTO edges (from_id, to_id, kind, created_at) VALUES (?, ?, ?, ?);

-- name: DeleteEdge :execrows
DELETE FROM edges WHERE from_id = ? AND to_id = ? AND kind = ?;

-- name: ListEdges :many
SELECT * FROM edges
WHERE from_id = sqlc.arg(id) OR to_id = sqlc.arg(id)
ORDER BY created_at, from_id, to_id, kind;

-- BlockingPathExists reports whether start reaches target through blocking
-- edges. Adding (A blocked on B) would close a cycle if B already reaches A.
-- name: BlockingPathExists :one
WITH RECURSIVE reach(id) AS (
  SELECT CAST(sqlc.arg(start) AS TEXT)
  UNION
  SELECT e.to_id FROM edges e JOIN reach r ON e.from_id = r.id
  WHERE e.kind IN ('blocks','waits_for')
)
SELECT EXISTS (SELECT 1 FROM reach WHERE reach.id = CAST(sqlc.arg(target) AS TEXT));

-- name: InsertLabel :execrows
INSERT INTO labels (ticket_id, label) VALUES (?, ?)
ON CONFLICT DO NOTHING;

-- name: DeleteLabel :execrows
DELETE FROM labels WHERE ticket_id = ? AND label = ?;

-- name: LabelsFor :many
SELECT ticket_id, label FROM labels
WHERE ticket_id IN (sqlc.slice(ids))
ORDER BY ticket_id, label;

-- name: InsertHistory :exec
INSERT INTO history (ticket_id, at, actor, op, revision, diff) VALUES (?, ?, ?, ?, ?, ?);

-- name: InsertComment :one
INSERT INTO comments (ticket_id, author, body, created_at) VALUES (?, ?, ?, ?)
RETURNING *;

