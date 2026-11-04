# Store core split (#404)

2026-11-04. Prep for a possible second backend, not a decision to build one.

- `pkg/store/sqlstore` holds the backend-neutral store: transactions,
  `mutate` with the revision check, history, claim rules, inbox, learnings,
  write keys and row mapping. It runs on `sqlcdb.Querier`, not on
  `*sqlcdb.Queries`.
- `pkg/store/sqlite` is the driver: DSN (`_txlock=immediate`, pragmas),
  goose migrations, the SQL files and `sqlc.yaml`, plus two hooks in
  `sqlstore.Driver`: `Classify` (constraint error to store error) and
  `RecallMatch` (FTS5 quoting, all words then any word).
- sqlc generates into `pkg/store/sqlstore/sqlcdb` with
  `sql_package: database/sql` and `emit_interface: true`. A second engine
  keeps the same parameter and row types (chosen over a converting
  adapter: no conversion code) and passes its own `Bind`.
- `sqlite.Store` and `sqlite.Options` are aliases, so callers did not
  change. storetest runs unchanged.
