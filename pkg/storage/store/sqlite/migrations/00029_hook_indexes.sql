-- +goose Up
-- The hook collections' indexes, as registration creates them. Created at
-- registration, each fresh database changed its schema three times after
-- Open, and every pooled connection then reloaded the schema and
-- re-prepared its statements (JOB-187). Here they are in the schema
-- before the first connection and registration finds nothing to do.
CREATE INDEX IF NOT EXISTS ri_hook_invocations_run ON records (CAST(json_extract(value, '$.run') AS TEXT), created_at, key) WHERE collection = 'hook_invocations';
CREATE INDEX IF NOT EXISTS ri_hook_invocations_trigger ON records (CAST(json_extract(value, '$.trigger') AS TEXT), created_at, key) WHERE collection = 'hook_invocations';
CREATE INDEX IF NOT EXISTS ri_hook_invocations_event ON records (CAST(json_extract(value, '$.event.delivery') AS TEXT), created_at, key) WHERE collection = 'hook_invocations';
CREATE INDEX IF NOT EXISTS ri_hook_calls_invocation ON records (CAST(json_extract(value, '$.invocation') AS TEXT), created_at, key) WHERE collection = 'hook_calls';
CREATE INDEX IF NOT EXISTS ri_hook_attempts_invocation ON records (CAST(json_extract(value, '$.invocation') AS TEXT), created_at, key) WHERE collection = 'hook_attempts';

-- +goose Down
-- Registration creates them again on the next start.
DROP INDEX IF EXISTS ri_hook_invocations_run;
DROP INDEX IF EXISTS ri_hook_invocations_trigger;
DROP INDEX IF EXISTS ri_hook_invocations_event;
DROP INDEX IF EXISTS ri_hook_calls_invocation;
DROP INDEX IF EXISTS ri_hook_attempts_invocation;
