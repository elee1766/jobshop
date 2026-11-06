-- +goose Up
-- Count by state and age (the inbox backlog, JOB-61). records_state has
-- lease_until before created_at, so a created_at range could not use it
-- and SQLite walked records_created, reading every old row's state.
CREATE INDEX IF NOT EXISTS records_state_created ON records (collection, state, created_at);

-- +goose Down
DROP INDEX IF EXISTS records_state_created;
