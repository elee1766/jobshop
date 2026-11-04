# Sessions ending stuck after #241 (#301)

2026-11-04.

Cause, from the journals of js-0032m34mdg002, js-0032vd0zq8002 and
js-0032w4gavm002: every unclosed turn ended `stop_reason=end_turn` right
after a jcode background task started (a CI watch or a test run, moved
to the background automatically or with `wake`). The agent's last words
were "I'll be woken when the watch finishes", "I'll wait for the wake
notification", "Waiting on the CI wake." The reminder turn got the same
answer, so the session ended stuck. jcode's wake only starts a new turn
in an interactive client; over ACP nothing prompts the session, so the
wait never ends. Not context pressure: the last turns ran at under 20%
of the context window.

Changes:
- The reminder after an unclosed turn says no wake will come, the next
  unclosed turn ends the session, and lists the concrete choices: close,
  wait in the foreground (`sleep 60; gh pr checks <n>`), `release
  --until`, or `release`.
- The ticket prompt footer says a background task's wake does not start
  a new turn.
- The stuck reason comment on the session (shown by `jobshop report
  incidents`) adds the last turn's stop reason and the tail of what the
  agent said.

Test: `TestStuck` checks the reminder text and the recorded reason.
