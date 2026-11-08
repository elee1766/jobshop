# pkg/control split: Linear orchestration (JOB-142, 2026-11-08)

Third subsystem out of `pkg/control` under JOB-139, after journal
shipping (JOB-119) and compression (JOB-89). No behavior change: delegation
eligibility, team scoping, opt-in ask mirroring (`linear.ask_on_issue`),
timeouts, backoff and cancellation are as in JOB-118 and JOB-114.

## What moved

`pkg/control/linearsvc` owns, from `linear{ask,health,intake,sessions,webhook}.go`:

- the outage backoff (`Service.Call`, `Do`, `DoKey`, `Down`, 30s doubling
  to 5m) and the per-request timeout (`PollerCtx`);
- the agent-session sweep (`SweepSessions`) and its poll throttle;
- `linear_issue_no_run` intake (`Signal`, `HasOpenRun`, the read cache,
  `IntakeStates` for `linear.intake`);
- the webhook handler, server and event check (`Handler`, `Serve`,
  `Delivered`, `Event`);
- question and escalation mirroring (`SweepAsks`, `Answers`, `Unblock`,
  `AskOf`).

It uses `pkg/linear` for every request and the controller's shared
per-identity clients through `Host.Client`, so there is no second HTTP or
auth client and no retry change.

## What stayed

`pkg/control/linearhost.go` is the adapter (`linearHost`) and the thin
steps (`sweepLinearAsks`, `signalLinearIssues`, `linearAnswers`,
`linearIntakeRPC`). The controller keeps:

- clients and credentials (`apiclients.go`, `identitySource`);
- every store write: meta packs (`setMetaPack`), comments, write keys
  (`storeKeys`), `answerQuestion` with its grant check;
- the off-pass machinery (`offGo`, `offTick`), so slow requests stay
  outside the pass, and the tick lock the webhook takes to run triggers;
- `intake.go` (intake hooks): the hook engine, `refuseIntake` ticket
  filing and the run-start gate. They only call Linear for a read and a
  label, through the shared client; they move with hooks, not here.
- one-shot Linear calls with no state: `decision.go` approve/reject,
  `runs.go` gate, `activity.go`, `runtitle.go`, `identity.go`.

Waits and subscriptions call `Service.Do`/`DoKey` for their reads.
`TestLinear*` in `pkg/control` still drive the service through the
controller; `TestIntakeMatch` is a unit test in the package.
