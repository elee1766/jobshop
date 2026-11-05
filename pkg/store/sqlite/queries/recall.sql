-- name: RecallSearch :many
-- Comments only: journals and learnings are collections (#579).
SELECT kind, ref, ticket_id, project, CAST(at AS INTEGER) AS at_ms,
       snippet(recall, 0, '[', ']', '...', 16) AS snippet,
       CAST(bm25(recall) AS REAL) AS score
FROM recall
WHERE body MATCH sqlc.arg(query)
  AND kind = 'comment'
  AND (sqlc.arg(project) = '' OR project = sqlc.arg(project))
  AND (sqlc.arg(ticket_id) = '' OR ticket_id = sqlc.arg(ticket_id))
ORDER BY score
LIMIT CAST(sqlc.arg(lim) AS INTEGER);
