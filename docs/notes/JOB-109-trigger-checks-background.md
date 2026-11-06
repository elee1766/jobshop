# Trigger checks run outside the tick (JOB-109)

2026-11-06.

A trigger `check` (and a `when.stall` signal with `check`) is a shell
command with `check_timeout` up to minutes. `runTriggers` ran them one
after another inside the pass, under `tickMu`. On jobcity two review
checks near their 2m timeout made `supervisor tick` 55s, delayed supervise,
provider_pass and the sweeps, and blocked the webhook and event passes
(JOB-104 already stopped reload waiting on the tick lock).

As built (`pkg/control/triggercheck.go`):

- A pass evaluates a trigger as before, with a `checkPass` on its ctx.
  Where evaluation would run a check, `runCheck` notes the command and
  returns `errCheckPending`: nothing fires, nothing is logged as an error.
  Every condition before the check (dedup, cooldown, cron, counts,
  learnings, edits) is still evaluated in the pass, so the check runs
  only when they hold.
- The commands then run in a goroutine, outside `tickMu`, at most 8 at
  once across triggers (`maxTriggerChecks`). Checks of one stall trigger
  run at the same time. The same command and timeout is one run.
- When they end, the goroutine takes the tick lock, reads the config
  again, and evaluates the trigger a second time with the results, so
  the conditions are read after the wait, not before it. Then it fires
  as a pass would.
- One trigger has one check job. A poll that finds it running skips the
  trigger (its `fallback` timer is not stamped). An event keeps only the
  newest event for it while the check runs, and runs it after: its
  fields are the check's input.
- The trigger fires when its check ends (the replay above), not in the
  pass that reached it, and a check slower than the tick interval no
  longer delays anything but its own trigger.
- Direct calls in tests (a `Controller` built without `Run`) keep the
  check inline, one synchronous evaluation. `trigger.list` still runs
  checks inline and in parallel (#522).
- Tracing (the issue's addendum). Both background paths run on a ctx
  with no span, so they now start a root span: `trigger checks <name>`
  per check round (its commands, its wait for `tickMu` and the replay
  nest under it) and `trigger event <names>` per event job (attribute
  `jobshop.triggers`). `lockTick` keeps who holds `tickMu` (kind,
  trigger, span context) under a small mutex, and `wait tickMu` records
  it: `jobshop.lock.for` is the holder, `jobshop.lock.holder.trigger` its
  trigger, `jobshop.lock.waiter` the waiter, and a span link points at
  the holder's span. Before this `lock.for` was the waiter. The holder
  is the one at the moment the wait began; a wait behind several jobs in
  a row names the first.
