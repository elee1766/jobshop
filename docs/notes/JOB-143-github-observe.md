# GitHub observation out of pkg/control (JOB-143, 2026-11-08)

Part of JOB-139. No behavior change.

## What moved to pkg/github (observe.go)

- Read-only types: `PRInfo`, `Review`, `Check`, `IssueInfo`,
  `WorkflowRun`, `Thread`, `Comment`.
- `CLI`, the gh-CLI reader the wait pump and subscription poller use by
  default: `PR`, `Reviews`, `Checks`, `Tag`, `Contains`, `Issue`,
  `WorkflowRuns` and `Comments`. It keeps the 2 minute bound per gh call,
  the exec span names, `--paginate`/`--slurp` paging and the "no checks
  reported" and 404 handling.
- Pure helpers: the `Parse*` functions for each response, `CheckBucket`,
  `NewestRun` (workflow name/file match, JOB-88), `CompareContains`,
  `NeedsRebase` (the mergeable_state rule of #511), `PRRefs` (PR URLs in
  comment text) and `RefNumber` (issue/PR number from ticket meta).

## What stays in pkg/control

- `waitSource` (wait.go) and `threadSource` (subscribe.go) are still
  consumer-owned interfaces; `github.CLI` satisfies the first, and
  `ghSource` is a thin adapter from `github.Comment` to the controller's
  `ghComment`, which also carries Linear fields.
- Per-pass caching (`waitPass`), exact-head review and check evaluation,
  wait clearing and invalidation, follower selection, delivery, reopen,
  `ignored`/`Self` filtering, and the rebase decision. Every write and
  transition is still the controller's.
- `revertSource`/`ghPR` and `liveActivity` (REST client path through
  `apiclients.go` identity scopes) are the activity child's (JOB-149);
  `ghcall.go` too. They already use `github.Client`.

## Coordination

- JOB-144 (waits) and JOB-149 (activity) should use `pkg/github` types and
  not add new PR/review/check types. Two readers exist on purpose: `CLI`
  uses the host `gh` token, `Client` uses identity tokens. Moving waits
  onto `Client` would change token scope and rate-limit behavior, so it is
  not done here.
