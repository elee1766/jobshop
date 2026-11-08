# CI go test time (JOB-117)

2026-11-08. Goal: the CI `go test -race ./...` step on master under 60s
(it took 64-119s in early November).

## Where the time went

Before, master CI run 42468831522 (step 64s, compile included):

| package | before |
| --- | --- |
| pkg/control | 51.8s |
| pkg/store/postgres | 8.9s |
| pkg/cli | 5.3s |
| pkg/config | 3.6s |
| packs | 3.3s |

pkg/control sets the step time. Its serial phase (tests without
`t.Parallel`) was about 23s, and inside it a few tests seeded hundreds
of rows one RPC at a time:

| test | before | after (local, alone) |
| --- | --- | --- |
| TestHookPublicInventoryIsBoundedBeyondOnePage | 14-16s | 0.7s |
| TestSessionListBeyondDefaultPage | 11s | 2.5-4s |
| TestTicketGraph | 4-12s | 3s |
| TestInboxPruneBatches (pkg/ingest) | 4-7s | <0.1s |
| TestCLIDoc (pkg/cli) | 3.5-8.6s | 0.1s |

## What changed

- Bulk seeding: `seedStore` (pkg/control/seed_test.go) opens the test
  store while no controller runs, so a test writes 100-500 rows in one
  `Apply` instead of one RPC each. The hook inventory test runs one
  real hook, then copies its audit record.
- The remaining serial pkg/control tests that touch no process-wide
  state (env, cwd, package vars) now call `t.Parallel()`.
- pkg/ingest: the inbox prune batch size is a var a test can shrink
  (`SetInboxPruneBatch`), so the batching test prunes 5 rows in batches
  of 2, not thousands.
- pkg/cli: `cliDoc` builds one kong parser and traces each command path
  instead of building a parser per command.
- CI sets `GORACE=atexit_sleep_ms=0`. The race runtime otherwise sleeps
  1s before every race-built process exits, once per test binary and
  per child process a test runs, about 34 package-seconds a run.
- `TestHookCapabilitiesOutsideControllerLocksAndCancelDiscardsActions`
  uses a 1m hook timeout: with more tests in parallel the 10s wall
  clock timeout could fire before cancel and end the hook `failed`.

After: see the PR's CI run and the next master runs on
[JOB-117](https://linear.app/gfx-labs/issue/JOB-117).
