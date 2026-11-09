# Ticket context menu (JOB-136)

2026-11-09. Desk has one ticket menu, mounted once in `App`
(`ui/src/components/TicketMenu.tsx`, state in `ui/src/state/ticketMenu.ts`).

- Opened from home in-flight rows, waiting items and not-moving rows, trace
  graph cards, the ticket details header, relationship entries and the parent
  line, and trace search results. Attention items that are not tickets (pool,
  config, ingest, journal lag) get none.
- Three ways in: right click, the context-menu key or Shift+F10 (the browser
  sends `contextmenu`), and a visible `⋯` button on every row, which is the
  touch and keyboard route. Shift+right click, links, inputs and selected text
  keep the browser's own menu.
- A row hands the menu a snapshot of its ticket (id, ref, session, whether the
  trace can inspect it). Actions use that snapshot, never the selection.
- Items: `open in trace` or `open as graph root` (hidden for the current
  root), `inspect` (graph tickets, hidden for the inspected one), `view
  session` (ticket with a holder, or a session), `discuss` (waiting items),
  `attach to chat`, `copy ticket id`. No lifecycle or destructive actions.
- The footer shows the exact id in a selectable block. A copy confirms with
  `copied <id>` and a failed copy shows an error banner and `could not copy` in
  the footer. Without `navigator.clipboard` (plain http) it falls back to
  `execCommand`.
- Base UI Menu does placement (flips and shifts inside the viewport), arrow
  keys, Escape and outside click. Escape and keyboard selection return focus to
  the control that opened it. A route change closes the menu.

Check: `docs/verify/ticket-menu.md`.
