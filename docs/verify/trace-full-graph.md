# Trace shows the whole graph (JOB-190)

2026-11-10. Checks that opening a trace on a sub-item shows the graph from its
top ticket with the opened ticket selected. Behavior:
[JOB-190](../notes/JOB-190-trace-full-graph.md).

## Automated

- `go test ./pkg/orchestration/control -run TestTicketGraph` (same tickets,
  relations and root from every member, timeline rows, project scope).
- `go test ./pkg/transport/desk -run TestTraceWholeGraph`.
- `go test ./pkg/storage/store/...` (`HeldAt` on both engines).
- `ui/e2e/trace-whole.spec.js` and `ui/e2e/timeline.spec.js`. The stream is
  mocked, so they run against any Desk and write nothing:
  `DESK_URL=http://127.0.0.1:<port> npx playwright test e2e/trace-whole.spec.js`.

## By hand

Use a controller with a run that a session created while holding a step of
another run (any `pr-work` run started by an agent from a running ticket).
Call the first run A, its step S, and the run the session made B.

1. Open `?view=trace&ref=<B>`. The timeline lists the rows of A (A, its steps,
   B and B's steps), not only B's. B's row is the picked one, is in view, and
   the details pane shows B.
2. Open `?view=trace&ref=<A>&inspect=<B>`. The same rows, the same selection.
3. Open `?view=trace&ref=<S>`. The same rows again, S selected.
4. Click another row. The rows stay on screen while the details change. The
   URL keeps `ref` and changes `inspect`.
5. In the ticket details and in the `⋯` menu of a card or row: no
   `open as graph root`. `open in trace` stays on home and system rows.
6. Switch to the graph layout (`graph` button). Every card is there, the
   opened one is selected and in view.
7. Open a step that waits on a ticket in the graph, then press its wait target
   in the timeline detail. The target is selected, the URL gets `inspect=`, and
   no new trace opens. A wait on a ticket outside the graph opens its trace.
8. `GET /api/trace/<B>` and `GET /api/trace/<A>` (Desk) return the same
   `graph.root` and node ids; `graph.requested` differs.
