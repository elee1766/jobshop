# TL page: timeline first, selection below (#1183)

2026-11-10. Behavior as built. Supersedes the swap in #1181 (the details
replacing the chart) and the marker panel under the chart in #1182.

- The page is named TL in the nav, the search bar (`/tl`, aria-label `TL
  reference`), the home link and the ticket menu (`open in TL`). The URL is
  unchanged: `?view=trace&ref=...&inspect=...&tab=...`, so old links work.
- The chart is always on top and never changes size. Whatever is selected
  shows in one lower area under it, in the page flow.
- The lower area starts with its tab bar, then one panel. Only the chosen
  panel is drawn. The tab bar is sticky, so it stays reachable while a long
  panel scrolls. The selected ticket's title is the bar's `h2`.
- Tabs: `events` (the "now" block, then the spans), `marker` (only while a
  tick is open), `session` (only while a session is watched), `details` (the
  old summary: parent, facts, time by state, actions), `input / result`,
  `relations`, `body`, `comments`, `raw`. `?tab=` holds the tab; an empty value
  is `events`, and a tab with nothing to show falls back to `events`.
- Pressing a name or a bar selects it. A comment or history tick opens the
  `marker` tab, a bar opens `events` on that span. "view session" and the
  ticket menu open the `session` tab. The read-only session is a tab, so it no
  longer changes the chart's size.
- The open tick lives in `markAtom`, not in the chart. It is cleared when
  another row is pressed.
- The verbose "child of" block is gone. The parent of the selected row is
  tinted (weaker than the selection) with a left rule, and its name button's
  tooltip ends with "parent of the selected ticket". The `details` tab still
  names and links the parent.
- Phone (max-width 599px): a press opens the lower area as a full-screen
  `dialog` over the page with a "close" button (Escape also closes). The chart
  stays mounted under it, so zoom, folds, scroll and selection are as left. A
  link with `inspect=` or `tab=` opens the popup; a link with only `ref=` shows
  the chart.
- e2e: `trace-view.js` `showDetails` / `showTimeline` work at both widths
  (on a phone they open and close the popup).
