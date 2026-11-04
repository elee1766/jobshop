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

## Holds beyond runs and pools (2026-11-04)

Model: a hold that needs a human mints one escalation session, shown in
Desk as an escalation session. The pack's prompt is context for the
model with room for the cause. The session summarizes first, then the
human asks or instructs, and the session messages workers or adjusts
tickets.

- Paths that mint a chat (`pkg/control/escchat.go`, `causeOf`), cause in
  `meta.escalation.cause` and `.held`:
  `run` (step at its retry limit, `escalate.prompt`), `pool` (quarantine,
  `escalate.pool_prompt`), `wait` (a wait past its timeout labels the
  ticket needs-human), `guard` (a `jobshop guard` ask_human ticket) and
  `label` (any open task or step labeled needs-human, by a worker with
  `jobshop label add`, a person, or a hook). The last three use
  `escalate.hold_prompt`, from the first workflow in the ticket's
  project that has a chat pool and a hold prompt. `label add
  needs-human` opens the chat at once, the sweep catches the rest.
- Left out on purpose: questions (`report --question`, `meta.question`)
  keep their answer path to the asking session; admission and intake
  approvals (`question` label) are decisions about inbound text and have
  no worker to unblock; nothing is minted for a ticket in a project with
  no workflow chat for the cause.
- Template vars, every one always set (empty when the cause has none):
  `ticket`, `title`, `cause`, `reason`, `body`, `held`, `held_title`,
  `worker`, `run`, `workflow`, `step`, `issue_ref`, `pool`,
  `quarantine` (`config.EscalationVars`). `worker` is the session that
  last held the failed step, raised the guard ask, parked the wait or
  added the label, only when it is a session.
- Context injection: `session.send` has a `context` field. The text
  rides in the prompt inside `[jobshop context]` ... `[end of jobshop
  context]` lines (ACP drops block annotations on the wire). The session
  stream's `turn_start` event splits it out: `text` is the visible
  message, `context` the rest. The escalation chat sends context only, so
  the first visible turn is the agent's summary. Message tickets carry it
  in `fields.context`.
- Powers: the chat pool (operator) is admin-socket-free. It already has
  tickets, comments, labels, meta, `jobshop send` (a message ticket to a
  worker session, `--interrupt`), `wait clear` and run resume/cancel
  (any session, its project). New pool grant `ticket-release`: `release`
  works on a ticket another session of the same project holds, so the
  chat can hand a blocked worker's ticket back. No other admin method is
  added (no reopen, no session send, no reap). The operator pack grants
  it.
- `ListParams.Newest` lists newest first (Desk's escalation list).
- Tests: `TestEscalationChat` (vars, context-only first turn),
  `TestEscalationChatHolds` (label and guard paths, question skipped, one
  chat per hold), `TestEscalationChatSweepsLabel`,
  `TestEscalationChatPowers` (grant and its limits).
