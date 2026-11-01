-- +goose Up
-- Schema source for sqlc. See docs/data-model.md for the design.

CREATE TABLE tickets (
  id           TEXT PRIMARY KEY,
  project      TEXT NOT NULL,
  kind         TEXT NOT NULL,
  title        TEXT NOT NULL DEFAULT '',
  body         TEXT NOT NULL DEFAULT '',
  status       TEXT NOT NULL DEFAULT 'open'
               CHECK (status IN ('open','in_progress','closed')),
  resolution   TEXT NOT NULL DEFAULT '',
  priority     INTEGER NOT NULL DEFAULT 2,
  pool         TEXT NOT NULL DEFAULT '',
  holder       TEXT NOT NULL DEFAULT '',
  lease_until  INTEGER NOT NULL DEFAULT 0,
  fence        INTEGER NOT NULL DEFAULT 0,
  parent_id    TEXT REFERENCES tickets(id),
  defer_until  INTEGER NOT NULL DEFAULT 0,
  ephemeral    INTEGER NOT NULL DEFAULT 0,
  fields       TEXT NOT NULL DEFAULT '{}',
  meta         TEXT NOT NULL DEFAULT '{}',
  revision     INTEGER NOT NULL DEFAULT 0,
  created_by   TEXT NOT NULL DEFAULT '',
  created_at   INTEGER NOT NULL,
  updated_at   INTEGER NOT NULL,
  closed_at    INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX tickets_queue  ON tickets(project, pool, status, priority, created_at);
CREATE INDEX tickets_holder ON tickets(holder) WHERE holder <> '';
CREATE INDEX tickets_lease  ON tickets(lease_until) WHERE status = 'in_progress';
CREATE INDEX tickets_parent ON tickets(parent_id);
CREATE INDEX tickets_kind   ON tickets(kind, status);
CREATE INDEX tickets_message_to ON tickets(json_extract(fields, '$.to'))
  WHERE kind = 'message';
CREATE UNIQUE INDEX tickets_session_key ON tickets(json_extract(fields, '$.key'))
  WHERE kind = 'session' AND status <> 'closed' AND json_extract(fields, '$.key') IS NOT NULL;

CREATE TABLE edges (
  from_id    TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  to_id      TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  kind       TEXT NOT NULL CHECK (kind IN ('blocks','waits_for','tracks')),
  created_at INTEGER NOT NULL,
  PRIMARY KEY (from_id, to_id, kind)
);
CREATE INDEX edges_to ON edges(to_id, kind);

CREATE TABLE labels (
  ticket_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  label     TEXT NOT NULL,
  PRIMARY KEY (ticket_id, label)
);
CREATE INDEX labels_label ON labels(label);

CREATE TABLE history (
  seq        INTEGER PRIMARY KEY AUTOINCREMENT,
  ticket_id  TEXT NOT NULL,
  at         INTEGER NOT NULL,
  actor      TEXT NOT NULL,
  op         TEXT NOT NULL,
  revision   INTEGER NOT NULL,
  diff       TEXT NOT NULL DEFAULT '{}'
);
CREATE INDEX history_ticket ON history(ticket_id, seq);

CREATE TABLE comments (
  id         INTEGER PRIMARY KEY,
  ticket_id  TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  author     TEXT NOT NULL,
  body       TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  edited_at  INTEGER NOT NULL DEFAULT 0,
  deleted    INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX comments_ticket ON comments(ticket_id, id);

-- +goose Down
DROP TABLE comments;
DROP TABLE history;
DROP TABLE labels;
DROP TABLE edges;
DROP TABLE tickets;
