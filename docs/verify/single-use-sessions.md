# Single-use sessions live check (JOB-189)

On a controller with a ticket pool that sets `max_tickets: 1` and
`max_sessions: 2` (an isolated one, not live Jobcity). Notes:
[JOB-189-single-use-sessions](../notes/JOB-189-single-use-sessions.md).

1. Create two ready tickets in the pool. Each is worked by its own session:
   `jobshop session list --all` shows both sessions closed `ticket_limit`
   once their tickets close, and no session worked two.
2. Create a ticket whose agent turn runs `jobshop wait add <cond> --resume`.
   `jobshop session list` shows that session `suspended`, `parked: <ticket>`,
   its host gone, and `jobshop pool list` counts it under `parked`, not
   `open`. Create two more tickets: two other sessions start for them
   while the first stays suspended.
3. Clear the wait (`jobshop wait clear <ticket>`). While the pool is full,
   the parked session resumes at once and `pool list` shows `open` one over
   `max`. `session journal <id>` shows `session/load` with the same
   conversation id. It closes the ticket and ends `ticket_limit`.
4. Park two sessions, fill the pool, then clear both waits. One resumes,
   the other stays suspended until a slot frees, ahead of new sessions for
   other ready tickets. With `limits.max_sessions` equal to the open count,
   neither resumes until a host slot frees.
5. Wait past `idle_timeout`: the suspended session stays open. Close its
   ticket by hand: the session ends `parked_gone`.
6. Default: a pool without `max_tickets` still hands one session the next
   ticket in the same conversation.
