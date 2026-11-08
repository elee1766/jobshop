# Step execution moved to control/stepexec (JOB-157, 2026-11-08)

Follow-up from the JOB-155 audit. No behavior change.

`pkg/control/stepexec` runs one attempt of a command, http or script step
and returns a `Result` (result, output, outcome, command, retry-after,
error, structured output). It holds:

- templating: `LoadTemplates` (run vars, pack vars, earlier steps' outputs),
  `Templates.Fill`, `Env` (`JOBSHOP_VAR_*`, step key, ticket);
- placement: `Executor.Place` (pool, working dir, run worktree), `RunDir`,
  `RemoveRunDir`, `StaleRunDirs`, `GitCommonDir`;
- the attempt: command (argv templating, redaction, exit mapping, `ExitNotYet`),
  http (request, `Idempotency-Key`, `hooks.RetryAfter`) and script
  (`hooks.Engine.StepAttempt`, outcome and schema checks);
- output helpers: `OutcomeOf`, `LastLines`, `LimitedBuffer`.

The package does not import `pkg/control`. What it needs from the
controller is the `Host` interface: `Run` (the isolated command runner, so
sandbox, mapped user and cgroup stay in control), `Worktree` (`addWorktree`
under `gitMu`), `SlotEnv` and `HostEnv` (`JOBSHOP_BIN` and PATH).

Stays in `autostep.go`: claim, lease and heartbeat, the per-step concurrency
limit, restart handling (`on_restart`), the step record and comment,
close/release by result, metrics, stop-on-shutdown (a cut attempt leaves
its claim), and the `stepHost` adapter. `scriptstep.go` keeps
`scriptInputs`, which reads the live callback runtime. The run-dir removal
and sweep stay as controller wrappers around `stepexec` because they run
off the tick (`offGo`) and take `gitMu`.

`ExitNotYet` is still exported from `pkg/control` (an alias of
`stepexec.ExitNotYet`). Unit tests for the package use a fake `Host`; the
step tests in `pkg/control/autostep_test.go` are unchanged.
