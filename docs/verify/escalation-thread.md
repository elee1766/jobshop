# Escalation thread round trip (#125)

Against the local tuwunel on gasserver (`http://127.0.0.1:6167`). Passed
2026-11-02 with a throwaway root, a jcode `type: chat` pool `operator`,
and a copy of the `drill` workflow whose step was failed by hand.

1. Set up the bot, a test user and a room as in
   `docs/verify/matrix-gateway.md` steps 1 to 3, and add
   `escalation_room: <room id>` to the gateway.
2. `jobshop serve`, then `jobshop gateway <name>`.
3. File a throwaway issue and start the drill on it
   (`docs/verify/escalation-drill.md`). The run goes `needs_human`.
4. Within 10s the gateway logs `gateway: escalation thread`, the room
   gets a top-level notice `Escalation <ticket>: run <run> ... needs a
   human`, and the session's summary follows in its thread.
5. As the test user, reply in the thread with instructions (comment on
   the issue, resume the run at the step, close the escalation ticket).

Pass:
- The session comments on the issue, `jobshop run show <run>` is back to
  `running`, and the escalation ticket is closed `done`.
- Its `meta.gateway` has the thread, the session and `final: true`, and
  the thread ends with `jobshop: escalation <ticket> closed (done)`.
- A later reply in the thread is answered by the same session.
- A second gateway pass opens no new thread for the same ticket.

Clean up: cancel the run, close the throwaway issue, stop the session.
