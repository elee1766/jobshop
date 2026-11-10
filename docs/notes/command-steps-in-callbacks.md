# Command steps from callbacks

2026-11-10.

- `start_step(type="command", run=..., pool=...)` starts a command step
  without a script wrapper. Reference:
  `docs/configuration/callbacks.md#command-steps`.
- Constructor checks are in `pkg/execution/automation/actions.go`
  (`checkStartStep`) and `placement.go` (`checkStepFields`): `type` is
  `agent` or `command`. A command step needs `run` and rejects `prompt`,
  `prompt_file`, `script`, `output_schema`, `retry_pool`, `placement`,
  `max_review_rounds` and `on_request_changes`. `run` and `on_restart`
  are command-only. `retry_after` is for script and command steps.
- `Action.Run` is a `string` or `[]string`. `planHookActions` turns it
  into argv and writes `command_spec` on the step (`domain.CommandSpec`:
  argv, pool, timeout, retry_after, on_restart, worktree). The ticket
  has no pool, like a script step.
- `stepexec.CommandConfig` builds the step's `config.Step` from the spec,
  so the existing claim, lease, retry and restart path runs it with no
  workflow definition. A workflow command step has no spec and is
  unchanged.
- `run` is not filled as a template for a callback step.
- `JOBSHOP_VAR_*` comes from the run's accepted inputs (`RunFields.Inputs`),
  not from pack vars read again.
- The spec's pool is rechecked when the planned batch is applied
  (`applyPlannedHook`, `checkPlannedTargets`, now also for a script
  step's pool) and again before each attempt: a pool that is gone or in
  another project fails the attempt. The pool's sandbox policy is read
  live.
- `hooks.StepInput` gives a callback the spec's pool as `step.pool`.
- A session's `ticket update` cannot add, replace or remove
  `command_spec`, or change the type of a step that has one
  (`checkCommandSpec`). The admin socket can.
- Not on callback command steps: `concurrency`. The limit is the default
  (4) per project and step id.

Tests: `pkg/execution/automation/engine_test.go` (`TestCommandStepAction`
and the command rows of `TestStartStepFieldValidation`),
`pkg/orchestration/control/dynamiccommand_test.go`.
