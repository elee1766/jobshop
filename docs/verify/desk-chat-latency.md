# Desk chat latency and decision actions (#603)

On gasserver after deploy. Record times in a comment on #603.

Latency, before and after:
1. Open Desk chat on an idle operator session. Note the clock, send "say
   hi". The message shows at once in a dashed box, "sent, waiting for
   the session".
2. `jobshop ticket show <session>` history and the journal
   (`jobshop session journal <session>`): message created, then
   `turn_start`. Target: under 1s (was about 4s, the runner poll).
3. Time from send to the first streamed event (tool row or text) in the
   pane. Target: about 5s or less.

Decisions, with a test proposal (an issue labeled `proposed` and a
ticket from the planner prompt's `ticket create` line):
1. Click the item in Attention, then Approve as target, OK. The issue is
   relabeled `target` with a comment naming you; the ticket closes
   `approved`; a planner ticket parked on it is ready within a minute.
2. A second proposal: Dismiss. It leaves Attention; the issue keeps
   `proposed`.
3. A third: Reject. The issue is closed as not planned.
4. From an attention chat session: `jobshop ticket undefer <id>` works
   on a deferred ticket in its project.

Pre-started attention chats (#606), record on #606:
1. File a test question: `jobshop ticket create "prestart test" -l
   question`. Within a minute `jobshop session list` shows an operator
   session keyed `attention:question:<id>` and its journal has one
   turn with the agent's summary.
2. Wait 30s, click the item in Desk. Measure click to first rendered
   agent message (target: under 1s). `jobshop session journal
   <session>` shows no second `session/prompt`.
3. Before writing, the session cannot write: from its socket a
   `jobshop ticket create` fails "read-only until the operator sends a
   message". Send "thanks" in the chat; writes work after.
4. Close the question ticket. Within a minute the session closes
   `resolved`.
