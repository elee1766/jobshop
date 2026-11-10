# CI go test time, second pass (JOB-187)

2026-11-10. Goal: the CI `go test -race` step on master under 60s again
(80-170s on the last seven green master runs). Follows JOB-117.

## Cause

`pkg/orchestration/browse` had `var scanner = redact.Builtin()`
(JOB-163). That compiles betterleaks' default rules plus rules.toml in
package init: 343 ms under `-race` (`GODEBUG=inittrace=1`). Every process
that links the package paid it at start. The control tests use their own
test binary as the fake agent (`FAKE_AGENT_BIN`, `control.test`), and every
session spawn, restart and host start runs that binary again. Start time of
`control.test -test.run XXX`:

| binary | start |
| --- | --- |
| JOB-117 commit (ffbb599a) | 0.05s |
| master before this change | 0.43s |
| after | 0.085s |

Serial per-test times (`-test.parallel 1`, same host, same flags as CI),
the tests that spawn the agent:

| test | JOB-117 | master | after |
| --- | --- | --- | --- |
| TestQuarantine | 1.01s | 5.47s | 1.10s |
| TestEnvRefreshLive | 0.52s | 3.06s | 0.57s |
| TestEnvRefreshStopsDaemon | 0.47s | 2.49s | 0.60s |
| TestSessionRestartBusy | 0.52s | 2.53s | 0.55s |
| TestSessionRestartIdle | 0.50s | 2.08s | 0.50s |
| TestPromptRules | 0.51s | 1.82s | 0.57s |
| sum of all ~400 tests | 53s | 95s | 57s |

`pkg/orchestration/control` pinned to 2 cores (`taskset -c 0,1`, one run
each, host load 11-19 so absolute numbers are noisy): 89s before, 41s after.
CI package time before: 63-87s.

`redact.Builtin` is cached, so the fix is to call it where the message is
classified (`classify`), not at init. No other test binary has an init over
20 ms now (inittrace on every package with tests).

## pkg/tooling/archcheck

`repoGraph` ran `go list` over the repo on every call: 30 subtests plus three
tests. Now `loadRepo` is a `sync.OnceValue`, and `repoGraph` returns a copy
(Package values and both import slices are cloned, because the subtests add
edges). The tests and subtests are `t.Parallel`. Package time with `-race`,
local: 16.3s before, 1.5s after. CI before: 9.6-22.6s.

## The 80s to 170s spread

The step time is the build of the test binaries plus the slowest package.
The package times printed in the log start after that package's binary is
built, so the step minus the control package time is mostly build and
scheduling time. From the CI logs of the seven runs:

| run | step | control package | rest | packages `(cached)` |
| --- | --- | --- | --- | --- |
| 42751923752 | 80s | 69.5s | 10s | 45 (release commit) |
| 42751418524 | 102s | 63.3s | 39s | 1 |
| 42751048962 | 109s | 70.2s | 39s | 3 |
| 42750787304 | 113s | 87.3s | 26s | 41 (release commit) |
| 42750283506 | 129s | 83.5s | 45s | 1 |
| 42738764495 | 170s | 66.9s | 103s | 0 (release commit) |
| 42738189613 | 107s | 70.9s | 36s | 12 |

The control package varies by 24s across runs, the rest by 93s. The run with
170s was a release commit that changed no Go source, but nothing was cached:
the Namespace cache volume was cold for it (its `go build` and `go vet`
steps took 9s and 14s, against 3s and 1s in the 80s run). That points at
compile time on a cold or partly cold volume, with no test change involved.

A local check of the size of that compile: `GOMAXPROCS=2 go test -p 2 -race
(CI flags) -run '^$' ./...` with an empty cache took 2m24s, with a warm cache
1s. This is a local number on a loaded host. It is not read from a CI run,
because the step redirects `go test` output to test.json and the log has no
build timing. Any change to the `-gcflags`, the Go version or a dependency
rebuilds every package that imports modernc.org.

So the test run can be brought under 60s (control is about 40s on 2 cores
now), but a commit that lands on a cold or partly cold cache can still exceed
60s from compile time that no test change affects. If three master runs are
not all under 60s after this, compare the `rest` column above first.

## Not changed

- `pkg/storage/store/sqlite` was 8.3s in some runs and under 1s in others:
  the under-1s run reported it `(cached)`. Its uncached time is 8s.
- `pkg/storage/store/postgres` (9-12s) is not touched.
- No test removed, skipped or tagged out. No race coverage change. Runner size
  unchanged.

## Acceptance still open

Three consecutive master `go test` steps under 60s cannot be shown before this
merges. Check after merge with `gh run view <id> --json jobs` (step `go test`,
started to completed) and list the run ids on JOB-187. If a warm-cache source
commit is still over 60s, the next costs are, in order: sqlite statement
parsing in `pkg/storage/store/sqlite` (`cachedDB.QueryContext` prepares
`List` texts that differ per filter shape; ~12% of control CPU in the profile),
then the race runtime itself, then a 4-vCPU runner (the e2e job already uses
one).
