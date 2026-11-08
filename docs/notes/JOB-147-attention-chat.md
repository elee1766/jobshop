# pkg/control split: attention, questions and managed chats (JOB-147, 2026-11-08)

Child of JOB-139. Extracts the pure rules of attention, questions and
managed-chat lifecycle into three packages. No behavior change: the
existing control tests pass through them unchanged.

## What moved

- `pkg/control/question`: what a question ticket carries. `Meta`
  (meta.question), `Options` validation, `Context` (links and recent
  comments), `OptionsText`, `ChosenText`, `ChatPrompt`, the answered
  resolution and the answer title prefix.
- `pkg/control/attention`: the attention projection. `Build` lists what
  waits on a human from a `Source` the controller implements (tickets by
  label, needs_human runs, timed-out waits, subsystem rows, stale rows,
  config and pool health). It keeps the deduplication: an escalation
  stands for its run and pool once, a ticket with two labels is one row,
  a stale ticket already listed is not repeated. `Actions` describes the
  choices of a row (Resume, Clear wait, Allow, Approve, ...) as data and
  never executes them.
- `pkg/control/chatlife`: the rules of the three managed chat types.
  Key prefixes (`desk:`, `escalation:`, `attention:`), name
  normalisation and uniqueness, what `new`, `rename`, `convert`,
  `archive`, `resolve` and `restore` may do to which chat, the view
  command checks, the read-only method and tool lists, the attention
  prompt, the summary and recommendation read from a chat's first reply,
  the sweep that finds attention chats whose item left, and the
  escalation chat policy: `Hold`, `ClassifyHold` (cause), `Policy` (pool
  and prompt), `Escalation` (meta.escalation).
- `pkg/store` gained small shared readers the packages need so none of
  them imports `pkg/control`: `MetaString`, `MetaMap`, the built-in pack
  names, `WaitOf`, `StepExhausted`, `ReviewCapped`, `LastClaimant`.

The new packages read tickets (`store.Store` or values passed in) and
never write.

## What stays in the controller, and why

- Grants and scope: `attention` (grant `attention`), `attentionAnswer`
  (`attention-answer`, project scope, operator attribution),
  `deskCommand`, `chatManage` (admin only).
- Stale-option validation and execution of an answer: `attentionAnswer`
  re-reads the items, matches the option id and `ExpectedTitle` against
  the fresh `Actions`, then `executeAttention` calls `answerQuestion`,
  `decisionResolve`, `runResume`, `runCancel`, `poolUnquarantine`,
  `waitClear` and `closeTicketIn`. All persisted ticket and run
  mutations stay there.
- Question filing and answering (`report.go`: `fileQuestion`,
  `answerQuestion`), follow-ups and re-asks (`questionchat.go`).
- Everything that needs a live session or the supervisor: starting a
  keyed session, `sendMessage`, `resumeSession`, `humanWrote`, the
  runner's `suspend` and `suspended`, the read-only gate
  (`checkReadOnly` reads the `readOnly` map under `c.mu`), the
  `chatMu` held across `chatManage`, `sweepAttentionChats` and
  `sweepEscalationChats` (they run in the pass with its config).
- Revision-checked writes: `updateSessionFields`, `updateDeskMeta`,
  `updateEscalationMeta`, `recordAttentionChat`.
- Close versus resolve: `archive` and `resolve` both close the session
  through `sessionCloseRPC` and keep the journal. `resolve` first closes
  the escalated ticket as done, and only when the chat is still that
  ticket's current session (`chatlife.CheckCurrent`). `convert` keeps
  the `escalation:<ticket>` key, so pending messages and source
  references stay valid, and `escalationClosed` skips a converted chat.
  `restore` keeps the ACP session id and journal and resumes on the next
  send.

The controller files (`attention.go`, `attentionchat.go`,
`escalatechat.go`, `deskchat.go`) keep thin aliases (`questionMeta`,
`EscalationMeta`, `escalationMeta`, `holdTicket`) so the rest of
`pkg/control` and its tests did not change.

## Not moved

`linearask.go` (posts a question to Linear) and the intake and
admission approvals that file needs-human tickets stay with the Linear
and ingest children. `decision.go` is the decision resolver, a write
path, and stays.
