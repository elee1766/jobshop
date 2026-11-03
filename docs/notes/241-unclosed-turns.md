# Turns that end without close or release (#241)

2026-11-03.

Cause: case 1, agents ending a turn to wait. The journal shows
`stop_reason=end_turn` on every affected turn, not provider errors or
restarts. The jcode harness moves any bash command running longer than a
few seconds to the background and returns control to the model; the
model then ends its turn "waiting for the background task" (a CI watch,
a test run). The controller sends one reminder, the agent ends again,
and the ticket was released as a failed attempt. The work itself was
fine: the same session closed the ticket `done` on a later turn.

Changes:
- The first controller release of a ticket for an unclosed turn is a
  deferred release (10 minutes), not a failure. A second one on the same
  ticket (seen in its history) counts as before, so a ticket nobody
  finishes still reaches its retry limit. The session's stuck counter
  (two unclosed turns in a row) is unchanged.
- The ticket prompt footer and packs/github implement, ci-fix and merge
  prompts say to wait inside the turn: keep polling a backgrounded
  command (`gh pr checks <n>`, `sleep 60`), or `jobshop release --until`
  for long waits.

Test: `TestUnclosedTurnFirstReleaseNoFailure`.
