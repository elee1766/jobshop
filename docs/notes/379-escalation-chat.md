# Escalation chat owned by the controller (#379)

2026-11-04.

- When the controller files an escalation ticket (a run at its retry
  limit, or a pool quarantine), it opens one chat session keyed
  `escalation:<ticket>` in the configured chat pool and sends the pack's
  escalation prompt as the first message. `meta.escalation =
  {key, pool, session}` is written with `if_revision`, before the send,
  so a restart or a repeated tick never opens a second chat; the prompt
  only goes to a session this call started.
- Config: a workflow's `escalate: {chat_pool, prompt|prompt_file,
  pool_prompt|pool_prompt_file}`, or the top-level `escalation:` block
  for workflows without one. A pool quarantine uses the block of a
  workflow escalating to the same pool, else the top-level block. Prompt
  vars: `ticket run workflow issue_ref body`; pool prompt: `ticket pool
  error body`, written in the declaring pack's template syntax (#383:
  `{{.ticket}}` in a `template: go` pack, `${ticket}` otherwise). Prompt rules can match the chat with
  `keys: ["escalation:*"]`. No chat pool: ticket only. The `*_file`
  fields load through the #382 text file loader, with pack vars filled
  and the escalation vars kept.
- packs/github ships `prompts/escalation.md` and `escalation-pool.md` on
  every workflow, with `chat_pool: "{{.escalation_chat_pool}}"` (default
  empty, so off), and an `escalate-policy` system rule: reach a human
  only for a decision, a credential or access, or a host config problem.
- Closing the ticket sends one final message to the chat and sets
  `meta.escalation.final`. Questions keep their own answer path.
- Desk: the drawer shows `meta.escalation.session`; `/api/chat/send`
  routes an `escalation:<ticket>` key to the pool on the ticket; the chat
  list shows escalation chats as "Escalation <ticket>".
- Gateway (paused): it no longer sends a first prompt. With an escalation
  room it opens a thread only once the ticket has a chat, binds the
  thread to `meta.escalation.key` and pool, and posts the session's
  turns. Without escalation chat config it opens no thread.

## Holds beyond runs and pools (2026-11-04)

Every hold that needs a human mints one escalation chat, not only run
and pool escalations. The pack prompt is context for the model, not a
visible message: the session's first visible turn is its own summary,
then the human asks or instructs and the session messages workers or
adjusts tickets. (This supersedes "first message" above.)

- Causes, in `meta.escalation.cause` and `.held`: `run`
  (`escalate.prompt`), `pool` (`escalate.pool_prompt`), `wait` (a wait
  past its timeout, which labels the ticket needs-human), `guard` (a
  `jobshop guard` ask_human ticket) and `label` (any open task or step
  labeled needs-human by a worker, a person or a hook). The last three
  use `escalate.hold_prompt`, from the first workflow (by name) whose
  chat pool is in the ticket's project, else the top-level block.
  `label add needs-human`, a guard ask and a wait timeout open the chat
  at once, the sweep catches the rest.
- Left out on purpose: questions (`report --question`, `meta.question`)
  keep their answer path to the asking session. Admission and intake
  approvals (`question` label) are decisions about inbound text with no
  worker to unblock. A ticket with no workflow chat for its cause gets
  none.
- Template vars, always set (empty when the cause has none): `ticket`,
  `title`, `cause`, `reason`, `body`, `held`, `held_title`, `worker`,
  `run`, `workflow`, `step`, `issue_ref`, `pool`, `quarantine`, and
  `error` (the old name of `quarantine`). `worker` is the session that
  last held the failed step, raised the guard ask, parked the wait or
  added the label, only when it is a session.
- Context: `session.send` has a `context` field, carried in
  `MessageFields.Context`. The text rides in the prompt inside
  `[jobshop context]` ... `[end of jobshop context]` lines (ACP drops
  block annotations), and the session stream's `turn_start` event splits
  it out (`text` visible, `context` the rest). The escalation chat sends
  context only. The closing message stays a visible message.
- Powers: the chat pool gets no admin socket. It already has tickets,
  comments, labels, meta, `jobshop send` (`--interrupt`), `wait clear`
  and run resume/cancel in its project. New pool grant `ticket-release`:
  `release` works on a ticket another session of the same project holds.
  The operator pack grants it. No other admin method is added.
- `ListParams.Newest` lists newest first, as in the Desk escalations
  list (same field in the Desk session panel PR, #392).
- Tests: `TestEscalationChat` (vars, context-only first turn),
  `TestEscalationChatHolds` (label and guard paths, question skipped, one
  chat per hold), `TestEscalationChatSweepsLabel`,
  `TestEscalationChatPowers` (grant and its limits).

2026-11-05 (#566): prompts use `{{.ticket}}` only; `${ticket}` was
removed. See `docs/notes/566-drop-dollar-templates.md`.
