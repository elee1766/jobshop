# Reload does not wait for the pass (JOB-104)

2026-11-06.

Reload took the tick lock (#102) so one pass never mixed two configs,
which made every reload wait for the pass in flight (JOB-102).

As built:

- `tick()` loads the live config once and passes it (`cfg`, and the
  lease to `supervise`) to every pass function. The trigger event path
  (`onEvent`) and the Linear issue webhook load it once per event pass
  the same way.
- Reload no longer takes `tickMu`. It loads, writes resource caps, runs
  the isolation preflight on the new config, quarantines pools that fail
  it, then swaps the pointer and clears the other quarantines. A pass in
  flight finishes on its config; the next pass uses the new one.
- The `config reloaded` log line's `lock_wait` now only measures
  `reloadMu` (one reload at a time).
- RPC handlers and session runners still call `c.config()`: one read
  each, no pass to keep consistent.
- `TestPassReadsConfigOnce` walks the calls from `tick` and `onEvent`
  through Controller methods and fails on a read of `c.config()`,
  `c.live`, `leaseDur`, `scanner` or `redactor`. A function shared with
  RPC code takes `cfg`, or has an `...In(cfg, ...)` form
  (`closeTicketIn`, `linearClientIn`, `sessionCwdIn`, `poolSandboxIn`,
  `mappedHomeDirIn`). Goroutines the pass starts are not walked; a
  command step attempt keeps the config of the pass that started it.
- `isolatedCmd` with a pool needs `cfg` set; without it the command
  fails setup instead of reading the live config.
- `TestReloadDuringPass` holds a pass in a trigger check on a fifo,
  reloads, and checks the reload returns, the held pass fires the old
  trigger, and the next pass fires the new one.
