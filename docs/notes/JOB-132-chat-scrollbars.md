# Chat transcript scroll bars (JOB-132)

2026-11-08

Desk hides scroll bar chrome everywhere (#960) except the two chat
transcripts: `.transcript` in the chat rail and `.session-output` in the
read-only session viewer. These show an Acme-style bar: a `--column`
track with a flat, square `--paper` thumb, `--hi` on hover and drag, and
an `--ink` border while dragging. The tokens are `--scrollbar-*` in
`ui/shell/tokens.css` and follow the dark theme.

- Only the chrome is styled. Scrolling stays native, so wheel, keyboard,
  touch, the virtualizer's history loading and end-following are as before.
- `scrollbar-gutter: stable` reserves the 12px gutter before the transcript
  overflows, so rows do not shift when the bar appears.
- Chromium and Safari draw the `::-webkit-scrollbar` parts. Chromium skips
  them when `scrollbar-width` or `scrollbar-color` is set, so the rule
  keeps both at their defaults.
- Firefox has no `::-webkit-scrollbar`: it gets its native bar in the same
  colors through `scrollbar-color`.
- In forced colors, the transcripts set `scrollbar-width: thin` so Chromium
  draws its native, system-colored bar instead of the custom parts.
- Phones with overlay scroll bars (Android, iOS) keep their own transient
  indicator. The gutter is still reserved.
