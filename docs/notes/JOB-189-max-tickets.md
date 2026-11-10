# Single-use sessions: pool `max_tickets` (JOB-189)

2026-11-10.

- `max_tickets: N` on a ticket pool makes its sessions single use. A
  session counts each ticket it claims (`fields.tickets`, once per ticket
  via `fields.last_ticket`). At N it claims nothing new, finishes what it
  holds and ends `ticket_limit`. 0 (default) keeps today's behavior.
  Chat pools reject it.
- While the session's ticket waits with `--resume`, the session parks: the
  runner records `suspended` and `parked=<ticket>`, stops the host and
  keeps the session open (the #606 suspend path). A parked session is not
  in `open`, `working` or the `max_sessions` count that the scaler uses, so
  the pool starts another session for new work. `pool list` shows it as
  `parked`.
- The supervisor resumes a parked session (no host until then, then
  `session/load`) when its ticket is ready, pre-assigned to it by the wait,
  and the pool's counted sessions are below `max_sessions`. Otherwise it
  stays parked. A ready ticket held for a parked session is not demand.
- A parked session ends `parked_gone` when its ticket is closed, deleted or
  no longer waits to resume that session (the wait cleared while it was
  busy, or went back to the pool), and `drained` when its pool is drained.
- The park only happens for waits with `--resume` (the wait names the
  session). A plain deferral is not parked: the session sits idle until
  `idle_timeout`.
