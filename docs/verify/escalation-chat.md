# Escalation chat (#379)

Not yet run. Use a throwaway root as in `docs/verify/desk-chat.md`
(steps 1, 2 and 4), with these changes.

1. In the root's jobshop.yaml, include packs/github with
   `escalation_chat_pool: operator` (the `type: chat` pool whose agent is
   `control.test fake-agent`). The `drill` workflow fails at once.
2. `jobshop --root <root> run start drill --var issue=1 --key drill-chat`
   and let the dev session close the step failed (or close it by hand:
   `jobshop --root <root> close <step> -r failed`).
3. `jobshop --root <root> ticket list --label needs-human` shows one
   `Escalate run` ticket; `ticket show <id>` has `meta.escalation` with
   `key: escalation:<id>`, `pool: operator` and a `session`.
4. `jobshop --root <root> session list` shows exactly one session keyed
   `escalation:<id>`. Its first message is `prompts/escalation.md` with the
   ticket, run and body filled in (no `${...}` left, no gateway wording),
   sent as model-only context: the first visible turn is the agent's summary.
5. Restart `serve`. The session list is unchanged and the session got no
   second prompt.
6. Desk: the drawer for the ticket shows "Escalation chat"; the Chat
   view lists it as "Escalation <id>". With `DESK_CHAT_ESC_SEND=1` the
   Playwright escalation test sends through `/api/chat/send` by key.
7. `jobshop --root <root> close <id>`: the session gets one
   `jobshop: escalation <id> closed (done)` message, and
   `meta.escalation.final` is true.
8. `jobshop --root <root> config show` lists the `escalate-policy`
   prompt rule (`at: system`).

9. Holds (2026-11-04): from a worker session run `jobshop label add <id>
   needs-human` on its ticket. The ticket gets `meta.escalation` with
   `cause: label`, and the session list shows a second chat. A `report
   --question` ticket gets none.

Pass: steps 3 to 9 hold. Without `escalation_chat_pool` the drill files
the ticket only: no `meta.escalation`, no session.
