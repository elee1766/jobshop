-- +goose Up
-- Curator gate (#274): one row per merged curator PR. baseline holds the
-- metrics of the scoped pools before the merge; after, state and reason
-- are set by `jobshop learn evaluate`. learnings and pools are JSON arrays.
CREATE TABLE memory_edits (
  id           INTEGER PRIMARY KEY,
  pr           TEXT NOT NULL UNIQUE,
  learnings    TEXT NOT NULL DEFAULT '[]',
  pools        TEXT NOT NULL DEFAULT '[]',
  merged_at    INTEGER NOT NULL,
  baseline     TEXT NOT NULL DEFAULT '{}',
  state        TEXT NOT NULL DEFAULT 'pending' CHECK (state IN ('pending','helped','not_helped')),
  after        TEXT NOT NULL DEFAULT '{}',
  reason       TEXT NOT NULL DEFAULT '',
  evaluated_at INTEGER NOT NULL DEFAULT 0,
  created_at   INTEGER NOT NULL
);
CREATE INDEX memory_edits_state ON memory_edits(state, merged_at);
-- The gate's result on each cited learning: "helped: ..." or
-- "not_helped: ...", empty until evaluated.
ALTER TABLE learnings ADD COLUMN gate TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE learnings DROP COLUMN gate;
DROP TABLE memory_edits;
