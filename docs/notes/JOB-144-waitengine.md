# pkg/control split: wait evaluation and wake planning (JOB-144, 2026-11-08)

Part of JOB-139. No behavior change.

## What moved

`pkg/control/waitengine` decides what a wait (`fields.wait`) is waiting
for. It writes nothing.

- `Evaluate(ctx, pass, wait, now) Decision`: all/any, unmet notes, time
  conditions, PR head binding, head-move/merge/close/cancel invalidation,
  timeout detection. It takes the wait by value and returns the new one.
- `Pass`: per-pass cache of provider reads (one PR fetch per pass).
- `Source`, `LinearSource`, `RunSource`, `ReplySource`: what a pass
  observes through. `Source` is satisfied by `github.CLI` and uses the
  `pkg/github` value types (JOB-143). The Linear client adapter, run
  tickets and the reply filters (bots, ignore, authors) stay in
  `pkg/control`; the engine imports no provider client.
- `Wakes`, `CondSubjects`, `ItemSubjects`, `Names`, `NextDue`: targeted
  wake subjects and the next time/timeout deadline.
- `TimerHint`: the `release --until` guard text (was `waitguard.go`).

## What stays in the controller

`wait.go` and `waitwake.go` keep `waitAdd`/`waitList`/`waitClear`
(authorization, scope), `waitMu`, listing waiting tickets, `clearWait`
(comment, label removal, resume holder, dispatch eligibility), the
revision-checked save, the timeout label/comment/chat hold, and the pump
goroutine with its timers.

## Stale observations

A `Decision` is advice computed from the ticket snapshot the pass listed.
The controller applies it with `IfRevision` of that snapshot, so a wait
cleared, re-parked or replaced since (by hand, a release, an agent) fails
with `store.ErrConflict` and the next pass or wake evaluates the current
wait. The pass is created after the list, so cached reads are newer than
the snapshots. `Wakes.Take` runs before the list; a wake that arrives
during a pass stays queued for the next one. Covered by
`TestWaitStaleSnapshot` and `waitengine` unit tests, on fake clocks.

## Remaining

`subscribe.go` and `watch.go` (thread subscriptions) still share
`threadSource` and the comment filters with `thread.reply`; they move with
the triggers extraction.
