# JOB-42: standalone SQLite store, sqlstore removed

2026-11-06

`pkg/store/sqlstore` and its `Driver` are gone. `pkg/store/sqlite` has
its own `Store`; its queries moved to `pkg/store/sqlite/sqldb` (the
Postgres copy is `pkg/store/postgres/pgdb`, JOB-41). The two stores share
only `pkg/store` types and `storerules`.

- Production: one writer connection with `_txlock=immediate` (BEGIN
  IMMEDIATE), so a write never loses a lock upgrade and is not retried.
  Reads outside a transaction use the query-only reader pool (#742).
  Both have their own statement cache (#433).
- `Options.Concurrent` stays: `TestConformanceConcurrent` uses it to
  check claims do not rely on a single writer (#406). Only that mode
  opens deferred transactions on 8 connections and reruns on SQLITE_BUSY.
- References in `docs/layout.md` (dependency graph) and AGENTS.md now
  name the per-engine query packages.
