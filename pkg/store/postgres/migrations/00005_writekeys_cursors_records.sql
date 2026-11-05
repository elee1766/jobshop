-- +goose Up
-- write_keys and poll_cursors move onto store collections (#579), as
-- sqlite 00020. Each gets its partition, as registration would create.
CREATE TABLE IF NOT EXISTS records_write_keys PARTITION OF records FOR VALUES IN ('write_keys');
INSERT INTO records (collection, key, state, value, created_at, updated_at)
SELECT 'write_keys', key, CASE WHEN done = 1 THEN 'done' ELSE 'pending' END,
  jsonb_build_object('op', op, 'result', result, 'actor', actor), created_at, updated_at
FROM write_keys;
DROP TABLE write_keys;

CREATE TABLE IF NOT EXISTS records_poll_cursors PARTITION OF records FOR VALUES IN ('poll_cursors');
INSERT INTO records (collection, key, state, value, created_at, updated_at)
SELECT 'poll_cursors', poller || ':' || repo || ':' || stream, '',
  jsonb_build_object('poller', poller, 'repo', repo, 'stream', stream, 'since', since, 'etag', etag),
  updated_at, updated_at
FROM poll_cursors;
DROP TABLE poll_cursors;

-- +goose Down
CREATE TABLE write_keys (
  key        TEXT PRIMARY KEY,
  op         TEXT NOT NULL,
  result     TEXT NOT NULL DEFAULT '',
  done       BIGINT NOT NULL DEFAULT 0,
  actor      TEXT NOT NULL DEFAULT '',
  created_at BIGINT NOT NULL,
  updated_at BIGINT NOT NULL
);
INSERT INTO write_keys (key, op, result, done, actor, created_at, updated_at)
SELECT key, value->>'op', value->>'result', CASE WHEN state = 'done' THEN 1 ELSE 0 END,
  value->>'actor', created_at, updated_at
FROM records WHERE collection = 'write_keys';
DROP TABLE records_write_keys;

CREATE TABLE poll_cursors (
  poller     TEXT NOT NULL,
  repo       TEXT NOT NULL,
  stream     TEXT NOT NULL,
  since      BIGINT NOT NULL DEFAULT 0,
  etag       TEXT NOT NULL DEFAULT '',
  updated_at BIGINT NOT NULL,
  PRIMARY KEY (poller, repo, stream)
);
INSERT INTO poll_cursors (poller, repo, stream, since, etag, updated_at)
SELECT value->>'poller', value->>'repo', value->>'stream', (value->>'since')::BIGINT,
  value->>'etag', updated_at
FROM records WHERE collection = 'poll_cursors';
DROP TABLE records_poll_cursors;
