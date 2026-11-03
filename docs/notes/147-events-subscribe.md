# events.subscribe change stream (#147)

2026-11-02. Admin socket method `events.subscribe` streams store changes
as `events.event` notifications on the calling connection. `jobshop
events [--since N] [--kind K] [--label L]` prints them as JSON lines.

- Source: the `history` table. Every write already records a row with a
  global `seq`, so the change log is history itself. Comments now also
  write a `comment` row (no revision bump). The store calls the
  controller after each committed transaction; one pump reads new rows
  and fans them out.
- Event: `seq`, `time`, `kind` (`ticket`, `run`, `session`), `op` (the
  history op; sessions map `create`/`close` to `start`/`stop`/`crash`,
  runs map a state change to `state`), `id`, `revision`, `actor`, plus
  the ticket's status, labels and run state read when sent.
- Filters: `kinds`, and `labels` (the ticket has the label, or the op is
  `label_add`/`label_remove` of it).
- Each subscriber has a bounded buffer (1024). A full buffer drops the
  subscriber: it gets a `lagged` event, the call ends with `lagged: true`
  and the client resyncs. Writes and other subscribers never wait.
- Resume: `since_seq` replays history rows after that seq, then goes
  live. A `ready` event marks the switch. Without `since_seq` the stream
  starts at the current end.
- Gateway: escalation threads are driven by the stream filtered on
  `needs-human`. `ready` (each connect) runs one resync pass over open
  escalation tickets; `create`, `reopen` and `label_add` open a thread;
  `close` posts the final message. The 10s poll (`EscalationPoll`) is
  gone. `lagged` or an error reconnects after 2s and resyncs.
