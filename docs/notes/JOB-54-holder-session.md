# JOB-54: working session on in-progress cards and traces

2026-11-06.

- `session.list` reports `busy` while the controller runs a turn on a
  session, and `tool`, the title of its latest open tool call (from ACP
  `tool_call` / `tool_call_update`; completed and failed calls drop out).
- `ticket.trace` nodes carry `holder` while the ticket is in progress.
- Desk `api/work?detail=1` returns each run's steps and last comment
  (`detail`) and the holder sessions of in-progress steps (`sessions`: id,
  display name, pool, model, alive, busy, tool). Attention's In progress
  makes one request instead of 1 + N. At most 20 runs get detail.
- Desk `api/ticket/{id}` adds `sessions` for the ticket, its run steps and
  trace, and fills `trace[...].session`.
- `api.openSession(id, {readonly: true})` sets `ro=1` in the URL. The chat
  pane then hides the composer, Stop and Interrupt whatever `/access`
  says, and shows a read-only note with a link to the session's ticket.
  It is a view mode, not a permission. Opening another session clears it.
- Run cards, step lists and trace nodes open the holder read-only. The
  ticket drawer's "held by" opens read-only; "Message this session" opens
  it with the composer.
