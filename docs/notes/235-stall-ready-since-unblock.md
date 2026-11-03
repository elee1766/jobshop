# ready_unclaimed waits from the last blocker close (#235)

2026-11-03.

- `ready_unclaimed` aged a ticket from its own creation, history,
  comments and defer time. A step that waited behind a `blocks` edge
  became ready when its blocker closed, with nothing written on the step,
  so it looked unclaimed for an hour or more at once. Four false
  stall-check runs on gasserver on 2026-11-03 (#56, #187, #201 steps).
- Now the newest `closed_at` of tickets it waits on (`blocks` and
  `waits_for` edges from it) also counts. The defer-time rule is kept.
- Only `ready_unclaimed` uses it; other signals are about tickets that
  are already running.
- Test: `TestStallReadyUnclaimedSinceUnblock`.
