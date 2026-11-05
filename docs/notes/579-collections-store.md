# Store collections, part 2: the generic layer (#579)

2026-11-05. Step 2 of `docs/notes/579-collections.md`: `Store.Collection`
and the `Collection` handle on both backends. No domain uses it yet;
poll_cursors, write_keys, memory_edits, learnings and inbox move in
their own PRs.

- Types, spec validation and value checks: `pkg/store/collection.go`.
  SQL: each backend owns its own, built with github.com/nullism/bqb
  (`pkg/store/sqlite/collections.go` with `?`, `pkg/store/postgres/collections.go`
  with `$n`). No shared dialect; some duplication is expected.
  `sqlstore` only validates the spec and calls the backend's
  `Driver.Collection`. Values are always bound. The only literal is the
  collection name in SQLite queries and DDL (`quoteLit`), because SQLite
  uses a partial index only when the WHERE repeats its literal; Postgres
  binds it.
- Migrations: sqlite `00019_records.sql` (one `records` table,
  `WITHOUT ROWID`), postgres `00004_records.sql` (`PARTITION BY LIST
  (collection)`, no DEFAULT partition).
- Registration runs its DDL once per spec per process, `IF NOT EXISTS`.
  SQLite: `ri_<name>_<field>` partial indexes on `(field, created_at,
  key)`, `ru_<name>_<field>` unique partial indexes on non-empty values,
  `fts_<name>` FTS5 table. Postgres: partition `records_<name>`,
  `rf_<path>` field indexes on the parent (shared by collections that
  declare the same path), unique and GIN search indexes on the
  partition.
- Where values compare as text (`CAST(json_extract ...)` / `#>>`), so a
  JSON number 12 matches `"12"`.
- Update rules: revision must match (`ErrConflict`); a live lease
  admits only its holder (`ErrNotHolder`); zero `LeaseUntil` releases.
  Claim bumps revision and attempts; Renew bumps neither.
- Search uses the backend's recall match (all words, then any word).
  Postgres ranks `ts_rank / ln(2 + words)`, as the recall query, so both
  backends put a short focused match first.

Tested by the `storetest` Collection cases on SQLite (single and
concurrent connections) and Postgres, an `EXPLAIN QUERY PLAN` check
that field and unique lookups use their partial indexes (SQLite, after
`ANALYZE`), and a Postgres check that registration creates the partition,
an unregistered collection's insert fails, and lookups use the indexes.
