# Pool scaling (#281)

2026-11-04.

- Each tick the supervisor computes `desired` per ticket pool and
  reconciles toward it (`pkg/control/scale.go`). Auto:
  `clamp(min_sessions, max_sessions, ready + held)`, held at its peak over
  `scale_down_after` (default 10m). Scale up is immediate; a reload that
  lowers `max_sessions` clamps at once.
- `jobshop pool scale <pool> <n|auto>` sets a manual target (capped by
  max). `drain --pool` makes desired 0. Both are saved in
  `<root>/scale.json`, so a controller restart keeps them (drain used to
  be memory only).
- Surplus sessions get a `scale` cordon (`fields.cordoned` with reason,
  who and when), picked idle first, then least recent claim, then oldest.
  A cordoned session claims nothing new (runner and `claim --next`),
  finishes its held ticket, then ends `scaled_down`. A shortfall lifts
  scale cordons before starting new sessions.
- `session cordon|uncordon|drain <id>`: a manual cordon keeps the session
  open and is never lifted by the scaler; drain ends it `drained` when
  idle. `assign` to a cordoned session needs `--force`.
- Warm pool: idle sessions outlive `idle_timeout` while the sessions in
  rotation are at or below `min_sessions` (or the manual target).
- Backoff, crash-loop pause and quarantine block scale up only.
  `limits.max_sessions` caps open sessions across ticket pools; starts go
  to lower pool `priority` first.
- `pool list` shows min, max, target, desired, open, working, idle and
  cordoned. Events: `kind: pool` on `events.subscribe`. Metrics:
  `jobshop_pool_sessions_desired`, `jobshop_pool_sessions_cordoned`,
  `jobshop_pool_scale_events_total{reason}`. Grants: `pool-scale` and
  `session-cordon` (the operator pack has both).
- Desk buttons are a follow-up: the UI build needs Node, which the worker
  host lacks.
