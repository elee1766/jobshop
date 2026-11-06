# Worker question round trip (#186)

On gasserver after deploying #186, with the Matrix gateway running:

1. Start a session on a scratch ticket and, in it, run
   `jobshop report waiting "need a decision" -q "Test question: reply 'ok'" --context $'Runbook test.\nAny reply closes it.'`.
2. `jobshop ticket list --label needs-human -s open` lists `Question: ...`
   with `meta.question`; `jobshop ticket list --report waiting` lists
   the scratch ticket.
3. A thread opens in the escalation room with the question text. No chat
   session starts for it (`jobshop session list`).
4. Reply `ok` in the thread. The question closes `answered`, the thread
   gets "question ... closed (answered)", the session gets the answer as
   a message, and the scratch ticket has an `answer to question` comment
   and `meta.report.state` `progress`.
5. `jobshop trigger list` shows no `step_stalled` or `session_silent`
   stall for the scratch ticket while the question was open.

Record the ticket ids and thread link on #186.
