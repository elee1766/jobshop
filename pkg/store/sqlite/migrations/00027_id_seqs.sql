-- +goose Up
-- Per-prefix counters for the seq id generator (JOB-57). A row holds the
-- last number handed out; the insert transaction bumps it.
CREATE TABLE id_seqs (
  prefix TEXT PRIMARY KEY,
  last   INTEGER NOT NULL
);

-- +goose Down
DROP TABLE id_seqs;
