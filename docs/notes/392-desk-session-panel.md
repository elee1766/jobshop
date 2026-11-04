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
