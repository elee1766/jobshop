-- name: InsertMemoryEdit :one
INSERT INTO memory_edits (pr, learnings, pools, merged_at, baseline, created_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetMemoryEditByPR :one
SELECT * FROM memory_edits WHERE pr = ?;

-- name: ListMemoryEdits :many
SELECT * FROM memory_edits
WHERE (sqlc.narg(state) IS NULL OR state = sqlc.narg(state))
ORDER BY merged_at DESC, id DESC
LIMIT CAST(sqlc.arg(max_rows) AS INTEGER);

-- name: SetMemoryEditResult :one
UPDATE memory_edits
SET state = sqlc.arg(state), after = sqlc.arg(after), reason = sqlc.arg(reason), evaluated_at = sqlc.arg(evaluated_at)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SetLearningGate :exec
UPDATE learnings SET gate = sqlc.arg(gate) WHERE id = sqlc.arg(id);

-- name: ListStepsCreated :many
SELECT * FROM tickets
WHERE kind = 'step' AND created_at >= sqlc.arg(since) AND created_at < sqlc.arg(until)
ORDER BY created_at, id;

-- name: ListLearningsBetween :many
SELECT * FROM learnings
WHERE created_at >= sqlc.arg(since) AND created_at < sqlc.arg(until)
ORDER BY created_at, id;

-- name: ListLearningsByPR :many
SELECT * FROM learnings WHERE pr = ? ORDER BY id;
