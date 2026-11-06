# VictoriaLogs tests behind a build tag

2026-11-06. Tests that need a real `victoria-logs` build only with
`-tags victorialogs`, like the Postgres suite has its own CI job.

- `pkg/journal/vltest` carries the tag, so a test that imports it without
  the tag does not compile. The tag cannot be forgotten silently.
- Tagged tests: `pkg/control/journalship_vl_internal_test.go`.
  `TestJournalSearchNotConfigured` needs no VictoriaLogs and stays untagged.
- CI: the `go` job no longer downloads VictoriaLogs and does not see the
  tagged tests. The `victorialogs` job installs the release, runs
  `go vet -tags victorialogs`, then only the tests in tagged files, and
  fails on any skip. golangci-lint runs with `--build-tags=victorialogs`.
- Locally: `VICTORIALOGS_BIN=/path/to/victoria-logs go test -tags victorialogs -run 'TestJournal' ./pkg/control/`.

Longer term, journal shipping and search should be tested through the RPC
API instead of `pkg/control` internals, so the tagged tests can live in
their own package (JOB-34 direction).
