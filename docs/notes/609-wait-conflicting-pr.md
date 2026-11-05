# CI waits on a conflicting PR (#609)

2026-11-05. GitHub runs no `pull_request` CI on a PR that conflicts
with its base, so a `github.checks:pr=...` wait on it could never be
met and sat until its timeout.

- The wait evaluator reads `mergeable_state` from the PR it already
  fetches each pass. `dirty` clears a checks wait as invalidated with
  `PR #<n> conflicts with <base>; CI will not run`; with `--resume` the
  session gets the step back. `unknown` or empty (GitHub still
  computing) is read again next pass.
- The github pack's implement prompt says to rebase on `origin/<base>`,
  push `--force-with-lease` and set the wait again on that reason.
- A `github.checks:repo=...,sha=...` condition records `started` once it
  sees any check. 15 minutes after the wait started with no check seen,
  `attention` lists the ticket as kind `wait`, titled
  `CI did not start for <sha>`. The wait itself stays until its
  timeout.
