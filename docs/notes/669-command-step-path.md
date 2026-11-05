# Command steps find jobshop (#669)

2026-11-05.

- Command steps (workflow `type: command`) now get the controller's own
  binary in their env, the same way sessions do. `JOBSHOP_BIN` is the
  binary's path (`/proc/self/exe`, or `Options.HostExe` in tests), and
  its directory comes first on `PATH`. Before this, a step inherited
  the unit's PATH, which has no jobshop, so selfdev's `live-check` step
  failed with `jobshop: command not found` (run js-00339h6z98002).
- selfdev `live-check` calls `"${JOBSHOP_BIN:-jobshop}" runbook run`.
- The empty pool in the escalation was not a lost pool. Controller-run
  steps have no ticket pool on purpose, so no session claims them
  (#396), and placement still uses the workflow step's pool. The
  escalation now prints `command step, run by the controller` for them
  instead of `pool ` with nothing after it.

Test: `TestCommandStepFindsJobshop` (pkg/control) runs `jobshop
version` and `"$JOBSHOP_BIN" version` from a step while the
controller's PATH holds only sh's directory.
