# One SQL dialect for sqldb (#694)

2026-11-05.

- `pkg/store/sqlstore/sqldb` queries are written once in SQL both
  engines run: `$N` placeholders, `CAST(.. AS BIGINT)`, and
  `CAST(CAST(EXISTS ..) AS INT) AS BIGINT` for booleans. modernc SQLite
  binds `$N` by number.
- `postgres/rewrite.go` and `postgres/queries/overrides.sql` are gone.
  Comment search, the one engine-specific query, is `Driver.RecallSQL`,
  set by each backend; `sqlstore.Recall` runs it.
- `TestEveryQueryPrepares` prepares every sqldb query as is on Postgres;
  `TestNoSQLitePlaceholders` fails on a `?` or `AS INTEGER)` in sqldb.
- Postgres stays on the `database/sql` pgx shim: the core takes
  `*sql.DB`/`*sql.Tx`. A pgxpool move needs a core over pgx types: #696.
