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
