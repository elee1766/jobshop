# Rebase close checks the PR merge state (#511)

2026-11-05.

Before the controller passes `close -r rebase` on a step, it finds the
run's PR (the last PR link on the run or its steps, as revert detection
does) and reads `mergeable_state` from GET /repos/{r}/pulls/{n}.

- `dirty` or `behind`: the rebase stands, no failure counted.
- any other known state (clean, blocked, unstable, ...): the close
  becomes `failed`, counts an attempt, and the controller comments why.
- read error, or empty or `unknown` state: the rebase stands and the
  controller logs a WARN. A run with no PR link is not checked.

Test: `TestRebaseMergeState` (control).
