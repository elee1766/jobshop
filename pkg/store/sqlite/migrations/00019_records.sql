-- +goose Up
-- Store collections (#579, docs/notes/579-collections.md): one records
-- table for every collection. Registration adds partial indexes per
-- collection (WHERE collection = '<name>') and an FTS5 table per
-- collection with search fields. Domains move onto it in later
-- migrations, one per domain.
CREATE TABLE records (
  collection  TEXT NOT NULL,
  key         TEXT NOT NULL,
  state       TEXT NOT NULL DEFAULT '',
  value       TEXT NOT NULL DEFAULT '{}',
  revision    INTEGER NOT NULL DEFAULT 1,
  holder      TEXT NOT NULL DEFAULT '',
  lease_until INTEGER NOT NULL DEFAULT 0,
  attempts    INTEGER NOT NULL DEFAULT 0,
  created_at  INTEGER NOT NULL,
  updated_at  INTEGER NOT NULL,
  PRIMARY KEY (collection, key)
) WITHOUT ROWID;
CREATE INDEX records_state ON records (collection, state, lease_until, created_at);
CREATE INDEX records_created ON records (collection, created_at);
CREATE INDEX records_updated ON records (collection, state, updated_at);

-- +goose Down
DROP TABLE records;
