# Escalation chat owned by the controller (#379)

2026-11-04.

- Pack policy, not Go policy. A workflow's `escalate` block names the
  chat and the first message:
  `escalate: {chat_pool, prompt | prompt_file, pool_prompt | pool_prompt_file}`.
  `chat_pool` is a `type: chat` pool. Prompt files are read at load,
  relative to the declaring file and loaded by the #382 loader
  (`pkg/config/textfile.go`, `workflows.<n>.escalate.prompt` and
  `.pool_prompt`, sources shown by `config show`). Like step prompts they
  are not pack-substituted at load: the vars wait for the ticket. Setting
  both a prompt and its file is a load error. No `chat_pool`: the ticket
  is filed and no chat starts.
- Prompt vars (written in the pack's template syntax, `{{.name}}` in a
  `template: go` pack as #383 defines, filled by `Workflow.Render` when
  the ticket is filed, besides the pack vars): run escalations
  `ticket`, `run`, `workflow`, `issue_ref` (how to read the issue),
  `body`; pool quarantine escalations (the
  workflow whose `escalate_pool` receives the ticket, needs
  `pool_prompt`) `ticket`, `pool`, `quarantine`, `body`.
- `packs/github` ships `prompts/escalation.md` and
  `prompts/escalation-pool.md` (the former gateway wording, minus the
  gateway) and sets `chat_pool` from the pack var `escalation_chat_pool`
  (default empty, so no chat). selfdev passes it through. Desk's
  `--chat-pool` should be the same pool, since Desk sends by key in it.
- The controller opens the chat when it files the ticket and again each
  tick for open `needs-human` tickets without one: `session.send` to
  pool `chat_pool`, key `escalation:<ticket id>`, with the rendered
  prompt. It then writes `meta.escalation` with an `if_revision` update.
  A mutex covers send plus record, and a ticket that already has
  `meta.escalation` is skipped, so a restart or a retry starts no second
  session. A crash between the send and the record leaves the keyed
  session, and the retry's send reaches it again (a second first message
  is possible in that window).
- `meta.escalation = {key, session, pool, final?}`. When the ticket is
  closed the controller sends one message ("Ticket <id> is closed
  (<resolution>)") and sets `final`. Questions (`meta.question`) get no
  chat.
- Prompt rule `escalation-policy` (`at: system`, packs/github) tells
  every pool to escalate only for a decision, a credential or access, or
  an unfixable host config problem, to report progress through
  `jobshop report progress` and comments, to wait with `release --until`
  and to `recall` first. Pool prompts say the same in one paragraph.
  A host rule can match `keys: ["escalation:*"]` for chat-only policy.
- Desk: `GET /api/ticket/<id>` carries `meta.escalation` in the ticket;
  `/api/chat/send` with `{"key": meta.escalation.key}` reaches the
  session in Desk's chat pool; `/api/chat` titles `escalation:<id>`
  keys "Escalation <id>"; `/api/attention` ticket items carry `chat`
  (the key).
- Gateway (paused): it no longer sends a prompt or starts a session. For
  a ticket whose `meta.escalation` exists it opens a thread, records
  `meta.gateway` (now also `pool`, and `key` is the controller's key),
  posts the session's turns and sends thread replies to that key in the
  chat's pool. Tickets with no chat get no thread, except questions as
  before. The gateway's `pool` setting is unchanged for ordinary threads.
- Tests: `pkg/control/escchat_test.go` (first message from the pack
  prompt with vars, one session across a restart, closed message once,
  pool quarantine, no chat without config), `TestEscalationPolicy`
  (pack, config), `TestOperatorGrant` (rule in `config show`),
  `TestEscalationChatSend` (Desk), `pkg/gateway/escalate_test.go`.
