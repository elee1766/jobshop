# Trace shows the whole graph (JOB-190)

2026-11-10.

- `ticket.graph` and `ticket.timeline` return the graph of a ticket from its
  top ticket. `root` names that top ticket, the new `requested` names the
  ticket the call asked for (after a reference such as `owner/repo#12`
  resolved). Every member of a graph returns the same tickets, relations and
  root.
- The top ticket is found by climbing `parent_id` and `created` links from the
  requested ticket (`topTicket` in `tickettrace.go`). With several tickets that
  have nothing above them, the oldest by creation time wins, then the lowest id.
  The climb does not stop at chat or operator tickets, so a run started from a
  long-lived session opens that session's whole graph. A ticket outside the
  caller's project is not climbed through and not listed. The climb reads at
  most 2000 tickets.
- A `created` relation is now found from both sides. Before, only the holder's
  side saw it (`childIDs`). `creatorIDs` uses the same rule in reverse: the
  holders of the creator's hold windows that contain the creation time, with
  session tickets skipped and the parent left to `parent_id`. A ticket made
  while its creator held two tickets hangs from both. A ticket the same session
  made outside every hold window is not pulled in.
- The walk still starts at the requested ticket, so a graph cut off by the
  limit keeps that ticket's surroundings.
- Store: `HeldAt(actor, at)` lists the tickets an actor may have held at a time
  (a claim by the actor not ended before `at`). It is a superset: the caller
  confirms with the ticket's hold windows. Migration `sqlite 00028` and
  `postgres 00013` add a partial index on claim rows (`actor, at`), so the read
  does not scan history.
- Desk `GET /api/trace/<ref>`: the graph and the details come from the root.
  When the requested ticket is not the root and the request has no `inspect`,
  it is returned as `inspected` (with `inspection`). An explicit `?inspect=`
  wins. The timeline is asked for the requested ticket and returns the same
  rows.
- UI: the timeline and the graph always render the whole graph. The opened
  ticket is the selection and its row or card is brought into view (the
  chart's own scroll, or the graph camera when the card is off screen).
  Selecting a ticket sets `inspect` but never clears the trace, so the rows
  stay on screen while the details load. Only a change of `ref` clears it.
  `open as graph root` is gone from the details pane and from the ticket menu;
  `open in trace` stays for tickets outside a trace page. The wait target button
  in the timeline selects the target when it is in the graph and opens its trace
  otherwise. The URL is not rewritten after load: `?ref=<sub-item>` and
  `?ref=<root>&inspect=<sub-item>` show the same view.
- Not changed: `ticket.trace` (the drawer), the details pane layout (JOB-185).
