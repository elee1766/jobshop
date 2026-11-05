-- +goose Up
-- memory_edits moves onto the memory_edits store collection (#579,
-- pkg/memedit). The key is the old id as text; pr is a unique field, as
-- the partial index on the table was. The index is what registration
-- creates.
CREATE UNIQUE INDEX IF NOT EXISTS ru_memory_edits_pr ON records (CAST(json_extract(value, '$.pr') AS TEXT))
  WHERE collection = 'memory_edits' AND CAST(json_extract(value, '$.pr') AS TEXT) != '';
INSERT INTO records (collection, key, state, value, created_at, updated_at)
SELECT 'memory_edits', CAST(id AS TEXT), state,
  json_object('id', id, 'pr', pr, 'path', path, 'before', before,
    'existed', json(CASE WHEN existed != 0 THEN 'true' ELSE 'false' END),
    'learnings', json(learnings), 'pools', json(pools), 'merged_at', merged_at,
    'baseline', json(baseline), 'after', json(after), 'reason', reason, 'evaluated_at', evaluated_at),
  created_at, CASE WHEN evaluated_at > created_at THEN evaluated_at ELSE created_at END
FROM memory_edits;
DROP TABLE memory_edits;

-- +goose Down
CREATE TABLE memory_edits (
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
  path         TEXT NOT NULL DEFAULT '',
  before       TEXT NOT NULL DEFAULT '',
  existed      INTEGER NOT NULL DEFAULT 0
);
INSERT INTO memory_edits (id, pr, learnings, pools, merged_at, baseline, state, after, reason, evaluated_at, created_at, path, before, existed)
SELECT CAST(key AS INTEGER), json_extract(value, '$.pr'), json_extract(value, '$.learnings'),
  json_extract(value, '$.pools'), json_extract(value, '$.merged_at'), json_extract(value, '$.baseline'),
  state, json_extract(value, '$.after'), json_extract(value, '$.reason'), json_extract(value, '$.evaluated_at'),
  created_at, json_extract(value, '$.path'), json_extract(value, '$.before'), json_extract(value, '$.existed')
FROM records WHERE collection = 'memory_edits';
DELETE FROM records WHERE collection = 'memory_edits';
DROP INDEX IF EXISTS ru_memory_edits_pr;
CREATE INDEX memory_edits_state ON memory_edits(state, merged_at);
CREATE UNIQUE INDEX memory_edits_pr ON memory_edits(pr) WHERE pr != '';
