# Pool crash-loop pause names the error (JOB-27)

2026-11-05. When a pool pauses after `poolCrashes` crashed sessions in
`poolCrashWindow`, the controller keeps the last session's end reason
(the spawn or crash error) and, for an isolated pool, the failed checks
of `host.Preflight` with their `fix:` hints (#728). That text appears in:

- the controller log line `pool quarantined: sessions crash-looping` (`err`),
- `jobshop attention`: a `pool` item "pool <p> paused until <t>, sessions
  crash-looping: <error>" while the pause lasts, unless an open escalation
  ticket stands for the pool,
- the `pool_quarantined` stall detail, so stall-check's `{{.detail}}` names
  the error and diagnose does not need the incidents.

The reason is memory only, like the pause, and is cleared with it by
`pool unquarantine`.
