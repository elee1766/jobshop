# CI flaky rerun (#510)

2026-11-05. When `go test` fails in the `go` job of `.github/workflows/ci.yml`,
the step "rerun failed tests" runs `pkg/flaky/cmd/ci-flaky` on the
`go test -json` log (`test.json`). It reruns each failed top-level test with
`-run '^Name$' -count=20 -race` and gives a verdict:

- `flaky`: at least one rerun passed.
- `broken`: no rerun passed.

Output:

- Artifact `failed-tests`, file `failed-tests.json`: a JSON array of
  `{"test", "package", "verdict"}`. The planner's incident step reads it with
  `gh run download <run-id> -n failed-tests`.
- A table in the job summary (`$GITHUB_STEP_SUMMARY`).

The job still fails either way. The test step prints the plain log from the
JSON stream (`ci-flaky -print`), so the log reads as before.
