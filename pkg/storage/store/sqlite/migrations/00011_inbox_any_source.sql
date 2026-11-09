-- +goose Up
-- Generic ingest endpoints (#399): an inbox item's source is any endpoint
-- source name, not only github or linear. SQLite cannot drop a CHECK in
-- place, so the table is rebuilt; nothing references it.
CREATE TABLE inbox_new (
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
INSERT INTO inbox_new (id, source, delivery_id, kind, repo_or_team, number, comment_id, actor,
  received_at, payload, admitted, state, hook_results, holder, lease_until, attempts, updated_at,
  endpoint, natural_key)
SELECT id, source, delivery_id, kind, repo_or_team, number, comment_id, actor,
  received_at, payload, admitted, state, hook_results, holder, lease_until, attempts, updated_at,
  endpoint, natural_key FROM inbox;
DROP TABLE inbox;
ALTER TABLE inbox_new RENAME TO inbox;
CREATE INDEX inbox_state_lease ON inbox (state, lease_until);
CREATE INDEX inbox_received ON inbox (received_at);
CREATE UNIQUE INDEX inbox_natural_key ON inbox (natural_key) WHERE natural_key != '';

-- +goose Down
DELETE FROM inbox WHERE source NOT IN ('github', 'linear');
