# Trace axis, cards and bars (JOB-126)

2026-11-08. On an isolated controller and Desk (see `desk-redesign.md`),
never a live one: the spec creates and closes tickets.

```sh
npm --prefix ui run check && npm --prefix ui run build
go build -o "$JCODE_SCRATCH_DIR/jobshop" ./cmd/jobshop
DESK_URL=http://127.0.0.1:<port> DESK_MUTATE=1 \
  npm --prefix ui run test:e2e -- trace-axis.spec.js live.spec.js search.spec.js
```

`trace-axis.spec.js` (desktop and Pixel 7) checks:

1. Card left edges and relative tick labels within 1.5px of the axis scale at
   fit, zoomed-in-and-panned and zoomed-out cameras.
2. An active trace shows `start` and `now`; after closing every ticket it shows
   `ended` and no `now`, still aligned.
3. Focus and tap on a time show the exact date with time zone; Escape hides it.
4. Cards lead with the title, show one status letter with a full label, no
   status word, and no clipped content.
5. `/trace` and graph bars share a height, their controls share a height, the
   search field is at most 420px, the form is not clipped, the page does not
   scroll sideways, and typing opens no dropdown.

By eye, light and dark, at 1280, 790, 412 and 360px: on an active trace whose
last card is days old, `now` stays at the right edge with `→` until you pan
or zoom out to it, and zooming out switches the ticks to days; the tooltip
stays inside the graph near either edge.
