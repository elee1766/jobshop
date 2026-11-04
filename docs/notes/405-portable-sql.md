# Portable store SQL (#405)

2026-11-04. Prep for a possible Postgres backend; not a decision to build one.

Migration 00013 removes the SQLite-only SQL from ticket queries:

- `f_key`, `f_to`, `f_state`, `f_workflow`, `f_index` and `f_waiting`
  are plain columns on `tickets`. The store writes them from `fields` on
  every insert and write (`fieldColumns` in `pkg/store/sqlite/columns.go`).
  `fields` stays the source of record. A key that is absent or not the
  expected JSON type is NULL, as `json_extract` gave. The indexes
  `tickets_message_to`, `tickets_session_key`, `tickets_run_key` and
  `tickets_run_state` are rebuilt on the columns, and `tickets_children`
  is added for step order.
- The meta filter on `List` (`MetaPack`/`MetaKey`/`MetaValue`) reads
  `ticket_meta`: one row per string value `meta[pack][key]`. The store
  rewrites a ticket's rows on every write (`syncMeta`).
- Kind and pool lists are `sqlc.slice` parameters instead of
  `json_each`. In such a query every parameter is a bare `?`: sqlc gives
  named parameters fixed `?N` numbers from the unexpanded query, and a
  list of two or more would shift them. `InboxNextClaimable` puts `now`
  before its slice for the same reason.
- Labels use `INSERT ... ON CONFLICT DO NOTHING`.

The migration backfills the columns and `ticket_meta` with JSON
functions once. That is SQLite DDL, like the column types.

Still engine-specific:
- recall: FTS5 `MATCH`, `bm25`, `snippet` and the 00004 triggers, behind
  `IndexRecall`/`Recall`;
- `INTEGER PRIMARY KEY AUTOINCREMENT` on history and inbox, which are DDL
  types;
- `json_each` on `fields.resources` from #400, which landed after this
  change (#426).

Tests: storetest conformance passes. `TestPortableColumnsBackfill` rolls
00013 back on a database with runs, steps, a waiting task and meta, then
reopens it and checks the lookups.
