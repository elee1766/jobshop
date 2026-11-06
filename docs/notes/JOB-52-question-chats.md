# Question chats answer, follow and close (JOB-52)

2026-11-06. A question's attention chat now works like an escalation chat.

- Desk has no "answer the question" toggle. The composer always sends
  to the chat agent, so the operator's messages show in the transcript.
- `question.answer` is grantable as `question-answer`; the operator pack
  grants it to its pool. A session answers only questions in its own
  project. The answer is recorded as from the last human who wrote to the
  chat (`fields.operator`), never a `from` the session passes.
- After a session answers, the chat records `fields.follow`. When the
  asking session's turn that took the answer ends (or the asker is gone),
  the controller sends the chat the reaction as context: re-asked (a new
  open question from the same session on the same ticket), parked (wait
  or defer, with reason) or resolved (closed, a report, or its status).
- A question from the same session and ticket while a chat for an
  earlier one is open records that chat's key in
  `meta.desk.attention_chat`, so it reuses the chat and no new one is
  prestarted. The chat is told about it.
- The sweep keeps a question chat open while it follows, while a
  threaded question is open, or while a Desk client streams it
  (`session.subscribe`). Otherwise it closes once the question left
  attention, as before.
- The chat pane treats a `state: closed` event as the end: no "working"
  line and no composer, it shows "Session closed · <resolution>".
