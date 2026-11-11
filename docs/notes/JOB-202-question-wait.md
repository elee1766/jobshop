# Question wait (JOB-202)

2026-11-11. Extends docs/notes/381-waits.md and docs/notes/JOB-11-event-waits.md.

A step that asked a Desk question could only park on outside state
(`github.review`, `github.pr`, a timer), so the answer did not resume it.
New `jobshop wait add` condition:

- `question:id=<question ticket>`: met when that question is closed with
  any resolution (`answered`, `withdrawn`, `dismissed`, `cancelled`).
  The wait comment names the resolution.

Behavior:

- `wait add` refuses an id that is not an existing question ticket
  (`meta.question`) and one in another project (also for admin: the
  wait belongs to the parked ticket's project). An already-closed
  question is accepted and the wait clears on its first evaluation.
- Closing a question wakes its waits at once (subject `question:<id>`,
  like `run:<id>`): `answerQuestion`, the generic ticket close, and the
  controller's derived close (orphan sweep). No poll delay.
- Cleared with `--resume`, the step goes back to the parking session as
  for any other condition. The answer is still the comment on the step
  that `answerQuestion` wrote, so a new session sees it too.
- Composes with `--any`:
  `jobshop wait add --any 'question:id=<q>' 'github.pr:pr=<repo>#<n>,state=head_changed' --resume -r "..."`.
- `jobshop report waiting -q ... --context ... --wait` files the question
  and parks the ticket on it with `--resume` in one call (owner
  `operator`). It cannot combine with `--until` and needs a task or step.
  For a second condition, run `wait add --any` with the question id from
  the output.
- The ticket graph and timeline show the wait as an edge to the question
  ticket.

Tests: `pkg/orchestration/control/waitquestion_internal_test.go`.
Runbook: `docs/verify/question-wait.md`.
