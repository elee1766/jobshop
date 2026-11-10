# Desk timeline markers

2026-11-10. Run on an isolated controller (see `desk-redesign.md`). Never seed a
live instance.

1. Build the branch (`npm --prefix ui run build`, then `go build -o
   /tmp/tm/jobshop ./cmd/jobshop`), `jobshop init` a root under /tmp, serve with
   `--tick=1s`, serve Desk from the same binary.
2. Make a task. Add a long Markdown comment, a second comment, a label, a
   relationship to another ticket, and rename it.
3. Open `?view=trace&ref=<task>`. Expected: a tick on the task's bar for each,
   in the strip above the bar, no text in the chart. Ticks that land on the
   same spot are one flat block. Hover or focus names a tick (`comment: <first
   line>` or `event: label added: ...`).
4. Press a comment tick (click, touch, or Tab then Enter). Expected: a `Marker
   detail` panel under the chart with the author, the time and the full
   Markdown text, the tick marked pressed. `next`, `prev` and `close` work. For
   an event tick the panel shows its label and the raw history row.
5. Open `details`: the `events` tab lists spans and the close only, no comments.
   `comments N` appears only for comments with no tick (more than 40 on one
   ticket). `every history row` lists all rows. There is no `automation
   invocation` line under the trace, also for a ticket made by a hook.
6. Check 1000, 790 and 390 px, light and dark: no page overflow, the panel
   scrolls inside itself, ticks have a usable target on touch.
7. `DESK_URL=<desk> npm --prefix ui run test:e2e -- timeline.spec.js
   trace-detail.spec.js` passes on both projects. With `DESK_MUTATE=1` also
   `live.spec.js`.
