-- +goose Up
-- Runs (#89): key dedup for open runs and run state listing.
CREATE INDEX tickets_run_key ON tickets(json_extract(fields, '$.key'))
  WHERE kind = 'run' AND status <> 'closed';
CREATE INDEX tickets_run_state ON tickets(json_extract(fields, '$.state'))
  WHERE kind = 'run';

-- +goose Down
DROP INDEX tickets_run_state;
DROP INDEX tickets_run_key;
