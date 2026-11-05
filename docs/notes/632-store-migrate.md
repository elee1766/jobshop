# Store migrate: SQLite to Postgres

2026-11-05, #632. `pkg/store/storemigrate`, `jobshop admin store migrate`.

- `jobshop admin store migrate [--from sqlite:<path>] --to postgres
  [--dsn-file <f>] [--force]`. Runs without a controller and takes the
  controller lock, so it fails while one runs. Default source is the
  root's `jobshop.db`; default DSN file is the config's `store.dsn_file`.
- Both stores are opened (and migrated) by their backends, then rows
  are copied over database/sql in one Postgres transaction, tables in
  dependency order (`storemigrate.Tables`). Ids, revisions, history and
  timestamps are kept. Ticket parents are set after all tickets are in.
- Values are converted per target column type (SQLite keeps integers in
  TEXT columns such as `recall.at`); `records.value` is cast to jsonb,
  and each collection's partition is created as registration would.
- Recall: comment and learning rows come from the Postgres triggers as
  those rows are copied; only journal rows are read from FTS5.
- Identity sequences are set past the copied ids. Then row counts are
  compared per table; any difference rolls everything back.
- A target with rows is refused unless `--force`, which truncates it
  first.
- `TestTablesComplete` fails when a SQLite table is not in `Tables`.
  `TestCopy` fills SQLite with `storetest.Fill`, copies, reopens
  Postgres and reads it back with `storetest.CheckFilled`; it runs in
  CI's `postgres` job.
- NixOS: `services.jobshop.store.driver` and `.dsnFile` write config
  `store:` (in both managed and user mode); unset leaves it alone.
- Cutover runbook: `docs/verify/postgres-cutover.md`.
