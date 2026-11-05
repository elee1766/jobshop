# Attention opens the chat that resolves it (#535)

2026-11-05. `pkg/desk/attention.go`, `ui/plugins/attention`,
`ui/plugins/chat`.

- `/api/attention` merges the controller's `attention` list, so Desk
  shows `question`, `decision`, `stale`, `config`, `config-repo`,
  `resource` and `ingest` items with their issue or PR (`ref`).
- A click calls `POST /api/attention/chat {kind, id, title, why}`.
  Since #606 the controller owns this (`attention.chat`) and usually
  pre-started the chat when the item arrived, so the click only
  attaches (`docs/notes/606-attention-prestart.md`). Before #606:
  - a ticket with `meta.escalation` (#379) opens that chat;
  - otherwise the chat pool session keyed `attention:<kind>:<id>`;
    when none is open, Desk starts it with a context prompt (kind, id,
    title, issue or PR link, `run show` for runs).
  - The key is kept in `meta.desk.attention_chat` on tickets (`desk` is
    a built-in meta pack; the write is conditional on the revision), and
    in `<root>/desk/attention.json` (0600) for other kinds.
- The chat opens in the session panel with `?item=<kind:id>`, the item
  above the composer and a Resolve action per kind, each an existing RPC
  through `api.call` with the proxy user as actor (#515): close the
  ticket; `run.resume`; `pool.unquarantine`; `wait.clear`; for
  `decision`, relabel `proposed` to `target` or close it; for kinds with
  no RPC, Acknowledge (`POST /api/attention/ack`, hidden until it leaves
  the list and comes back).
- `question`: the composer answers through `question.answer` by
  default, so the answer reaches the asking session; a toggle switches
  to talking to the operator agent.
- The chat follows jcode-web's pane: Load older (`GET
  /api/session/<id>/events?before=`, decoded from `session.journal` by
  the same `rpc.EventDecoder` as `session.subscribe`), user, assistant
  (markdown, sanitized), reasoning, tool (updated in place, input and
  output on click), system and error blocks; Up and Down recall sent
  prompts; Enter sends, Shift+Enter is a newline, Ctrl+Enter interrupts,
  Escape stops. While a turn runs, Send reads Queue and Stop calls the
  new `session.cancel` (cancel the turn, send nothing).
- Session events gain `reasoning`, `usage` and tool `input`/`output`
  (cut to 8 KiB, redacted like the journal).
- Tests: `TestAttentionChat*`, `TestSessionEvents`, `TestChatCancel`,
  `TestSessionSendSubscribe` (cancel); `ui/e2e/attention-chat.spec.js`
  (mocked API, any build, also with `DESK_PREFIX=1`).
