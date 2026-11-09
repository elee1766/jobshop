-- +goose Up
-- Learnings (#264): corrections, failures and workarounds recorded by
-- agents and the controller. A table, not tickets, so pools never see them.
CREATE TABLE learnings (
  id         INTEGER PRIMARY KEY,
  kind       TEXT NOT NULL CHECK (kind IN ('correction','failure','workaround')),
  summary    TEXT NOT NULL,
  fix        TEXT NOT NULL DEFAULT '',
  scope      TEXT NOT NULL DEFAULT 'global',
  ticket     TEXT NOT NULL DEFAULT '',
  session    TEXT NOT NULL DEFAULT '',
  source     TEXT NOT NULL CHECK (source IN ('agent','controller')),
  status     TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','applied','rejected')),
  created_at INTEGER NOT NULL
);
CREATE INDEX learnings_created ON learnings(created_at);
CREATE INDEX learnings_scope ON learnings(scope, created_at);

-- +goose Down
DROP TABLE learnings;
