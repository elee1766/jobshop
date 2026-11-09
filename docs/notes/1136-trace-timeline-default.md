# Trace opens as the timeline (#1136)

2026-11-09

Desk's trace view now opens as the timeline (#1103). `?view=trace&ref=<id>`
with no `layout` is the timeline, `layout=graph` is the graph, and old
`layout=timeline` URLs still open the timeline. The `graph` button on the
timeline writes `layout=graph`, the `timeline` button on the graph removes
`layout`. The live stream asks for `timeline=1` on every trace now, since the
default view needs it.

`jobshop desk navigate view=trace ref=<id>` passes its arguments through as
query values, so it opens the timeline too. Add `layout=graph` for the graph.
