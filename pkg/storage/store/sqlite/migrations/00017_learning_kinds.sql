-- +goose Up
-- Learning kinds are config (#524): the kind CHECK becomes a name pattern
-- check; the controller checks kinds against learnings.kinds. SQLite
-- cannot drop a CHECK in place, so the table is rebuilt with its ids. The
-- recall triggers are dropped first so the rebuild leaves recall rows
-- alone, then recreated as in 00004.
DROP TRIGGER recall_learning_insert;
DROP TRIGGER recall_learning_update;
DROP TRIGGER recall_learning_delete;
CREATE TABLE learnings_new (
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
INSERT INTO learnings_new (id, kind, summary, fix, scope, ticket, session, source, status, created_at, pr, gate)
SELECT id, kind, summary, fix, scope, ticket, session, source, status, created_at, pr, gate FROM learnings;
DROP TABLE learnings;
ALTER TABLE learnings_new RENAME TO learnings;
CREATE INDEX learnings_created ON learnings(created_at);
CREATE INDEX learnings_scope ON learnings(scope, created_at);

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

-- +goose Down
-- Custom kinds cannot go back under the old CHECK; they become workaround.
UPDATE learnings SET kind = 'workaround' WHERE kind NOT IN ('correction','failure','workaround');
