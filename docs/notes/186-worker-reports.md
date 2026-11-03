# Structured worker reports and questions (#186)

2026-11-03. Phase 6 item; idea from deepseek-ai/dsh-superagent-prototype.

**Report.** `jobshop report progress|waiting|ready|blocked|failed|dropped
[note] [-q question] [--until T] [-t id]` (RPC `report`). A session
reports on a ticket it holds; admin on any. The controller writes
`meta.report` `{state, note, question, until, by, at}` and a comment
(`report <state>: <note>`), so the report counts as activity.
`blocked` adds the `blocked` label, any other state removes it, so the
existing stall `blocked` signal follows report states. `waiting --until`
releases the ticket deferred (as `release --until`). Report never
closes: `failed` and `dropped` record why before `close`.
`ticket list --report <state>` filters on `meta.report.state`.

**Question.** `--question` files a message ticket to `human` with label
`needs-human`, priority 1 and `meta.question {ticket, session}`. A
message is not claimable and no session's inbox matches `human`, so no
agent picks it up and it is not an orphan.

**Answer.** `jobshop answer <question> <text>` (RPC `question.answer`,
admin): comments the answer on the question and the asked ticket, sends
it as a message to the asking session when that session is open,
resets the report to `progress`, lifts a defer, removes `blocked`, and
closes the question `answered`.

**Gateway.** Questions are escalations to the gateway: it opens a
thread in `escalation_room` with the question text, but starts no chat
session. The first allowed reply in the thread is passed to
`question.answer`; the close posts "question ... closed (answered)".
Later replies go to a chat session as in any thread.

**Stall-check.** A held step or session whose last report is `waiting`
on an open question is not `step_stalled` or `session_silent`; the
question itself is an open needs-human ticket, so `needs_human` covers
it. Pack prompts: implement reports states and asks with `--question`;
plan reads `ticket list --report`; stall reads `meta.report`.

Evidence still to collect on gasserver: `docs/verify/worker-questions.md`.
