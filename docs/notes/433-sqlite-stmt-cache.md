# SQLite statement cache and seeded databases (#433, 2026-11-04)

Part of #433 (pkg/control -race time). A CPU profile of
`go test -race ./pkg/control` after PR #445 showed about 38% of CPU in
SQLite's parser and 17% in goose migrations at `sqlite.Open`.

- `pkg/store/sqlite` prepares each query text once (`stmtcache.go`) and
  reuses it; database/sql otherwise prepares and closes every ad hoc
  query. Texts first seen inside a transaction run unprepared and are
  prepared by the next call outside one: preparing on the DB inside a
  transaction would wait for the only connection. At most 512 texts
  (sqlc.slice queries make one text per argument count).
- A new database file is copied with `VACUUM INTO` from a migrated
  in-memory template built once per process (`template.go`), then
  linked into place, so an existing file is never replaced. Open still
  runs goose afterwards, which finds nothing to do. Production opens one
  database, so this only helps tests.
- Measured on a loaded gasserver: 138s to 63s wall, 290s to 120s CPU.
  What is left is per-connection schema parsing and the tests' own
  passes; 15s needs fewer controllers or passes per test.
