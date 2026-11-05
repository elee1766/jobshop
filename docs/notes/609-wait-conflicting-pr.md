# CI waits on a conflicting PR (#609)

2026-11-05. GitHub runs no `pull_request` CI on a PR that conflicts with
its base, so a `github.checks:pr=...` wait for success or failure never
ended and sat until its timeout.

- The wait evaluator reads `mergeable` and `mergeable_state` from the PR
  fetch it already makes each pass. `mergeable: false` with state
  `dirty` clears every checks condition on that PR as invalidated:
  `PR #<n> conflicts with <base>; CI will not run`. With `--resume` the
  parking session gets the step back. `mergeable: null` (GitHub has not
  computed it) is read again next pass.
- The implement prompt says: on that reason, rebase on `origin/<base>`,
  push `--force-with-lease`, set the wait again.
- A `github.checks:repo=...,sha=...` wait that has seen no check at all
  15 minutes after it started is listed in attention (kind `wait`) as
  `CI did not start for <sha>`, before its timeout. It drops out once a
  check appears.
