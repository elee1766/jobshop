# Resource waits (#408)

2026-11-04. A step held back only by a full resource (open, unblocked,
not deferred, not parked on a wait, every slot of one of its keys taken)
is out of the ready frontier. It is now visible:

- `store.ResourceWaits` lists such tickets in claim order per key
  (priority, created_at, id), with how many waiters are ahead, held and
  capacity. A ticket naming several full keys is listed under the first.
- `run show` prints `waiting on <key>, <n> ahead (<held>/<cap> held)` on
  the step's resources line; JSON has `waits`.
- Stall signal `resource_wait` (store signal, needs `after`): fires for
  a waiter idle that long, counted from its last activity or its last
  blocker closing.
- `attention.resource_wait: <duration>` lists such waiters in
  `jobshop attention` (kind `resource`). Unset: off.
