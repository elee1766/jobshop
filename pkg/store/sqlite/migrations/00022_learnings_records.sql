-- +goose Up
-- learnings move onto the learnings store collection (#579,
-- pkg/learning). The key is the old id as text; the state is the
-- status; the value keeps the columns plus the ticket's project, which
-- recall filters on. Indexes and fts_learnings are what registration
-- creates. Learning rows leave the shared recall index: recall reads
-- fts_learnings instead.
CREATE INDEX IF NOT EXISTS ri_learnings_kind ON records (CAST(json_extract(value, '$.kind') AS TEXT), created_at, key) WHERE collection = 'learnings';
CREATE INDEX IF NOT EXISTS ri_learnings_scope ON records (CAST(json_extract(value, '$.scope') AS TEXT), created_at, key) WHERE collection = 'learnings';
CREATE INDEX IF NOT EXISTS ri_learnings_ticket ON records (CAST(json_extract(value, '$.ticket') AS TEXT), created_at, key) WHERE collection = 'learnings';
CREATE INDEX IF NOT EXISTS ri_learnings_pr ON records (CAST(json_extract(value, '$.pr') AS TEXT), created_at, key) WHERE collection = 'learnings';
CREATE VIRTUAL TABLE IF NOT EXISTS fts_learnings USING fts5(key UNINDEXED, body);

INSERT INTO records (collection, key, state, value, created_at, updated_at)
SELECT 'learnings', CAST(l.id AS TEXT), l.status,
  json_object('id', l.id, 'kind', l.kind, 'summary', l.summary, 'fix', l.fix, 'scope', l.scope,
    'ticket', l.ticket, 'session', l.session, 'source', l.source, 'pr', l.pr, 'gate', l.gate,
    'project', COALESCE((SELECT project FROM tickets WHERE id = l.ticket), '')),
  l.created_at, l.created_at
FROM learnings l;
INSERT INTO fts_learnings (key, body)
SELECT CAST(id AS TEXT), summary || char(10) || fix FROM learnings;

DROP TRIGGER recall_learning_insert;
DROP TRIGGER recall_learning_update;
DROP TRIGGER recall_learning_delete;
DELETE FROM recall WHERE kind = 'learning';
DROP TABLE learnings;

-- +goose Down
CREATE TABLE learnings (
  id         INTEGER PRIMARY KEY,
  kind       TEXT NOT NULL CHECK (kind != '' AND kind NOT GLOB '*[^a-z0-9-]*' AND kind GLOB '[a-z]*'),
  summary    TEXT NOT NULL,
  fix        TEXT NOT NULL DEFAULT '',
  scope      TEXT NOT NULL DEFAULT 'global',
  ticket     TEXT NOT NULL DEFAULT '',
  session    TEXT NOT NULL DEFAULT '',
  source     TEXT NOT NULL CHECK (source IN ('agent','controller')),
  status     TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','applied','rejected')),
  created_at INTEGER NOT NULL,
  pr         TEXT NOT NULL DEFAULT '',
  gate       TEXT NOT NULL DEFAULT ''
);
CREATE INDEX learnings_created ON learnings(created_at);
CREATE INDEX learnings_scope ON learnings(scope, created_at);
INSERT INTO learnings (id, kind, summary, fix, scope, ticket, session, source, status, created_at, pr, gate)
SELECT CAST(key AS INTEGER), json_extract(value, '$.kind'), json_extract(value, '$.summary'),
  json_extract(value, '$.fix'), json_extract(value, '$.scope'), json_extract(value, '$.ticket'),
  json_extract(value, '$.session'), json_extract(value, '$.source'), state, created_at,
  json_extract(value, '$.pr'), json_extract(value, '$.gate')
FROM records WHERE collection = 'learnings';
INSERT INTO recall (body, kind, ref, ticket_id, project, at)
SELECT summary || char(10) || fix, 'learning', CAST(id AS TEXT), ticket,
       COALESCE((SELECT project FROM tickets WHERE id = learnings.ticket), ''), created_at
FROM learnings;

-- +goose StatementBegin
CREATE TRIGGER recall_learning_insert AFTER INSERT ON learnings
BEGIN
  INSERT INTO recall (body, kind, ref, ticket_id, project, at)
  SELECT NEW.summary || char(10) || NEW.fix, 'learning', CAST(NEW.id AS TEXT), NEW.ticket,
         COALESCE((SELECT project FROM tickets WHERE id = NEW.ticket), ''), NEW.created_at;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER recall_learning_update AFTER UPDATE OF summary, fix ON learnings
BEGIN
  DELETE FROM recall WHERE kind = 'learning' AND ref = CAST(OLD.id AS TEXT);
  INSERT INTO recall (body, kind, ref, ticket_id, project, at)
  SELECT NEW.summary || char(10) || NEW.fix, 'learning', CAST(NEW.id AS TEXT), NEW.ticket,
         COALESCE((SELECT project FROM tickets WHERE id = NEW.ticket), ''), NEW.created_at;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER recall_learning_delete AFTER DELETE ON learnings
BEGIN
  DELETE FROM recall WHERE kind = 'learning' AND ref = CAST(OLD.id AS TEXT);
END;
-- +goose StatementEnd

DELETE FROM records WHERE collection = 'learnings';
DROP TABLE IF EXISTS fts_learnings;
DROP INDEX IF EXISTS ri_learnings_kind;
DROP INDEX IF EXISTS ri_learnings_scope;
DROP INDEX IF EXISTS ri_learnings_ticket;
DROP INDEX IF EXISTS ri_learnings_pr;
