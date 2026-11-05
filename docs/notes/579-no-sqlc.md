# sqlc removed (#579)

2026-11-05. Step 3, last group.

- sqlc is gone: `sqlc.yaml`, `pkg/store/sqlite/queries/`, the
  `go tool sqlc` dependency and the "sqlc generated code is current"
  CI step. The remaining ticket queries (writes, claims, edges, labels,
  slots, comment search) were sqlc output and now live, edited by hand,
  in `pkg/store/sqlstore/sqldb` (renamed from `sqlcdb`). Collections and
  `List` already built their SQL in Go with bqb.
- The SQL stays written for SQLite; Postgres still rewrites it as it
  runs (`postgres/rewrite.go`) with full-text overrides in
  `postgres/queries/overrides.sql`. Each query keeps its `-- name: X`
  line so `TestEveryQueryPrepares` and the override lookup find it.
- The unused `ListStepsCreated` query is deleted.
- Final count: `store.Store` has 20 methods (the 19 of the approved
  interface plus `Recall` for comment search), `store.Collection` 9.
- Not done here: Postgres on a pgx pool instead of `database/sql`, and
  dropping the rewriter by giving each backend its own SQL. Those are
  follow-ups, filed separately.
