# SQLite store: quiet tick latency (JOB-191)

2026-11-10. A quiet supervisor tick took about 1s on selfdev (SQLite, 1.1 GB,
`records` 907 MB) and 0.13s on jobcity (Postgres). Four causes, one change each.

## What changed

- **Statistics.** `sqlite.Open` runs `PRAGMA optimize=0x10002` once, and the
  controller repeats it every 6h in the tick step `maintain_store`
  (`Store.Maintain`, a no-op on Postgres). Production had no `sqlite_stat1`, so
  `LastFire` for a trigger with no accepted invocation read the whole
  collection through `records_created`. A full analysis, not `analysis_limit`:
  with a sampled one the planner went back to `records_created`. After the
  first run it analyzes only what changed a lot.
- **Claim.** `Collection.Claim` finds its candidate with one query per state,
  not `state IN (a, b)`. With one state SQLite uses `records_state_created`
  and reads only that state's rows (0.05ms), where the `IN` form walked
  `records_created` through every terminal row (200ms), and `ANALYZE` did not
  change that. The best of the per-state candidates is claimed, so the order
  is still oldest first. Same shape on Postgres: its plan was already an
  index on `state`, and the split keeps it so with several states.
- **Reader pool.** `Get`, `List`, `Count` and `Search` of a collection read on
  the 4-connection reader pool (`docs/notes/store-read-pool.md`), so they no
  longer queue behind the writer's claim and kick. Reads inside a write
  transaction still use the transaction. A reader that gets SQLITE_SCHEMA
  while a collection registers its FTS5 table repeats, as `Recall` does.
- **`evaluate`.** `EditDue` checks age first with no query, then asks
  `Store.CountRuns` (one `COUNT(DISTINCT parent_id)` over steps) instead of
  listing every step and learning since the merge. `learnEvaluate` calls
  `MeasureEdit` only for an edit that is due. The decision is unchanged
  (`TestCountEditRunsMatchesMeasure` compares the SQL count with
  `EditMetrics.Runs` on the same data). The selfdev `evaluate` trigger gets
  `fallback: "10m"`: `pollDue` already honors `fallback` without `on:`.

## Numbers

A copy of the production database (2026-11-10, modernc v1.61.1, idle
connection, median of 7):

| Query | Before | After |
|---|---|---|
| Claim candidate, `hook_invocations`, none claimable | 197ms | 0.28ms |
| `LastFire` query, trigger with no accepted row | 301ms | 0.15ms |
| `evaluate` run count, 524 steps in 72h | 65ms (`Measure`) | 8ms (`CountRuns`) |

`PRAGMA optimize=0x10002` took 1.6 to 2.4s on that file the first time, at
`Open`. The controller starts once, so this is a startup cost, and the
6-hourly repeat costs microseconds when little changed.

Quiet-tick numbers from selfdev after deploy are read as in
`docs/verify/tick-latency.md` and reported on JOB-191.

## Not done

The 7.3s tick at 10:00:33 UTC (every store call at about 90ms or a multiple)
is not explained by these changes. Do not call it fixed until a tick with that
pattern is traced again.
