# VictoriaLogs tests behind a build tag

2026-11-06. Tests that need a real `victoria-logs` build only with
`-tags victorialogs`, like the Postgres suite has its own CI job.

- `pkg/journal/vltest` carries the tag, so a test that imports it without
  the tag does not compile. The tag cannot be forgotten silently.
- Tagged tests: `pkg/control/journalvl` (JOB-47 moved them out of
  `pkg/control` onto the RPC API, see JOB-47-rpc-tests.md).
  `TestJournalSearchNotConfigured` needs no VictoriaLogs and stays untagged.
- CI: the `go` job no longer downloads VictoriaLogs and does not see the
  tagged tests. The `victorialogs` job installs the release, runs
  `go vet -tags victorialogs`, then `go test` on that package, and
  fails on any skip. golangci-lint runs with `--build-tags=victorialogs`.
- Locally: `VICTORIALOGS_BIN=/path/to/victoria-logs go test -tags victorialogs ./pkg/control/journalvl/`.
