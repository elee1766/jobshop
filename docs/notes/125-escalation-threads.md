# Escalation threads (#125)

2026-11-02.

- Config: `gateway.<name>.escalation_room`, a room id that must also be
  in `rooms`. NixOS: `services.jobshop.gateway.<name>.escalationRoom`.
- The controller does not post. Its escalation ticket (label
  `needs-human`, `meta.run.escalation`) is the event. Every 10s the
  gateway lists open and in-progress `needs-human` tickets over the
  admin socket.
- For a ticket with no `meta.gateway`: post a top-level notice in the
  escalation room (its event id roots the thread), record
  `meta.gateway = {gateway, room, root, key}` with an `if_revision`
  update, then `session.send` the first prompt to the gateway's chat
  pool keyed `matrix:<room>:<root>` and record the session. The
  conditional update is the dedup: one thread per ticket, across passes,
  restarts and gateways. A thread that loses the race says so and is
  not used. A ticket owned by another gateway name is skipped.
- First prompt: the ticket id, run, how to read the run and the issue,
  what to do with the operator's answer (comment, `run resume`, follow
  ups, then claim and close the ticket and say so), and the ticket body
  (step history). The session posts a summary and waits.
- Replies in the thread go through normal routing (allowlists, same
  key), so they reach the same session, also after the ticket closed.
- When a ticket seen open is closed, the thread gets one final message
  and `meta.gateway.final` is set. A ticket that only lost its label
  gets none. `gateway` is a built-in meta pack name.
- Tests: `pkg/gateway/escalate_test.go` (thread, prompt, dedup, reply
  routing, final message, foreign owner), `TestParseGateway`. Live
  check: `docs/verify/escalation-thread.md`.
