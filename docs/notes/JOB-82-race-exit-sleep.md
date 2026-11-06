# pkg/control child processes skip the race exit sleep (JOB-82, 2026-11-06)

- The race runtime sleeps `atexit_sleep_ms` (default 1000) before a
  process exits. Every host, agent, command step and helper a
  pkg/control test spawns is the race-built test binary, so each child
  that exited held its test for an extra second: a host stop, an agent
  turn ending, a restart. That second was most of a `poolRootReal`
  test's runtime, not controller startup.
- `TestMain` now calls `noRaceExitSleep` (`race_test.go`), which adds
  `atexit_sleep_ms=0` to `GORACE` before any test runs. Children inherit
  the env. A `GORACE` that already sets it is left alone. A build
  without -race ignores the variable.
- Race reports are not lost: a child's report is written when the race
  is found. Only the wait at exit is skipped.

## Measured (CI flags: -race, checkptr off for modernc, host load 10-16)

- One `poolRootReal` test run alone, after the first run in the process
  (the first pays about 0.5s once to build the schema template):
  TestRateLimitBackoff 1.44s/1.48s before, 0.43s/0.44s after.
  TestControllerRestart 1.31s/1.33s before, 0.30s/0.32s after.
- Whole package, same binary, `GORACE=atexit_sleep_ms=1000` vs unset:
  31.1s and 51.5s before, 28.2s and 28.5s after. Package wall time is
  CPU bound (about 150s CPU on 6 cores) and noisy under host load, so it
  moves less than the per-test numbers.

## Not changed

Controller startup itself is about 0.1s per test once the schema
template exists: the per-database cost left is SQLite parsing the
schema on each new connection (goose's version check, the reader pool)
and closing them, which is SQLite's own work.
