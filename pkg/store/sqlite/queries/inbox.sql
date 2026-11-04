-- name: InboxInsert :execrows
-- No conflict target: a clash on delivery_id or on natural_key (#335)
-- both mean the item is already stored.
INSERT INTO inbox (source, delivery_id, kind, repo_or_team, number, comment_id, actor, received_at, payload, updated_at, endpoint, natural_key)
VALUES (sqlc.arg(source), sqlc.arg(delivery_id), sqlc.arg(kind), sqlc.arg(repo_or_team), sqlc.arg(number),
        sqlc.arg(comment_id), sqlc.arg(actor), sqlc.arg(received_at), sqlc.arg(payload), sqlc.arg(now), sqlc.arg(endpoint),
        sqlc.arg(natural_key))
ON CONFLICT DO NOTHING;

-- name: InboxGetByNaturalKey :one
SELECT * FROM inbox WHERE natural_key = ? AND natural_key != '';

-- name: InboxGetByDelivery :one
SELECT * FROM inbox WHERE delivery_id = ?;

-- name: InboxGet :one
SELECT * FROM inbox WHERE id = ?;

-- now comes before the slice: sqlc numbers later parameters ?N, which
-- would point into an expanded list of two or more states.
-- name: InboxNextClaimable :one
SELECT * FROM inbox
WHERE (holder = '' OR lease_until <= sqlc.arg(now)) AND state IN (sqlc.slice(states))
ORDER BY received_at, id
LIMIT 1;

-- name: InboxTake :one
-- Takes the row only if it is still claimable, so two claimers racing on
-- the same row cannot both win.
UPDATE inbox SET holder = sqlc.arg(holder), lease_until = sqlc.arg(lease_until),
  attempts = attempts + 1, updated_at = sqlc.arg(now)
WHERE id = sqlc.arg(id) AND (holder = '' OR lease_until <= sqlc.arg(now))
RETURNING *;

-- name: InboxRenew :one
UPDATE inbox SET lease_until = sqlc.arg(lease_until), updated_at = sqlc.arg(now)
WHERE id = sqlc.arg(id) AND holder = sqlc.arg(holder) AND lease_until > sqlc.arg(now)
RETURNING *;

-- name: InboxSetState :one
-- Fenced by holder and a live lease. The lease is dropped so the next
-- stage can claim the row in its new state.
UPDATE inbox SET state = sqlc.arg(state),
  admitted = COALESCE(sqlc.narg(admitted), admitted),
  hook_results = COALESCE(sqlc.narg(hook_results), hook_results),
  holder = '', lease_until = 0, updated_at = sqlc.arg(now)
WHERE id = sqlc.arg(id) AND holder = sqlc.arg(holder) AND lease_until > sqlc.arg(now)
RETURNING *;

-- name: InboxList :many
SELECT * FROM inbox
WHERE (sqlc.arg(state) = '' OR state = sqlc.arg(state))
  AND (sqlc.arg(source) = '' OR source = sqlc.arg(source))
ORDER BY received_at DESC, id DESC
LIMIT CAST(sqlc.arg(lim) AS INTEGER);

-- name: InboxPrune :execrows
DELETE FROM inbox
WHERE state IN ('dispatched', 'denied', 'failed') AND received_at < sqlc.arg(before);

-- name: InboxRequeue :one
-- Moves an unleased item from one state to another (held -> pending
-- once its needs-human ticket is answered, #353). Guarded by the
-- current state so a racing stage cannot be overwritten.
UPDATE inbox SET state = sqlc.arg(to_state), updated_at = sqlc.arg(now)
WHERE id = sqlc.arg(id) AND state = sqlc.arg(from_state) AND (holder = '' OR lease_until <= sqlc.arg(now))
RETURNING *;

-- name: InboxStateCounts :many
SELECT state, COUNT(*) AS n FROM inbox GROUP BY state;

-- name: InboxCountOlder :one
-- Rows in state received before before: the admission backlog (#443).
SELECT COUNT(*) FROM inbox WHERE state = sqlc.arg(state) AND received_at < sqlc.arg(before);
