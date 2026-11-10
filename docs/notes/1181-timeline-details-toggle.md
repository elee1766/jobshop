# Trace: timeline or details, not both (#1181)

2026-11-10.

- The trace page opens on the timeline. A `details` button in the timeline
  heading swaps it for the selected ticket's detail pane, a `timeline` button in
  the details heading swaps back. Both have accessible names
  (`Show details of <id>`, `Show the timeline`).
- The timeline stays mounted and is hidden, so zoom, folds, the chosen segment,
  the selected row and the scroll offset are as they were. `useSize` keeps the
  last chart width while the chart is hidden, so it does not reflow on return.
- A deep link with `tab=` opens the details. The details box scrolls inside
  itself on desktop and the page scrolls on mobile.
- A watched session no longer shrinks the chart: under the timeline it takes its
  own block, and beside the details it sits in the details view.
- #1183 reworks this page (selection-driven lower panel, mobile popup). This
  is the minimal swap and does not add a second permanent layout.

Check: `docs/verify/desk-redesign.md` browser acceptance, plus
`ui/e2e/timeline.spec.js` (the swap and the deep link).
