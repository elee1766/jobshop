-- +goose Up
-- Recall (#265): one FTS5 index over ticket comments, session journal
-- text and learnings (#264). Triggers keep comments and learnings
-- current; the
-- controller writes journal text after each turn.
CREATE VIRTUAL TABLE recall USING fts5(
  body,
  kind UNINDEXED,
  ref UNINDEXED,
  ticket_id UNINDEXED,
  project UNINDEXED,
  at UNINDEXED,
  tokenize = 'porter unicode61'
);

-- +goose StatementBegin
CREATE TRIGGER recall_comment_insert AFTER INSERT ON comments
WHEN NEW.deleted = 0
BEGIN
  INSERT INTO recall (body, kind, ref, ticket_id, project, at)
  SELECT NEW.body, 'comment', CAST(NEW.id AS TEXT), NEW.ticket_id,
         COALESCE((SELECT project FROM tickets WHERE id = NEW.ticket_id), ''), NEW.created_at;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER recall_comment_update AFTER UPDATE OF body, deleted ON comments
BEGIN
  DELETE FROM recall WHERE kind = 'comment' AND ref = CAST(OLD.id AS TEXT);
  INSERT INTO recall (body, kind, ref, ticket_id, project, at)
  SELECT NEW.body, 'comment', CAST(NEW.id AS TEXT), NEW.ticket_id,
         COALESCE((SELECT project FROM tickets WHERE id = NEW.ticket_id), ''), NEW.created_at
  WHERE NEW.deleted = 0;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER recall_comment_delete AFTER DELETE ON comments
BEGIN
  DELETE FROM recall WHERE kind = 'comment' AND ref = CAST(OLD.id AS TEXT);
END;
-- +goose StatementEnd

-- Backfill comments written before this migration.
INSERT INTO recall (body, kind, ref, ticket_id, project, at)
SELECT c.body, 'comment', CAST(c.id AS TEXT), c.ticket_id, COALESCE(t.project, ''), c.created_at
FROM comments c LEFT JOIN tickets t ON t.id = c.ticket_id
WHERE c.deleted = 0;

-- Learnings (#264): summary and fix are searchable. Project comes from the
-- learning's ticket; a learning with no ticket has project ''.
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

INSERT INTO recall (body, kind, ref, ticket_id, project, at)
SELECT l.summary || char(10) || l.fix, 'learning', CAST(l.id AS TEXT), l.ticket, COALESCE(t.project, ''), l.created_at
FROM learnings l LEFT JOIN tickets t ON t.id = l.ticket;

-- +goose Down
DROP TRIGGER recall_learning_delete;
DROP TRIGGER recall_learning_update;
DROP TRIGGER recall_learning_insert;
DROP TRIGGER recall_comment_delete;
DROP TRIGGER recall_comment_update;
DROP TRIGGER recall_comment_insert;
DROP TABLE recall;
