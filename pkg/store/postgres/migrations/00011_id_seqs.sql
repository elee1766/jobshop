-- +goose Up
-- Per-prefix counters for the seq id generator (JOB-57), as sqlite 00025.
CREATE TABLE id_seqs (
  prefix TEXT PRIMARY KEY,
  last   BIGINT NOT NULL
);

-- +goose Down
DROP TABLE id_seqs;
