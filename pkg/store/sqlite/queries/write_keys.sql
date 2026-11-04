-- name: GetWriteKey :one
SELECT * FROM write_keys WHERE key = ?;

-- name: InsertWriteKey :execrows
INSERT INTO write_keys (key, op, actor, created_at, updated_at)
VALUES (sqlc.arg(key), sqlc.arg(op), sqlc.arg(actor), sqlc.arg(now), sqlc.arg(now))
ON CONFLICT (key) DO NOTHING;

-- name: FinishWriteKey :one
UPDATE write_keys SET result = sqlc.arg(result), done = 1, updated_at = sqlc.arg(now)
WHERE key = sqlc.arg(key) AND done = 0
RETURNING *;

-- name: PruneWriteKeys :execrows
DELETE FROM write_keys WHERE created_at < sqlc.arg(before);
