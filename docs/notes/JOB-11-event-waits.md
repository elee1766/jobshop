# Event waits (JOB-11)

2026-11-05. Extends docs/notes/381-waits.md.

New `jobshop wait add` kinds, evaluated in the same waits loop:

- `github.tag:repo=o/r,tag=v1.2.3` (`release=` is the same key): the tag
  ref exists.
- `deploy:repo=o/r,sha=<sha>`: the controller's own build commit
  (`jobshop version --server`) is sha or a descendant (GitHub compare
  `identical` or `ahead`). A controller with no build commit refuses the
  wait at `wait add`, so nothing parks on a deploy it cannot see. The
  issue suggested reading deploy state from tickets, but #577 left only
  failure tickets (rollback, config-check), not the deployed rev. The
  running binary is the deployed rev on a self-deploy host.
- `run:id=<run>,state=done|closed`: done is state done (a cancelled run
  invalidates the wait); closed is any end.
- `linear.state:issue=KEY,state=Done` matches the state name or type
  (`completed`), case-insensitive. `linear.label:issue=KEY,label=X[,present=false]`.

Release guard: `release --until` whose reason names CI, checks, a PR,
deploy, a tag or a merge is refused with the matching `wait add` lines.
`--time-wait` (RPC `time_wait`) marks a real timer: a window or a
backoff.

Prompts: post-merge deploy and tag steps, implement's "later" clause,
escalate-policy, operator, the agent pack's deliver and reviewer prompts,
and stall-check now name the wait kind. stall-check undefers a step
whose timed release names an event that already happened, and records a
learning. Prompt grep for `release --until`: what remains is
documentation of the flag, the "do not use for CI" lines, plan.md's
"do not park on a proposal", and real timers (reviewer backoff).

Tests: TestWaitEventKinds, TestWaitRunCancelled, TestWaitDeployNoCommit,
TestWaitEventKindsValidate, TestReleaseGuard (pkg/control).
