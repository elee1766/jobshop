-- +goose Up
-- Ingest inbox (#321, #318): every webhook delivery or poller find is
-- stored here first, deduped by delivery_id, then claimed under a lease
-- and moved through admission and dispatch. At least once: a claim whose
-- lease expires is claimed again with attempts + 1.
CREATE TABLE inbox (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  source       TEXT NOT NULL CHECK (source IN ('github', 'linear')),
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
  updated_at   INTEGER NOT NULL
);
CREATE INDEX inbox_state_lease ON inbox (state, lease_until);
CREATE INDEX inbox_received ON inbox (received_at);

-- +goose Down
DROP TABLE inbox;
