# JOB-41: standalone Postgres store

2026-11-06

`pkg/store/postgres` no longer uses `pkg/store/sqlstore`. It has its own
`Store`, queries (`pgdb`, copied from `sqldb` and now free to diverge),
transaction loop and collections. JOB-42 does the same for SQLite and
removes `sqlstore`.

- Two pools. Writes run in `BEGIN ISOLATION LEVEL SERIALIZABLE` on the
  write pool and rerun on SQLSTATE 40001/40P01 (up to 16 times, with
  jittered backoff). Reads outside a transaction use a second pool with
  `default_transaction_read_only=on` at the server default isolation,
  so Desk polls neither take predicate locks nor abort claims.
- `ClaimNext` picks its candidate with `FOR UPDATE OF t SKIP LOCKED`.
  Concurrent claimers skip a row another claimer holds instead of all
  reading it and losing the serializable commit.
- `Store.Retries()` counts reruns. `storetest.Bench` reports it as
  `retries/op` on the ClaimNext benchmarks for any store that has it.
- Connections default to read committed now, so `storemigrate.Copy`
  asks for serializable itself.
- Still `database/sql` over pgx's stdlib driver: the collections and
  `storemigrate` take `*sql.DB`/`*sql.Tx`, and a native `pgx.Tx` port can
  follow once SQLite no longer shares the code.
