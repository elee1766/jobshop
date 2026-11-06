# Store benchmarks per engine (JOB-43, 2026-11-06)

`storetest.Bench` runs the store hot paths against one seeded store. Each
engine has a `BenchmarkStore`. They are not run in CI (`go test ./...`
runs no benchmarks).

    go test ./pkg/store/sqlite/ -run '^$' -bench Store
    JOBSHOP_TEST_POSTGRES=postgres://... go test ./pkg/store/postgres/ -run '^$' -bench Store

Seed (`JOBSHOP_BENCH_SESSIONS`, default 1000, scales all): 1000 sessions,
1000 runs with 3 steps each, 2000 tasks (a third labeled `bug`), 80%
closed, 5 comments on half the tickets: 7000 tickets, ~17500 comments.
This is near gasserver's ticket count, not its 357MB size (no journals,
recall or big bodies).

Cases: List of open sessions, open `bug` tickets, running runs, the ready
queue; Get; ClaimNext plus release with 1, 4 and 16 claimers (ns/op is
per claim, wall time); Mixed: one supervisor pass (ready, sessions, runs
lists, a claim, a comment) against 4 Desk pollers.

Baseline, AMD EPYC 7413, load ~13, file SQLite (WAL, default options),
local Postgres 17.11 over a unix socket, default config:

| case | sqlite | postgres |
|---|---|---|
| ListSessions | 5.8 ms | 1.5 ms |
| ListLabeledOpen | 15.9 ms | 2.3 ms |
| ListRunState | 15.9 ms | 1.9 ms |
| ListReady | 33.3 ms | 3.6 ms |
| Get | 148 us | 89 us |
| ClaimNext, 1 claimer | 32.5 ms | 4.6 ms |
| ClaimNext, 4 claimers | 39.5 ms | 7.5 ms |
| ClaimNext, 16 claimers | 44.5 ms | 12.8 ms |
| Mixed (one pass) | 334 ms, 12 polls | 45 ms, 15 polls |

SQLite's ready query (and ClaimNext, which runs it) is the slowest path;
it grows with closed history, so it is the first place to look.
