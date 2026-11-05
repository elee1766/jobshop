# Pre-started attention chats (#606)

2026-11-05. `pkg/control/attentionchat.go`, `pkg/desk/attention.go`.

- The attention chat logic moved from Desk to the controller as
  `attention.chat {kind, id, title, why}`, so Desk, gateways and digests
  share it. Desk's `POST /api/attention/chat` only calls it.
- Every tick, `sweepAttentionChats` starts the chat of each item in
  attention, oldest first, in the attention chat pool
  (`attention.prestart_pool`, default `operator` when it is a chat
  pool), keyed `attention:<kind>:<id>`. The key goes in
  `meta.desk.attention_chat` on tickets. Escalation tickets keep the
  escalation chat (#379) and are skipped.
- The context prompt goes as model-only context: the first visible turn
  is the agent's short summary and the actions available (decision
  buttons, CLI commands). Its text, clipped, becomes the row summary
  (`meta.desk.attention_summary`, #607) while the chat is read-only.
- Cap: `attention.prestart_max` (default 4) pre-started chats open at
  once, and never more than the pool's `max_sessions` (the human slot
  stays free). Items past the cap start on click.
- Read-only until a human writes: the session's socket answers only
  read methods (`readOnlyMethods`), and edit, delete and move tool
  calls are rejected at the permission prompt. A `session.send` lifts
  it. Kept in `SessionFields.ReadOnly`, so it survives a restart.
- A click attaches: `attention.chat` returns the open session and sends
  nothing; the pane replays the journal (#529) with the first reply, or
  streams the turn if it still runs.
- Idle past the pool's `idle_timeout`, a pre-started chat is suspended:
  the host stops, the session stays open (`SessionFields.Suspended`).
  A click or a message resumes it with `session/load`.
- The chat closes (resolution `resolved`) when its item leaves
  attention; Desk's Acknowledge closes it too (`acknowledged`).
- Config: `attention.prestart` (default true when a pool resolves),
  `attention.prestart_pool`, `attention.prestart_max`.
- Tests: `TestAttentionPrestart`, `TestAttentionChatAttaches`,
  `TestAttentionAck`.
