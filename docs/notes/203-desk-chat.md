# Chat in Desk (#203)

2026-11-03. Desk can now talk to sessions on the same `session.send` and
`session.subscribe` API the Matrix gateway uses.

- `jobshop desk --chat-pool operator` (the default) names the `type: chat`
  pool Desk sends to. Empty `--chat-pool` makes chat read-only.
- Chat view (`?view=chat`): the chat pool's open sessions, labelled by key
  (`matrix:...` from the gateway, `desk:...` from Desk). "New session"
  starts a fresh one keyed `desk:<12 random chars>` with the typed text
  as its first message (`POST /api/chat/new`).
- Escalation panel: the ticket drawer shows a chat panel when the ticket
  has `meta.gateway.session` (#125). It streams that session and sends
  by `meta.gateway.key` in the chat pool, so a Desk reply is one more
  turn in the session the Matrix thread feeds. Matrix posts the turn's
  reply to the thread as usual. Desk does not show the other human
  messages in the Matrix room, only the session's turns.
- Attach: the Sessions view's journal tail is the same panel. Desk checks
  `GET /api/session/<id>/access`, which makes an empty `attach` send
  that the controller accepts only when the pool has `attach: rw`. Other
  sessions are read-only, with the controller's reason shown.
- Composer: Enter sends, Shift+Enter is a new line, "Interrupt" cancels
  the current turn and sends (`interrupt: true`). Tool calls show as one
  line each (title and status); http(s) links in output are anchors, the
  rest is text.
- Writes: `POST /api/chat/send` (`{key}` or `{session}`, `text`,
  `interrupt`) and `/api/chat/new`, behind the `X-Jobshop-Desk` header
  and a same-origin `Origin` check. Text is 1 to 32 KiB.
- Prompt and config changes made from chat go through the operator
  pack's PR flow; Desk has no config or file write endpoint.
- Tests: `pkg/desk/chat_test.go`; Playwright `ui/e2e/chat.spec.js`
  (runbook `docs/verify/desk-chat.md`).
