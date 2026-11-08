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

Removing those did not move pkg/control on the CI runner (51.6s on the
PR run, no test over 3.6s). The runner has 2 vCPUs and the package is
CPU bound: the summed test time is ~127s, and the profile puts ~40% of
CPU in race-instrumented sqlite C (`modernc.org/sqlite/lib`: parsing
and preparing statements). Building the pkg/control test binary with
`-gcflags=modernc.org/sqlite/lib=-race=false` cut user CPU from 3m04s to
2m04s on this host.

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
- CI and ci-flaky pass `-gcflags=modernc.org/sqlite/lib=-race=false`,
  next to the checkptr flag from #433. Tradeoff: the race detector no
  longer instruments sqlite's translated C. That code is single
  threaded behind sqlite's own mutexes, and the Go wrapper and all our
  packages stay instrumented.
- `TestHookCapabilitiesOutsideControllerLocksAndCancelDiscardsActions`
  uses a 1m hook timeout: with more tests in parallel the 10s wall
  clock timeout could fire before cancel and end the hook `failed`.

## After (CI, PR #1060)

| package | master 42468831522 | PR, tests only (34a9910) | PR, with sqlite flag (58921e3) |
| --- | --- | --- | --- |
| pkg/control | 51.8s | 51.6s | 31.0s |
| pkg/store/sqlite | not in top 5 | 12.2s | 4.2s |
| pkg/store/postgres | 8.9s | 8.7s | 7.6s |
| pkg/cli | 5.3s | 1.6s | 1.2s |
| go test step | 64s | 76s | 64s |

The PR step times probably include a cold build (not verified): the PR
cache is read-only (only master commits to it), and the new `-gcflags`
for `modernc.org/sqlite/lib` changes the build of every package that
imports it under `-race`. The `go build` step took 8s on the PR runs
against 4s on the master run. Package test times are not affected by
this, the step time may be. Judge the acceptance (three consecutive
master steps under 60s) on master runs after the first one post-merge.
Check with `gh run view <id> --json jobs` (step `go test`, started to
completed). The step was 64s on the PR run, so it is not yet shown to be
under 60s.

If it is still over 60s: the next cost in pkg/control is sqlite statement
parsing (~18% of CPU), from stores opened and queries prepared per test,
then the race runtime itself. See `pkg/control` CPU profile with
`go test -race -c` and `-test.cpuprofile`.
