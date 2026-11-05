# Postgres store backend

2026-11-05, #493. `pkg/store/postgres`.

- Same core as SQLite (`pkg/store/sqlstore`) and the same sqlc code
  (`sqlcdb`, generated from `pkg/store/sqlite/queries`). Bind wraps the
  connection so each query text is rewritten once and cached:
  placeholders to `$N` (a bare `?` numbers after the highest seen, as
  SQLite does), `AS INTEGER)` to `AS BIGINT)`, `CAST(EXISTS ...)` gets
  an inner `AS INT`, and `?N IS NULL` gets `CAST(.. AS TEXT)` so
  Postgres can type it.
- `queries/overrides.sql` replaces whole queries by name. Only
  RecallSearch: a `tsvector` column with a GIN index, `ts_rank` divided
  by the log of the body's word count for ordering (bm25 favors short
  matches; ts_rank's own normalization counts lexemes, not words),
  `ts_headline` for the `[..]`
  snippet. Comment and learning triggers keep `recall` current, as the
  FTS5 triggers do.
- `migrations/00001_init.sql` is the schema as of sqlite 00016, with
  BIGINT integers and identity ids. New sqlite migrations need a
  Postgres one; `TestEveryQueryPrepares` catches a missing column.
- Transactions run serializable (DSN runtime param). Serialization
  failures and deadlocks count as busy, so the core reruns them, as with
  SQLITE_BUSY. Pool: 8 connections.
- Config: `store: {driver: postgres, dsn_file: <secret>}`; the file
  holds a URL or key=value DSN. Default stays SQLite at `jobshop.db`.
  Read at controller start.
- Tests skip unless `JOBSHOP_TEST_POSTGRES` names a server; they clone a
  migrated template database per test. CI's `postgres` job sets it to a
  `postgres:17` service.
- No data migration from SQLite yet.
