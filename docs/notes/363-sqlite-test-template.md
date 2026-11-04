# #363: pkg/store/sqlite tests run from a migrated template

2026-11-04.

`pkg/store/sqlite/main_test.go` migrates one database in `TestMain`, reads the
file after `Close` (which checkpoints the WAL), and each test copies it into
its own temp dir (`testPath`, `openTest`). `Open` still runs goose, which
finds every migration applied and does nothing. Running all the migrations
per conformance case was most of `TestConformance`'s 3.9s under `-race`.
It now takes about 1s. Top-level tests and the `storetest.Run` cases are
`t.Parallel()`, because each one has its own database file and clock.
The package runs in about 2.3s under `-race` on a loaded gasserver (load 15),
about 1.2s of which is the test binary itself.

`TestRecallBackfill` still exercises real migration: it rolls the copied
database back past 00004 and reopens it.
