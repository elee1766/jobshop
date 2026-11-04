# Desk session chat panel

2026-11-04. `ui/src/views/SessionPanel.svelte`, `ui/src/lib/panel.svelte.js`. PR #392.

- Any session id in Desk opens that session's chat in a panel: Sessions
  rows, the chat list, the ticket drawer (holder, comment authors,
  escalation chat) and run progress reports. No backend change.
- Desktop: a third grid column on the right (default 440px, drag the left
  edge or use arrow keys to resize, double-click resets, width kept in
  `localStorage`). The page stays visible and live beside it.
  Phone (max-width 768px): the panel is fixed and full screen with a
  header back button, safe-area padding, and `body` scroll locked.
- State is `?session=<id>` (works on any view). Opening pushes a history
  entry marked `{panel: 1}`; the close button and Esc step back when that
  entry is the current one, so the browser back button and close agree.
  Opening a session closes the ticket drawer. Esc closes the drawer first.
- The chat is the existing `Chat.svelte`. Composer shows when the pool
  allows rw attach (`/api/session/<id>/access`) or the session is a live
  chat-pool session (sends by key). Closed sessions are read-only.
- The old inline journal row in Sessions (`Journal.svelte`) and the
  two-pane Chat view were replaced by the panel. Breakpoint for the phone
  layout is now 768px everywhere (was 720px).

## Escalation sessions (2026-11-04, #379)

- `GET /api/escalations` lists the tickets labeled needs-human that have
  `meta.escalation` (the controller's chat): ticket, title, cause
  (run, pool, wait, guard, label), first line of the body, held ticket,
  session, open or closed with resolution, age. Open first, oldest first,
  then the latest 50 closed. Needs `ListParams.Newest` from the
  escalation-chat-379 branch (the same two-line change is on both).
- Chat view: an Escalations table under the operator chat list; escalation
  keys are left out of that list. A row opens the session panel. In
  Attention, a ticket item that has an escalation chat opens the panel
  instead of the drawer (the drawer still shows the chat inline).
- A `turn_start` event with `context` (session.send context) shows as a
  collapsed "Context given to the session" block, not as an operator
  bubble. Older controllers send no context and nothing changes.
- Tests: `TestEscalations` (Go), `ui/e2e/escalations.spec.js` (mocked API,
  runs against any build: `npx vite preview --port 7480` or a live Desk).
