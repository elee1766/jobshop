# Automation mechanics audit: what moved and what stays (JOB-148, 2026-11-08)

Audit of `pkg/control` `hook*.go`, `scriptstep.go`, `autostep.go` against
`pkg/hooks` (durable invocations, audit, replay, transports) and
`pkg/automation` (Starlark engine and capability packages, not touched).
No behavior change. Callback manifests, run pins, scope checks, the
command sandbox, retries, recovery and unknown-effect handling are as
before.

## Moved into pkg/hooks

- `schema.go`: `CheckStepOutput` (size limit plus JSON Schema) and
  `CheckSchema` (compile-only for `start_step`, validating for finished
  steps). Was `control/hookoutput.go`. The controller passes
  `cfg.Automation.OutputBytes`; the package reads no config.
- `snapshot.go`: `StepInput`, `RunSnapshot` and `RunDue`. They build the
  accepted run/step envelope a callback sees, the read set, and decide
  which invocation a running callback workflow is waiting for
  (`on_start`, then `on_step_completed` for the first finished step
  whose revision is not completed) with its dedup key. This was the
  body of `enqueueRunHooks`; `scriptInputs` shares `StepInput`.
- `worker.go`: `Worker`, the bounded drain (enqueue, process up to 64,
  re-enqueue, reschedule on pending kicks) that was `kickHooks` plus two
  atomics on `Controller`. The host gives it a goroutine home, its
  context and the two steps.
- `RetryAfter` (exported from `http.go`): `autostep.go` had a copy.

Unit tests for each are in `pkg/hooks/extracted_test.go`. The existing
callback, restart, replay, scope, sandbox and script-step tests in
`pkg/control` run through the moved code unchanged.

## Stays in control (the host adapter), and why

- `hookhost.go` (`hookHost`): admission (`AdmitHook`), commit
  (`CommitHook`), apply (`ApplyHookPlan`), failure and applied hooks.
  They decide policy against the live config, take `runMu`, and end in
  one `store.Apply`. Moving them would put the whole controller behind
  an interface or duplicate `runMu`.
- `hookactions.go`: turning actions into a graph. It uses `planRun`,
  `buildRun`, `planTitle`, pools, placements and workflow rendering,
  which are the controller's, and it is where actor/pack scope
  (`hookReference`, `runScopeRecorded`, `checkPlannedTargets`) and the
  stale-read check (`hookReadSetCurrent`) are enforced before the
  atomic commit.
- `hookresume.go` (`resumeHook`, `repinRun`): run-state changes under
  `runMu` with the admin check. The replay safety rule itself is
  `hooks.Engine.Resume`.
- `hookcommands.go`: runs a guard hook command in the pool sandbox via
  `runIsolated`; sandbox and process groups are the controller's.
- `hook.go`, `hooks.go` (`enqueueHookEvents`): guard policy, ask-human
  tickets, claim denial and trigger matching against config.
- `hookaudit.go`: admin-only RPC wrappers over `Engine.List` and `Show`.
- `scriptstep.go`: placement, sandbox exec and worktree for a script
  step; the attempt journal is `Engine.StepAttempt`.
- `autostep.go`: command and http step execution, claims, leases and
  escalation. It is step machinery, not callback machinery. Its http
  step does not share `hooks/http.go` beyond `RetryAfter`: one posts a
  signed guard body, the other a templated step request.
