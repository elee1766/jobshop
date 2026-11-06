# Slice changes under home-manager keep units running

2026-11-06, JOB-96.

A home-manager switch that changed only `memoryHigh`/`memoryMax` stopped
`jobshop.slice`. sd-switch stop-starts a changed active unit, and
stopping a slice stops every unit in it. Only the slice was started
again, so the controller, Desk and ingest stayed down.

- The home-manager slice unit sets `X-SwitchMethod=keep-old`. The
  switch's daemon-reload applies the new `MemoryHigh`, `MemoryMax` and
  `CPUQuota` to the running slice (checked on a scratch user slice:
  `memory.high` changed, the unit in it kept running). NixOS's switch
  never restarts slices, so the system module needs nothing.
- `jobshop-shutdown`'s ExecStop runs `jobshop shutdown` only when the
  service manager is not `running` or `degraded`, that is when it is
  going down. A stop during a switch or by hand leaves sessions running;
  a full stop stays `jobshop shutdown`.
