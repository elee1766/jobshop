-- +goose Up
-- Ingest poller cursors (#341): one row per poller, repo and event
-- stream. since is the newest updated_at handled (unix ms), etag the
-- last response's ETag, so a restart resumes and an idle poll is a 304.
CREATE TABLE poll_cursors (
  poller     TEXT NOT NULL,
  repo       TEXT NOT NULL,
  stream     TEXT NOT NULL,
  since      INTEGER NOT NULL DEFAULT 0,
  etag       TEXT NOT NULL DEFAULT '',
  updated_at INTEGER NOT NULL,
  PRIMARY KEY (poller, repo, stream)
);

-- +goose Down
DROP TABLE poll_cursors;
