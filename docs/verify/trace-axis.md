# Trace bars (JOB-126)

2026-11-08, trimmed 2026-11-10 when the graph view was removed (#1179). On an
isolated controller and Desk (see `desk-redesign.md`), never a live one: the
spec creates and closes tickets.

```sh
npm --prefix ui run check && npm --prefix ui run build
go build -o "$JCODE_SCRATCH_DIR/jobshop" ./cmd/jobshop
DESK_URL=http://127.0.0.1:<port> DESK_MUTATE=1 \
  npm --prefix ui run test:e2e -- trace-axis.spec.js live.spec.js search.spec.js
```

`trace-axis.spec.js` (desktop and Pixel 7) checks that the `/trace` search bar
and the timeline heading share a height, the bar's controls share a height, the
search field is at most 420px, the form is not clipped, the page does not
scroll sideways, and typing opens no dropdown.

The time axis of the timeline itself is covered by `desk-timeline.md`.
