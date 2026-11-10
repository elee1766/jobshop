# Tick latency (JOB-191)

How to read how long a supervisor tick takes, and what a quiet one costs.

1. **Whole tick.** `GET /api/system` (Desk) or `system.status` returns
   `tick_seconds` (the last pass), `p95_seconds` and `samples` (the last 60
   passes), and `steps`: seconds per tick step for the last pass. The same
   step times are the histogram `jobshop.tick.step.duration{step}`, and each
   step is a span `tick <step>` under `supervisor tick`.
2. **Which step.** Sort `steps`. A quiet instance should have `run_triggers`,
   `supervise` and `sweep_attention_chats` each in the low tens of ms. Over
   100ms: open the `tick <step>` span in the trace and look at its `store`
   children (`store.SlowCall` spans, 50ms or more).
3. **A quiet tick.** Nothing ready, no pending edits due. Expect
   `tick_seconds` under 0.1s on SQLite and Postgres alike (jobcity: 0.13s,
   selfdev before JOB-191: 0.7s to 1.9s). Before JOB-191 on selfdev,
   `run_triggers` p50 was 614ms (`trigger evaluate` 317ms) and `supervise`
   127ms.
4. **Statistics.** On SQLite, `select count(*) from sqlite_stat1` on a copy
   of the database is non-zero. `maintain_store` is the step that refreshes
   it every 6h and is a few microseconds when idle.
5. **Planner check.** On a copy of the database, `EXPLAIN QUERY PLAN` of the
   claim query for `hook_invocations` with one state names
   `records_state_created`, not `records_created`; `LastFire` for a trigger
   name that never fired names `ri_hook_invocations_trigger`.
6. **A slow tick that is not these.** A tick where every store call, even a
   4ms reader-pool list, takes about 90ms or a multiple is a different
   problem (JOB-191, not explained). Capture the trace of that tick and
   `/proc/pressure/cpu` before changing anything.
