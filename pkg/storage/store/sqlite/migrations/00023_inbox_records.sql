-- +goose Up
-- The inbox moves onto the inbox store collection (#579, pkg/ingest
-- inbox.go). The key is the old id as text, the state and lease are the
-- record's, created_at is received_at (claims, prune and the backlog
-- count order and filter by it), and updated_at is kept. The id counter
-- in inbox_ids starts past the highest id (no counter on an empty inbox,
-- so a fresh store stays empty). The indexes are what
-- registration creates.
CREATE INDEX IF NOT EXISTS ri_inbox_source ON records (CAST(json_extract(value, '$.source') AS TEXT), created_at, key) WHERE collection = 'inbox';
CREATE UNIQUE INDEX IF NOT EXISTS ru_inbox_delivery_id ON records (CAST(json_extract(value, '$.delivery_id') AS TEXT)) WHERE collection = 'inbox' AND CAST(json_extract(value, '$.delivery_id') AS TEXT) != '';
CREATE UNIQUE INDEX IF NOT EXISTS ru_inbox_natural_key ON records (CAST(json_extract(value, '$.natural_key') AS TEXT)) WHERE collection = 'inbox' AND CAST(json_extract(value, '$.natural_key') AS TEXT) != '';

INSERT INTO records (collection, key, state, value, revision, holder, lease_until, attempts, created_at, updated_at)
SELECT 'inbox', CAST(id AS TEXT), state,
  json_object('id', id, 'source', source, 'delivery_id', delivery_id, 'endpoint', endpoint,
    'natural_key', natural_key, 'kind', kind, 'repo_or_team', repo_or_team, 'number', number,
    'comment_id', comment_id, 'actor', actor, 'payload', json(payload),
    'admitted', CASE WHEN admitted IS NULL THEN NULL ELSE json(admitted) END,
    'hook_results', json(hook_results)),
  1, holder, lease_until, attempts, received_at, updated_at
FROM inbox;
INSERT INTO records (collection, key, state, value, created_at, updated_at)
SELECT 'inbox_ids', 'next', '', json_object('next', (SELECT MAX(id) FROM inbox) + 1),
  CAST(unixepoch('subsec') * 1000 AS INTEGER), CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE EXISTS (SELECT 1 FROM inbox);
DROP TABLE inbox;

-- +goose Down
CREATE TABLE inbox (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  source       TEXT NOT NULL CHECK (source != ''),
  delivery_id  TEXT NOT NULL UNIQUE,
  kind         TEXT NOT NULL DEFAULT '',
  repo_or_team TEXT NOT NULL DEFAULT '',
  number       INTEGER NOT NULL DEFAULT 0,
  comment_id   TEXT NOT NULL DEFAULT '',
  actor        TEXT NOT NULL DEFAULT '',
  received_at  INTEGER NOT NULL,
  payload      TEXT NOT NULL DEFAULT '{}',
  admitted     TEXT,
  state        TEXT NOT NULL DEFAULT 'pending'
    CHECK (state IN ('pending', 'admitted', 'denied', 'held', 'dispatched', 'failed')),
  hook_results TEXT NOT NULL DEFAULT '[]',
  holder       TEXT NOT NULL DEFAULT '',
  lease_until  INTEGER NOT NULL DEFAULT 0,
  attempts     INTEGER NOT NULL DEFAULT 0,
  updated_at   INTEGER NOT NULL,
  endpoint     TEXT NOT NULL DEFAULT '',
  natural_key  TEXT NOT NULL DEFAULT ''
);
CREATE INDEX inbox_state_lease ON inbox (state, lease_until);
CREATE INDEX inbox_received ON inbox (received_at);
CREATE UNIQUE INDEX inbox_natural_key ON inbox (natural_key) WHERE natural_key != '';
INSERT INTO inbox (id, source, delivery_id, kind, repo_or_team, number, comment_id, actor,
  received_at, payload, admitted, state, hook_results, holder, lease_until, attempts, updated_at,
  endpoint, natural_key)
SELECT CAST(key AS INTEGER), json_extract(value, '$.source'), json_extract(value, '$.delivery_id'),
  json_extract(value, '$.kind'), json_extract(value, '$.repo_or_team'), json_extract(value, '$.number'),
  json_extract(value, '$.comment_id'), json_extract(value, '$.actor'), created_at,
  json_extract(value, '$.payload'),
  CASE WHEN json_type(value, '$.admitted') IN ('object', 'array', 'true', 'false', 'integer', 'real', 'text')
    THEN json_extract(value, '$.admitted') END,
  state, json_extract(value, '$.hook_results'), holder, lease_until, attempts, updated_at,
  json_extract(value, '$.endpoint'), json_extract(value, '$.natural_key')
FROM records WHERE collection = 'inbox';
DELETE FROM records WHERE collection IN ('inbox', 'inbox_ids');
DROP INDEX IF EXISTS ri_inbox_source;
DROP INDEX IF EXISTS ru_inbox_delivery_id;
DROP INDEX IF EXISTS ru_inbox_natural_key;
