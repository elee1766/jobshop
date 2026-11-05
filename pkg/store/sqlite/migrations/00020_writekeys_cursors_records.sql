-- +goose Up
-- write_keys and poll_cursors move onto store collections (#579,
-- docs/notes/579-collections.md). Neither declares a field, so there are
-- no per-collection indexes to create.
INSERT INTO records (collection, key, state, value, created_at, updated_at)
SELECT 'write_keys', key, CASE WHEN done = 1 THEN 'done' ELSE 'pending' END,
  json_object('op', op, 'result', result, 'actor', actor), created_at, updated_at
FROM write_keys;
DROP TABLE write_keys;

INSERT INTO records (collection, key, state, value, created_at, updated_at)
SELECT 'poll_cursors', poller || ':' || repo || ':' || stream, '',
  json_object('poller', poller, 'repo', repo, 'stream', stream, 'since', since, 'etag', etag),
  updated_at, updated_at
FROM poll_cursors;
DROP TABLE poll_cursors;

-- +goose Down
CREATE TABLE write_keys (
  key        TEXT PRIMARY KEY,
  op         TEXT NOT NULL,
  result     TEXT NOT NULL DEFAULT '',
  done       INTEGER NOT NULL DEFAULT 0,
  actor      TEXT NOT NULL DEFAULT '',
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL
);
INSERT INTO write_keys (key, op, result, done, actor, created_at, updated_at)
SELECT key, json_extract(value, '$.op'), json_extract(value, '$.result'),
  CASE WHEN state = 'done' THEN 1 ELSE 0 END, json_extract(value, '$.actor'), created_at, updated_at
FROM records WHERE collection = 'write_keys';
DELETE FROM records WHERE collection = 'write_keys';

CREATE TABLE poll_cursors (
  poller     TEXT NOT NULL,
  repo       TEXT NOT NULL,
  stream     TEXT NOT NULL,
  since      INTEGER NOT NULL DEFAULT 0,
  etag       TEXT NOT NULL DEFAULT '',
  updated_at INTEGER NOT NULL,
  PRIMARY KEY (poller, repo, stream)
);
INSERT INTO poll_cursors (poller, repo, stream, since, etag, updated_at)
SELECT json_extract(value, '$.poller'), json_extract(value, '$.repo'), json_extract(value, '$.stream'),
  json_extract(value, '$.since'), json_extract(value, '$.etag'), updated_at
FROM records WHERE collection = 'poll_cursors';
DELETE FROM records WHERE collection = 'poll_cursors';
