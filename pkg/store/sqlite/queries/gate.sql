-- Curator gate reads (#274). Memory edits moved to the memory_edits
-- collection (#579, pkg/memedit).

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
