# Desk session chat panel

2026-11-04, PR #392. Updated 2026-11-05 for #535: the panel frame is
`ui/src/views/SidePanel.svelte` (shell) and its body is the `session`
plugin slot, filled by the built-in chat plugin (`ui/plugins/chat`,
`docs/notes/535-attention-chat.md`).

- Any session id in Desk opens that session's chat in a panel: Sessions
  rows, the chat list, ticket drawer links (holder, comment authors),
  run progress reports and attention items.
- Desktop: a third grid column on the right (default 440px, drag the left
  edge or use arrow keys to resize, double-click resets, width kept in
  `localStorage`). The page stays visible and live beside it.
  Phone (max-width 768px): the panel is fixed and full screen with
  safe-area padding, and `body` scroll locked.
- State is `?session=<id>` (works on any view), plus `?item=<kind:id>`
  when it resolves an attention item. Opening pushes a history entry
  marked `{panel: 1}`; close and Esc step back when that entry is the
  current one, so the browser back button and close agree. Opening a
  session closes the ticket drawer. Esc closes the drawer first, and
  stops a running turn when the composer has focus.
- Escalation sessions (#379): `GET /api/escalations` lists tickets
  labeled needs-human with `meta.escalation`. The Chat view shows them
  apart from the operator chats; a row opens the panel. A `turn_start`
  with `context` shows as a collapsed system block, not an operator
  message. Tests: `TestEscalations` (Go), `ui/e2e/attention-chat.spec.js`.
