# replan counts ready work in its own instance (#1221)

2026-11-11.

- `tickets()` takes `ready=True`, mapped to `store.Filter.Ready`: open,
  not deferred, waiting or blocked, with a free resource slot. It is the
  same rule as a trigger's `ready_below`.
- `tickets()`, `runs()` and `sessions()` called from a trigger or
  workflow callback map `pool` and `workflow` to the declaring pack
  instance's prefixed name when the instance has one, as `create_ticket`
  and `start_run` already did. Other names pass through unchanged: reads
  still see the whole store. Guard hooks, desk commands and signal
  scripts have no owner and read names as given.
- The github pack's `replan` uses `tickets(pool=dev_pool, ready=True)`.
  Before, an open ticket deferred for days in the unprefixed `dev` pool
  stopped every instance's replan from filing "Plan next step".
