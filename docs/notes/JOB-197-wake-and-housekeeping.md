# Wake idle sessions on ready work; housekeeping off the pass (JOB-197)

2026-11-10.

## Wake instead of poll

- Every idle runner used to call `ClaimNext` each `Poll` (5s): a new ticket
  waited up to 5s and an idle pool cost steady queries.
- Now each pool has a wake channel (`readywake.go`). A wake closes it and
  drops it, so it reaches every runner waiting on the pool; the next
  `readyChan` makes a fresh one. A runner takes the channel at the top of
  its loop, before it reads messages or claims, so work committed after
  that read ends the wait at once instead of falling between the read and
  the wait.
- Who wakes: the events pump (`pumpEvents`) already reads every committed
  history row once, so `noteChange` runs there and no writer needs to know
  about runners. An open task or step woken by `create`, `release`,
  `handoff`, `lease_expired`, `reopen`, `update` (wait cleared, deferral
  lifted, pool or holder set) or `edge_remove` wakes its pool. Work
  pre-assigned to a session wakes that session's pool. A `close` wakes
  every pool (a blocker or a resource slot freed can be in another). A new
  message wakes the pool it names, its session's pool, or all pools for a
  key. The provider pass wakes all pools (a slot can turn ready with no
  ticket write).
- `Options.Poll` is only the fallback now, default 30s (was 5s), for what
  sends no wake: a `defer_until` running out.
- Test: `TestIdleSessionWakesOnNewTicket` runs with `Poll` of one hour on a
  fake clock that nothing advances, and no pass runs while it waits, so only
  the wake can deliver the ticket. `TestReadyWake` and `TestNoteChangeWakes`
  cover the channel and which rows wake.

## One query per pass

- `supervise` reads every in-progress ticket once (`inProgress`) and uses it
  for the heartbeats, the session gauges (`recordGauges`) and the scale plan
  (`reconcile` -> `scaleSessions`). It used to list per live runner, per open
  session in `recordGauges` and per open session in `scaleSessions`.
- If that read fails the pass skips scaling (an empty held set would look
  like every session idle and scale a busy pool down).
- `keepWarm` and `poolStatusCtx` run outside the pass and read the set
  themselves (nil `held`).
- Test: `TestSuperviseQueriesDoNotGrowWithSessions` counts `List` calls of
  one `supervise` with 2 and with 20 sessions: equal.

## Housekeeping

- `prune_inbox`, `prune_write_keys`, `prune_sync`, `prune_artifacts`,
  `compress_journals` and `maintain_store` left the tick. They run in
  `housekeeping.go`: a goroutine at start and every minute, without `tickMu`,
  serialized by `housekeepMu` (which owns the `*At` interval fields). Each
  still has its own interval guard (1h, 10m, 6h ...).
- A controller with `Options.Pass` does not start the loop. A requested pass
  runs `housekeeping()` after the tick, so tests drive everything with
  `pass`.
- Telemetry: `jobshop.housekeeping.duration{step}` and a `housekeeping` root
  span with one `housekeeping <step>` child each. `jobshop.tick.step.duration`
  and the system status `steps` no longer carry these six.
