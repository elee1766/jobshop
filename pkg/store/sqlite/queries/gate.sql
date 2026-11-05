-- Curator gate reads (#274). Memory edits and learnings moved to store
-- collections (#579, pkg/memedit, pkg/learning).

-- name: ListStepsCreated :many
SELECT * FROM tickets
WHERE kind = 'step' AND created_at >= sqlc.arg(since) AND created_at < sqlc.arg(until)
ORDER BY created_at, id;
