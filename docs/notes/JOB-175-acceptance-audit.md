# JOB-175 acceptance audit of the package restructuring

2026-11-09. Audit of JOB-164 (JOB-165 to JOB-174) at master 15eb26d5.

## Result

Accepted. One stray directory removed here, one pre-existing bug filed (#1128).

## Checked

- Graph: 67 packages under 13 top-level families, 0 known archcheck
  violations. The sibling notes JOB-165 to JOB-174 match the tree.
- Enforcement: real prohibited imports fail `go run ./pkg/tooling/archcheck/cmd/archcheck`
  (`pack` importing `store`, `config` importing `cuelang.org/go/cue`). The
  deliberate-edge test covers the other rules.
- Gates: `go build`, `go vet`, `go test ./...`, `golangci-lint` clean.
  `go test -race` passes on pack, config, orchestration and transport.
- Storage: the SQLite suites pass. A disposable PostgreSQL 17 passes the
  postgres, storemigrate and journal suites (81 tests, 0 skipped). Without
  `JOBSHOP_TEST_POSTGRES` those tests skip, so CI-less local runs do not
  prove the Postgres side.
- Restart compatibility: the controller restart, slot, ingest and hook
  recovery tests pass. A controller built at v0.9.22 (before the
  restructuring) started a run with a live turn, and the new binary adopted
  the same host process with no respawn. The run finished on the original
  host. A last-good snapshot written by the old binary loads in the new tree
  and is served by `control.Run` when the pack is broken.

## Found

- `pkg/desk/dist` was committed again by #1126, next to the real, ignored
  `pkg/transport/desk/dist`. Nothing reads it. Removed.
- `jobshop serve` refuses to start on a failing pack even with a last-good
  snapshot, because `startTelemetry` loads config strictly before the
  controller runs. Same on v0.9.22, so not a regression. Filed as #1128.

## Testing restarts

The host environment hash includes volatile shell variables such as `_`. Pin
the environment of both binaries, or a restart test shows a respawn that is
not a regression.
