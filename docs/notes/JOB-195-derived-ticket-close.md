# Derived tickets close with their cause (JOB-195, 2026-11-10)

A ticket the controller files for another ticket records the cause in its
meta already: an escalation has `meta.run.id` and `meta.run.escalation`
(`<run>/<n>`), a question has `meta.question.ticket`. No new key.
`pkg/orchestration/control/derived.go` closes them:

- Run escalations. `settleEscalations` runs when a new round is filed, when
  a run is set running again (resume) and when it closes (done, cancelled).
  Earlier rounds close `superseded`, comment "superseded by <id>" (a new
  round) or "run <id> resumed". A closed run closes every open round, with
  the run's resolution (`done`, `cancelled`) and "run <id> done|cancelled".
- Held escalations. A round a session holds is not closed while its run
  lives: it gets one comment saying why it is stale and stays open. When the
  run closes it is closed too, and the comment says it was held.
- Questions. `sweepDerived` (each pass) closes an open question whose asked
  ticket is closed, resolution `withdrawn`, comment "asked ticket <id>
  closed (<resolution>)", and clears the Linear Blocked label like an answer.
- The same sweep repairs escalations filed before this change or missed by a
  crash. It uses the escalation chat's final message path (`escalationClosed`).
- Attention chats already end when their item leaves attention
  (`sweepAttentionChats`). Worker-authored messages ("patch (push blocked)")
  are created by the agent with no cause link, so they are not covered.
