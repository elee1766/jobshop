# Desk graph view removed (#1179)

2026-11-10

The trace has one view, the timeline (#1103). The card graph, its `graph` and
`timeline` toggle buttons, the `layout` URL value and the graph-only code and
styles (`Graph.tsx`, `graphLayout`, `.graph-*` rules for cards, camera and
axis) are gone. `?layout=graph` and `?layout=timeline` are ignored and open the
timeline.

Kept: `ticket.graph` and the `graph` part of the trace stream. The ticket
details pane reads it for relations, parent, history labels and the selection.
Wait edges (#1091) are still in that data and in the details pane; the
timeline draws blocking arrows only.
