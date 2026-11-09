-- +goose Up
-- Session journal summaries move from the shared recall index onto the
-- journals store collection (#579, pkg/journal). The key is the session
-- id (the old ref). Indexes and fts_journals are what registration
-- creates. recall then holds only comments.
CREATE INDEX IF NOT EXISTS ri_journals_ticket ON records (CAST(json_extract(value, '$.ticket') AS TEXT), created_at, key) WHERE collection = 'journals';
CREATE INDEX IF NOT EXISTS ri_journals_project ON records (CAST(json_extract(value, '$.project') AS TEXT), created_at, key) WHERE collection = 'journals';
CREATE VIRTUAL TABLE IF NOT EXISTS fts_journals USING fts5(key UNINDEXED, body);

INSERT OR IGNORE INTO records (collection, key, state, value, created_at, updated_at)
SELECT 'journals', ref, '',
  json_object('ticket', ticket_id, 'project', project, 'at', CAST(at AS INTEGER), 'text', body),
  CAST(at AS INTEGER), CAST(at AS INTEGER)
FROM recall WHERE kind = 'journal';
INSERT INTO fts_journals (key, body)
SELECT key, CAST(json_extract(value, '$.text') AS TEXT) FROM records WHERE collection = 'journals';
DELETE FROM recall WHERE kind = 'journal';

-- +goose Down
INSERT INTO recall (body, kind, ref, ticket_id, project, at)
SELECT json_extract(value, '$.text'), 'journal', key, json_extract(value, '$.ticket'),
  json_extract(value, '$.project'), json_extract(value, '$.at')
FROM records WHERE collection = 'journals';
DELETE FROM records WHERE collection = 'journals';
DROP TABLE IF EXISTS fts_journals;
DROP INDEX IF EXISTS ri_journals_ticket;
DROP INDEX IF EXISTS ri_journals_project;
