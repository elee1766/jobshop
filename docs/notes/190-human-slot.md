# Reserved human slot per pool (#190)

2026-11-03.

- The supervisor starts sessions for ready work up to `max_sessions`.
  Human-started sessions get one slot over that: `session.send` by pool
  and key (operator chat, gateway threads, escalation threads) and
  `session.new` with a pool start a session while the pool has fewer
  than `max_sessions + 1` open, so a pool full of ticket work never
  blocks a human from reaching a session.
- At `max_sessions + 1` open, a new key is refused with a conflict error
  naming the pool. Sending to a key that already has an open session
  always works.
- Pools with `max_sessions: 0` are manual-only and stay uncapped for
  humans, as before.
- The supervisor never uses the slot; it still stops at `max_sessions`.
- Test: `TestHumanSlot`.
