# Pass wake for new and finished steps

2026-11-10 (JOB-193). A step a callback created waited up to 30s for the
next supervisor pass, because command and http steps start in the pass
(`runAutoSteps`) and sessions are sized in the pass (`reconcile`).

## What wakes a pass

All wakes go through `Controller.wakePass` (`offpass.go`): a non-blocking
send on `offWake`, a channel of size 1. The serve loop runs a whole pass
on receiving it. Wakes come from:

* `offFinish`: an off-pass call finished and left a result (JOB-103).
* `HookApplied`: a callback invocation was applied and its plan created at
  least one ticket (a step or a run). A plan with only comments, a finish
  or field changes does not wake. Edges in a plan only join new steps to
  existing ones, so new tickets are the only readiness change a callback
  makes.
* `finishAutoStep` and `closeTicketIn`: a step closed `done` or `skipped`.
  The steps behind it in a workflow without callbacks, or the ones a
  callback has not yet created, may now be ready. The callback for the
  closed step is already kicked from `checkRun`.

The ticker, `Options.Pass` and reload still run passes as before.

## Bound on extra passes

The channel holds one wake. Wakes that arrive while a pass runs, or before
the serve loop takes one, make one further pass in all. Several invocations
applied during a pass cost one pass.

No loop: every pass kicks the callback worker (`kickHooks`), but the worker
applies an invocation only when `hooks.RunDue` finds a finished step whose
revision was not completed yet (or a run not started). Applying records
`Completed`, so a run with nothing new due applies nothing, calls
`HookApplied` not at all, and asks for no pass. An applied invocation that
creates a ticket wakes one pass; that pass can only apply another
invocation if a step it started finished meanwhile, which is work done.
`TestCallbackStepsStartWithoutTicker` and
`TestIdleRunAppliesNoInvocationPerPass` check this.

## Not covered

These still wait for the ticker (30s) or the idle session poll (5s):

* A wait that resolves (`github.checks`, `run`, a deploy). The wait pump
  has its own wake (`waitWake`); the step it frees is started by the next
  pass.
* A step reopened by `ticket reopen`, request-changes or a lease expiry.
* A step deferred with a time (`retry_after`, `release --until`): this is
  a real time, the pass at the ticker finds it.
* A step closed `failed` with retries left: it reopens and is retried at
  the next pass.

The same helper fits each of them (call `wakePass` where the step becomes
ready). They were left out because each needs its own check against a loop.
