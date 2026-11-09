# Ticket context menu

2026-11-09. Run only against an isolated controller (the spec starts and
cancels runs). Define a disposable workflow accepting `repo`, `pr` and `issue`
variables that stays open, for example one command step with `run: 'exit 75'`
and a long `retry_after` (see `compact-run-links.md`). Rebuild the frontend and
the Go binary, restart Desk, then:

```sh
DESK_URL=http://127.0.0.1:7492 DESK_MUTATE=1 \
  DESK_LINK_WORKFLOW=link-acceptance \
  npm --prefix ui run test:e2e -- ticket-menu.spec.js --workers=1
```

Expected on desktop and mobile: the menu opens from a home row (right click on
desktop, the `⋯` button everywhere), a trace card and a relationship entry. The
footer shows that row's exact id and `copy ticket id` puts the same id on the
clipboard with a `copied` notice. Escape and a click outside close it without
changing the URL, and Escape returns focus to the button. `inspect` acts on
the entry the menu was opened on, not the selection. On desktop the menu stays
inside a 700x420 viewport when opened at each corner.

Also look at, by hand: the context-menu key on a focused `⋯` button; right
click on a PR or issue link (the browser's menu, not Desk's); selected text in a
row (browser menu); light and dark themes; a long-press on a phone.
