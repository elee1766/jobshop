# Trace: comments and history on the timeline (#1182)

2026-11-10.

- Comments and history events are ticks on their row's bar at their own
  times. Ticks sit in the strip above the bars so they never cover one. The
  chart holds no prose. A tick is a button: it opens a `Marker detail` panel
  under the chart with the author, time and full comment (Markdown), or the
  event's label and raw history row. `prev`, `next` and `close` step through
  that row's markers. Ticks closer than 10px draw as one flat block that opens
  the first of them.
- `ticket.timeline` markers gained `comment` (the comment id) and `seq` (the
  history row), and a kind `event` for history rows no bar shows: `handoff`,
  `reopen`, `label_add`, `label_remove`, `edge_add`, `edge_remove` and `update`
  edits of title, body, pool, priority, holder or meta. Claims, releases,
  closes, waits and turns are already bars. Events are capped at the newest 40
  per row, like comments.
- The full text comes from the selected ticket's thread (`inspection`), which
  Desk already loads on selection. A tick is selected with its row, so the
  panel says `loading the full text…` for a moment.
- The events tab lists spans and the close only. The bottom comments tab holds
  just the comments the server sent no marker for (past the newest 40, or a
  ticket the timeline does not draw). `every history row` is unchanged.
- The `automation invocation` line under the trace is gone. Nothing stored or
  invoked changed: the invocation is still in `meta`, the audit view and the
  `raw` tab.

Check: `docs/verify/desk-timeline-markers.md`.
