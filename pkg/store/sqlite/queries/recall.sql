-- name: RecallSearch :many
SELECT kind, ref, ticket_id, project, CAST(at AS INTEGER) AS at_ms,
       snippet(recall, 0, '[', ']', '...', 16) AS snippet,
       CAST(bm25(recall) AS REAL) AS score
FROM recall
WHERE body MATCH sqlc.arg(query)
  AND (sqlc.arg(project) = '' OR project = sqlc.arg(project) OR (kind = 'learning' AND project = ''))
  AND (sqlc.arg(kind) = '' OR kind = sqlc.arg(kind))
  AND (sqlc.arg(ticket_id) = '' OR ticket_id = sqlc.arg(ticket_id))
ORDER BY score
LIMIT CAST(sqlc.arg(lim) AS INTEGER);

-- name: RecallDelete :exec
DELETE FROM recall WHERE kind = sqlc.arg(kind) AND ref = sqlc.arg(ref);

-- name: RecallInsert :exec
INSERT INTO recall (body, kind, ref, ticket_id, project, at)
VALUES (sqlc.arg(body), sqlc.arg(kind), sqlc.arg(ref), sqlc.arg(ticket_id), sqlc.arg(project), sqlc.arg(at));
