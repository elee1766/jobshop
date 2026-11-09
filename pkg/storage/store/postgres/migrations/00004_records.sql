-- +goose Up
-- Store collections (#579), as sqlite 00019, but partitioned by
-- collection. Registration creates each collection's partition; there is
-- no DEFAULT partition, so a write to an unregistered collection fails.
-- Indexes on the parent apply to every partition.
CREATE TABLE records (
  collection  TEXT NOT NULL,
  key         TEXT NOT NULL,
  state       TEXT NOT NULL DEFAULT '',
  value       JSONB NOT NULL DEFAULT '{}',
  revision    BIGINT NOT NULL DEFAULT 1,
  holder      TEXT NOT NULL DEFAULT '',
  lease_until BIGINT NOT NULL DEFAULT 0,
  attempts    BIGINT NOT NULL DEFAULT 0,
  created_at  BIGINT NOT NULL,
  updated_at  BIGINT NOT NULL,
  PRIMARY KEY (collection, key)
) PARTITION BY LIST (collection);
CREATE INDEX records_state ON records (state, lease_until, created_at);
CREATE INDEX records_created ON records (created_at);
CREATE INDEX records_updated ON records (state, updated_at);

-- +goose Down
DROP TABLE records;
