-- name: InsertLearning :one
INSERT INTO learnings (kind, summary, fix, scope, ticket, session, source, status, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: ListLearnings :many
SELECT * FROM learnings
WHERE created_at >= sqlc.arg(since)
  AND (sqlc.narg(scope) IS NULL OR scope = sqlc.narg(scope))
  AND (sqlc.narg(kind) IS NULL OR kind = sqlc.narg(kind))
  AND (sqlc.narg(status) IS NULL OR status = sqlc.narg(status))
  AND (sqlc.narg(ticket) IS NULL OR ticket = sqlc.narg(ticket))
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(max_rows);

-- name: UpdateLearning :one
UPDATE learnings
SET status = sqlc.arg(status), pr = sqlc.arg(pr)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: CountLearnings :one
SELECT COUNT(*) FROM learnings
WHERE created_at >= sqlc.arg(since)
  AND (sqlc.narg(status) IS NULL OR status = sqlc.narg(status));

-- name: GetLearning :one
SELECT * FROM learnings WHERE id = ?;
