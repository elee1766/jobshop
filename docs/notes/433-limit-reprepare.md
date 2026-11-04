# Bound LIMIT re-prepares, template kept as bytes (#433, 2026-11-04)

Part of #433 (pkg/control -race time), after the statement cache (#458).
A CPU profile still showed 17% of CPU in `sqlite3Reprepare` and 14% in
the per-open `VACUUM INTO` of the template.

- SQLite re-prepares a statement when a bound `LIMIT ?` changes value
  (the plan reads it), so the cache paid a full parse whenever List ran
  with a different limit. Every query with a bound limit now says
  `LIMIT CAST(? AS INTEGER)`, which is evaluated at run time. sqlc
  rejects `LIMIT (SELECT ?)`; the bare `?` slice queries (Ready,
  Orphaned) keep bare `?` because numbered params break with
  sqlc.slice, so their param field is `ColumnN`.
- The template database is built once per process and kept as file
  bytes; a new file is written from them instead of a `VACUUM INTO`
  per open, which re-ran every CREATE statement.
- `go test -race -count=1 ./pkg/control` on gasserver: 82s to 61s wall,
  154s to 103s CPU. What is left is first-time prepares per controller
  and the tests' own passes.
