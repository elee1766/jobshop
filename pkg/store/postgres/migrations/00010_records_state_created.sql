-- +goose Up
-- Count by state and age (the inbox backlog, JOB-61), as sqlite 00025.
-- records is partitioned by collection, so the index leads with state.
CREATE INDEX IF NOT EXISTS records_state_created ON records (state, created_at);

-- +goose Down
DROP INDEX IF EXISTS records_state_created;
