-- name: GetPollCursor :one
SELECT * FROM poll_cursors WHERE poller = ? AND repo = ? AND stream = ?;

-- name: UpsertPollCursor :exec
INSERT INTO poll_cursors (poller, repo, stream, since, etag, updated_at)
VALUES (sqlc.arg(poller), sqlc.arg(repo), sqlc.arg(stream), sqlc.arg(since), sqlc.arg(etag), sqlc.arg(now))
ON CONFLICT (poller, repo, stream) DO UPDATE SET
  since = excluded.since, etag = excluded.etag, updated_at = excluded.updated_at;
