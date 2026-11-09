# JOB-171: conversation mechanics in pkg/conversation

2026-11-09. Child of JOB-164. No behavior change: the control tests for
managed chats, attention, questions, escalation chats, attachments and
reconnects pass unchanged.

## What moved

- `pkg/control/chatlife` is now `pkg/conversation` (managed-chat rules, the
  escalation chat policy). The name follows the JOB-164 layout.
- `pkg/control/attention` and `pkg/control/question` are now
  `pkg/conversation/attention` and `pkg/conversation/question`. They stay
  separate leaves: `question` is used alone (config-free, no store), and
  `attention` is the projection `Build` that Desk and RPC read. Merging
  them into one package would only make the leaf harder to use alone. This
  removes the two `family-conversation` baseline exceptions.
- From `pkg/control`: the message envelope (`messagesxml.go`), message
  addressing and the per-batch attachment budget (`inbox`, `deliver`), and
  the attachment limits, file layout and prompt blocks
  (`chatattachment.go`). New names: `Envelope`, `MessageElement`,
  `EnvelopePlaceholder`, `Recipients`, `Addressed`, `FitsBudget`,
  `ValidateAttachments`, `AttachmentDir`, `SaveAttachments`,
  `RemoveAttachments`, `PromptAttachments`.
- Desk used its own copies of the `desk:`, `escalation:` and `attention:`
  prefixes. It now calls `conversation.Managed`, `DeskPrefix`,
  `EscalationPrefix` and `attention.Key`, so the prefix rules have one owner.
  The A2A gateway only sends on a key from the escalation meta and has no
  key logic.

## What stays in control

- Routing, admission and authorization: grants, project scope, the
  stale-option check and every answer (`attentionAnswer`, `answerQuestion`),
  `fileQuestion`, escalation filing (`escalate.go`), `chatManage`, the
  read-only gate, `sessionSend`.
- Everything with a goroutine, lock or session: the runner's `deliver`
  (claims messages, runs the turn, reacts to interrupts through
  `interruptible`), `suspended`, `chatMu`, the revision-checked writes, the
  sweeps in the pass. `pkg/conversation` starts no goroutine, holds no lock
  and takes no context except `ClassifyHold`/`Build` reads. Cancellation and
  shutdown are therefore still the runner's.
- Persistence: `conversation` writes only the attachment files. The ticket
  rows (message, session fields, meta) are written by control. Attachment
  root and secret redaction (`redactWrite`) are passed in: `SaveAttachments`
  takes a redact function, so the package never learns the scope or the
  activity feed.

## Not moved

`messages.go` `deliver` and `interruptible` need the ACP connection, the
runner and the store claim, so they are the session host's. `escalatechat.go`,
`attentionchat.go`, `questionchat.go` and `deskchat.go` start and resume
keyed sessions and write session fields. Their pure rules were already in
this package after JOB-147.
