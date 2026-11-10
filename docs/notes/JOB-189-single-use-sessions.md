# Single-use sessions: pool `max_tickets` (JOB-189)

2026-11-10.

- `max_tickets: N` on a ticket pool makes its sessions single use. 0
  (default) is today's behavior. A chat pool rejects it, and so does a
  negative value.
- A ticket counts when the session closes it, with any resolution
  (`fields.tickets`, set when the runner sees the ticket it claimed,
  `fields.claimed`, closed). A release (a wait, a defer, a drop) counts
  nothing. At N the runner treats the session as cordoned with the derived
  reason `tickets` (never stored): it claims nothing new and ends
  `ticket_limit`. A replacement starts through the scaler as usual, and a
  warm `min_sessions` session that used its ticket is replaced the same way.
- A session whose ticket goes onto a wait with `--resume`, or onto a timed
  defer the session made (`release --until`, assigned back to it), parks:
  `fields.suspended` and `fields.parked=<ticket>`, host stopped, session open
  (the #606 suspend path). It claims no other ticket and does not end on
  `idle_timeout`.
- Every suspended session, not only parked ones, is out of the slot counts:
  the scaler (`scaleSessions`), the human slot (`humanSlot`), the
  attention-chat pre-start cap and `limits.max_sessions`. `session list`
  and `pool list` still show them (`suspended`, `parked`).
- Resume: when the wait ends the ticket is assigned to the parking session.
  The supervisor resumes ready parked sessions oldest first
  (`session/load`, the same conversation id). One resume may take the pool
  to `max_sessions` + 1, so a finished wait does not queue behind new work.
  Further resumes stay suspended and resume as slots free, before new
  sessions for other ready work (their tickets are not counted as demand).
  `limits.max_sessions` is never exceeded: at the host limit the session
  stays suspended. A resumed session has `fields.resumed` until it settles
  its ticket, and the scaler does not cordon it as surplus.
- A parked session ends `parked_gone` when its ticket is closed, deleted,
  or no longer assigned to it (the wait cleared while it was busy and the
  ticket went back to the pool), and `drained` when its pool is drained.
