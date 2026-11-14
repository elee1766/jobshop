# Postgres store tests: connection bound

2026-11-14, JOB-209. `pkg/storage/store/postgres` tests.

- "sorry, too many clients already" (RaceSlotRelease,
  CollectionConcurrentClaims and others, 2026-11-10..12) came from
  parallel tests scaling with GOMAXPROCS, each opening two pools of 8
  against a server with `max_connections=100`.
- #1252 (#1251) fixed it in the harness (`postgres_test.go`): at most 6
  stores open at once (`openStores`), pools of 4 per store
  (`postgres.Options.MaxConns`, production default stays 8), admin pool
  of 4. About 64 connections worst case.
- Checked on master a98ce1bb against a local postgres 17 with
  `max_connections=100`: `GOMAXPROCS=32 go test -race -count=20
  ./pkg/storage/store/postgres/` passed (91s), no "too many" in the
  output. No code change was needed beyond #1252.
- If it returns, count the connections first
  (`SELECT count(*) FROM pg_stat_activity`) and lower `openStores`;
  a new test must open at most one store through `openNamed`.
