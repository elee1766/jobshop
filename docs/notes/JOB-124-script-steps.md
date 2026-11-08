# Starlark step functions (JOB-124)

2026-11-07.

- `start_step(script=fn, pool=...)` makes a script step: a Starlark
  function the controller runs in-process through the command-step
  path (claim, lease, retries, restart, escalation). Reference:
  `docs/configuration/callbacks.md#script-steps`.
- Inside it `shell()` goes through `runIsolated` with the pool's
  sandbox and the run worktree (`api.Options.Exec`). `http` and `jev`
  keep the service grants. No `jobshop step exec` process and no code
  mounted into the sandbox: the run's pinned program is already loaded.
- A script step runs only while its run's callback pin still matches.
  Its fields record `script`, `script_pool` and `script_hook` (the
  callback whose modules hold the function).
- Attempts are journaled as hook invocations (`step_*` states) with
  their calls; results are never replayed.
- Selfdev `live-check` uses it under `starlark_workflows`. Not yet on
  `start_step`: `needs_resources`, `retry_after`, `placement`,
  `max_review_rounds`, an explicit `on_request_changes` (Jobcity needs
  them; added with its migration).
