# Desk session console

2026-11-06. The session panel is a console with tabs instead of a column
on the right.

- Desktop: a strip along the bottom of the page column, resizable by
  its top edge (height in `localStorage` `desk-console-h`). Minimize
  (chevron or Esc) hides the body and keeps the tab bar; a tab click or
  restore opens it again.
- Phone: full screen with the tab bar on top; minimize returns to the page.
- Tabs: opening a session anywhere (`api.openSession`, a `?session=` link)
  adds a tab. `url.session` is the active tab, so links and back/forward
  work as before. The list persists in `localStorage` `desk-console-tabs`,
  at most 12. Tab labels are the pool, the full name and model in the tooltip.
- The console lives in the built-in `desk` app (`ui/plugins/desk/Console.svelte`).
- The chat pane moved its close button to the console bar and shows the
  model (and effort), token usage and session id in a footer bar.
