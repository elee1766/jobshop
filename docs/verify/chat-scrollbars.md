# Chat transcript scroll bars

2026-11-08. Run against an isolated controller with an operator chat pool
(see `docs/verify/chat-responsiveness.md`), never live Jobcity.

```sh
DESK_URL=http://127.0.0.1:7492 npm --prefix ui run test:e2e -- desk.spec.js -g 'scroll bar'
DESK_URL=http://127.0.0.1:7492 DESK_MUTATE=1 npm --prefix ui run test:e2e -- chat-responsiveness.spec.js --workers=1
```

Playwright's headless Chromium passes `--hide-scrollbars`, so the bar is
not painted in its screenshots. To inspect it, launch with
`chromium.launch({ ignoreDefaultArgs: ['--hide-scrollbars'] })`.

Expected, on a chat with more than a screen of turns and in the read-only
session viewer, desktop and phone, light and dark:

- A 12px bar at the right edge: `--column` track, square paper thumb with a
  thin border, highlighted on hover and while dragging. No rounded corners.
- The text column width does not change when a transcript starts to
  overflow. The thumb does not cover message text.
- Dragging the thumb to the top loads older history until "start of
  conversation", keeping the reading position. A new message follows the
  end only when the view was at the end.
- Wheel, PageUp/PageDown and touch scrolling work as before.
- Firefox shows its native bar in the same colors. With forced colors
  (`forcedColors: 'active'` in a Playwright context) a native,
  system-colored bar with a visible thumb.
- Other panes (archive table, tables) still show no scroll bar chrome.
