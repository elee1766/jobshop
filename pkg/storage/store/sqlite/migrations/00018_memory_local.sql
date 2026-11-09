-- +goose Up
-- Local memory writes (#592): the curator writes files under
-- <root>/memory through the controller instead of opening a PR. A row is
-- one write: path (relative to the memory dir), the text it replaced
-- (before, existed 0 when the file was new), and pr empty. Old rows keep
-- their pr. pr stays unique when set. SQLite cannot change a column
-- constraint in place, so the table is rebuilt with its ids.
CREATE TABLE memory_edits_new (
  id           INTEGER PRIMARY KEY,
  pr           TEXT NOT NULL DEFAULT '',
  learnings    TEXT NOT NULL DEFAULT '[]',
  pools        TEXT NOT NULL DEFAULT '[]',
  merged_at    INTEGER NOT NULL,
  baseline     TEXT NOT NULL DEFAULT '{}',
  state        TEXT NOT NULL DEFAULT 'pending' CHECK (state IN ('pending','helped','not_helped')),
  after        TEXT NOT NULL DEFAULT '{}',
  reason       TEXT NOT NULL DEFAULT '',
  evaluated_at INTEGER NOT NULL DEFAULT 0,
  created_at   INTEGER NOT NULL,
  -- Last, as Postgres adds them, so SELECT * scans the same on both.
  path         TEXT NOT NULL DEFAULT '',
  before       TEXT NOT NULL DEFAULT '',
  existed      INTEGER NOT NULL DEFAULT 0
);
INSERT INTO memory_edits_new (id, pr, learnings, pools, merged_at, baseline, state, after, reason, evaluated_at, created_at)
SELECT id, pr, learnings, pools, merged_at, baseline, state, after, reason, evaluated_at, created_at FROM memory_edits;
DROP TABLE memory_edits;
ALTER TABLE memory_edits_new RENAME TO memory_edits;
CREATE INDEX memory_edits_state ON memory_edits(state, merged_at);
CREATE UNIQUE INDEX memory_edits_pr ON memory_edits(pr) WHERE pr != '';

-- +goose Down
DELETE FROM memory_edits WHERE pr = '';
CREATE TABLE memory_edits_old (
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
INSERT INTO memory_edits_old (id, pr, learnings, pools, merged_at, baseline, state, after, reason, evaluated_at, created_at)
SELECT id, pr, learnings, pools, merged_at, baseline, state, after, reason, evaluated_at, created_at FROM memory_edits;
DROP TABLE memory_edits;
ALTER TABLE memory_edits_old RENAME TO memory_edits;
CREATE INDEX memory_edits_state ON memory_edits(state, merged_at);
