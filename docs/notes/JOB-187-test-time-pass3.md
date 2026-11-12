# CI go test time, third pass (JOB-187)

2026-11-12. The `go test` step was back at 89-116s on master after the
second pass (`docs/notes/JOB-187-test-time.md`) brought control to 40s.

## Cause

Two things, measured separately.

1. The runner changed. #1162 and #1163 moved the go job from Namespace
   2x4 (2 vCPUs) to `talos-metal-gfx-public`: 24-32 CPUs, no CPU quota
   (`cpu.max` is `max`, zero throttled periods), load 3-6 when the step
   starts. Control ran 40-43s on Namespace and 56-74s on talos-metal at
   the same commit range (master runs 42759505869 to 42786922733).
   Control's run time does not scale with cores. A sweep on the runner,
   one control process, same binary:

   | `-test.cpu` / `-parallel` | control |
   | --- | --- |
   | 2 | 69s |
   | 4 | 62s |
   | 6 | 60s |
   | 8 | 66s |
   | 32 | 76s |

   With 24-32 Ps, Go ran ~38 control tests at once (12 on Namespace) and
   each slowed: `TestTicketGraph` 1.6s to 53s. The mutex profile shows
   why: half the contention is `modernc.org/libc` `allocMu`, one lock for
   every SQLite allocation in the process, and SQLite statement
   preparation. More Ps in one process only add waiters.

2. Code grew. On 2 pinned local cores, control at f219ae9 took 39s and
   master 56s (total CPU 59s to 81s). 57 new tests add 16s serial; the
   rest is per-test cost. The largest single new cost was schema reloads:
   `sqlite3Reprepare` 1.6s to 3.9s, `sqlite3InitOne` 2.7s to 4.7s. Every
   fresh database changed its schema after connections had prepared
   statements, so each pooled connection reloaded the schema:
   - `Open` runs `PRAGMA optimize` (JOB-191), which on a fresh copy of
     the template creates `sqlite_stat1`/`sqlite_stat4` (schema version
     130 to 131).
   - Registering the hook collections created five indexes (131 to 134).

## Changes

- The template database runs `PRAGMA optimize` before it is copied, so
  `Open` on a fresh copy finds nothing due.
- Migration 00029 creates the hook collections' indexes, as 00022-00024
  did for learnings, inbox and journals. Registration then runs no DDL.
- `.github/scripts/go-test.sh` runs every package but control in one
  `go test`, and control's test binary (built once) as five processes,
  slices by test name; the last slice skips what the others select, so
  each test runs once (checked: 457 of 457). Output is the same
  `go test -json` stream, so ci-flaky and the skip check are unchanged.

Local, 2 pinned cores: control 57.3s to 51.6s, CPU 81s to 74s,
`sqlite3Reprepare` 3.9s to 1.2s.

## Result

| run | step | control |
| --- | --- | --- |
| master 43112373165 (before) | 89s | 83.1s |
| master 43111377359 (before) | 116s | 84.1s |
| PR #1226, 43123897909 | 38s | slices 10-20s |

Slices on the runner: A-D 13s, E-L 20s, M-Q 10s, R-S 18s, rest 12s.
The step is now the build plus the slowest of control's slices and
sqlite (17s), postgres (12s), packs (16s).

archcheck stays at 1.6s.

## Build versus run

The PR run restored master's cache. Build and vet took 8s and 6s; the
step's 38s includes compiling the changed packages. A cold cache adds
the build of every race package (2m24s on 2 cores locally in the second
pass; less on 24 cores).

## Next cost

If the step creeps back: the slowest slice (E-L, 20s) decides it, so
rebalance the name ranges first. Then `modernc.org/libc`'s global
allocator lock, which caps one process whatever the core count. The
test count itself grows about 14% in two days.

Acceptance (three master runs under 60s) is listed on JOB-187 after merge.
