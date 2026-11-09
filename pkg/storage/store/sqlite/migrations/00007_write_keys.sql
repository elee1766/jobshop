-- +goose Up
-- Idempotency keys for tracker writes (#280). A write first stores its
-- key with done = 0, then the result with done = 1. A retry that finds
-- a done key returns result; one that finds a pending key reads the
-- entity back by the id derived from the key instead of posting again.
CREATE TABLE write_keys (
  key        TEXT PRIMARY KEY,
  op         TEXT NOT NULL,
  result     TEXT NOT NULL DEFAULT '',
  done       INTEGER NOT NULL DEFAULT 0,
  actor      TEXT NOT NULL DEFAULT '',
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL
);

-- +goose Down
DROP TABLE write_keys;
