# JOB-61: inbox backlog count index

2026-11-06

`ingest.InboxCountOlder` (Attention's backlog line and the
`ingest_backlog` stall signal) counts one state over a `created_at`
range. `records_state` is `(collection, state, lease_until, created_at)`,
so SQLite chose `records_created` and read every old inbox row: 2-4 s on
gasserver with 10.6k records.

- New index `records_state_created`: `(collection, state, created_at)`
  on SQLite (00025), `(state, created_at)` on the partitioned Postgres
  table (00010).
- `TestCollectionCountPlan` checks the plan with `EXPLAIN QUERY PLAN`.
- Retention for terminal inbox records is a separate change.
