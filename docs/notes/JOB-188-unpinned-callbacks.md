# Callback runs follow the live callbacks (JOB-188)

2026-11-10. Reverses the default of [#1012](1012-hook-repin.md): a callback
workflow run is no longer pinned to the callbacks it started under.
Pinning is opt-in per workflow with `pin_callbacks: true`. No versions,
patch markers or migration of old runs were added.

## Default (unpinned)

- At each callback, script step, planned-action apply and resume, an open
  run uses the live entry points, code, host grants and reference scope of
  its workflow. A reload no longer stops it with `hook manifest changed`.
- The controller moves the run with `repinRun` wherever a pinned run would
  be refused (admit, apply of a journaled plan, script step, `run resume`).
  The run's fields change and it gets one comment, `Repinned to the live
  callback manifest (<new>, was <old>) by controller.`, per change, not per
  callback.
- An invocation queued or planned before the reload completes under the
  live callbacks. A move supersedes the queued invocation, which is queued
  again. The pool, placement, resource and reference rechecks of a plan stay.
- Replay safety is unchanged. Unknown, failed or redacted call records
  still refuse retry, and a retried invocation whose calls differ from its
  journal fails with the recorded-result error.
- A workflow that is no longer a loaded callback workflow, or a missing
  entry point, still fails the run.
- Plain `jobshop run resume <run>` works for a run stranded before this
  change. `--repin` on an unpinned run is accepted and does nothing extra.
- Authority: an unpinned run picks up changed grants and reference scope on
  reload with no admin action. The operator controls the configuration.

## Pinned (`pin_callbacks: true`)

Unchanged from #1012: the run stops with `hook manifest changed`, the
escalation names `--repin`, and `--repin` stays admin socket only. The
option is only valid on a callback workflow (config load error otherwise).

## Operator notes

- The option is read from the live configuration, not recorded on the run.
  Turning it on later pins open runs to the manifest they last ran under.
  Turning it off releases them: a run stopped on a changed manifest resumes
  with plain `run resume`.
- When an edit was incompatible, the callback fails on the run's existing
  steps or output and the run goes `needs_human` with the callback's error
  (not `hook manifest changed`). Fix the callback, deploy, then
  `jobshop run resume <run>`. If the run cannot be fixed, cancel it.
- The escalation text names `--repin` only for a pinned workflow.

Tests: `pkg/orchestration/control/unpinned_test.go`, `TestHookRunResumeRepin`
and `TestHookRunPinSurvivesUnrelatedReloadOnly` (both set `pin_callbacks`),
`TestWorkflowErrors` (config validation).
