# Starlark reads return comments (JOB-186)

2026-11-10

`tickets(...)` and `runs(...)` take `comments=False`. With `comments=True`
each result has `comments`: `{id, author, body, created_at}` oldest first,
`[]` when the ticket has none. Without it there is no key. `sessions(...)`
is unchanged.

- `api.TicketQuery.Comments` carries the flag. It is part of the call's
  kwargs, so the recorded input digest differs and a retried callback
  replays the result its first attempt saw, comments included.
- `hookHost.ReadTickets` loads comments only when set, with one
  `Store.CommentsFor(ctx, ids)` call (one `WHERE ticket_id IN (...)` query
  in `sqlite` and `postgres`), not a `Thread` read per ticket.
- A result over `Limits.DataBytes` fails as before; the error now starts
  with the built-in name (`runs: structured data byte/depth limit exceeded`).
- Tests: `storetest` `CommentsFor`, `pkg/execution/automation`
  `TestStoreReadComments*`, `pkg/orchestration/control` `TestHookReadsComments`.
