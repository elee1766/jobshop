# Turn indicator

2026-11-10. Isolated controller with the ACP test agent only (see
`desk-after-react.md`). `npm --prefix ui run test:e2e:isolated` runs
`turn-phase.spec.js` on desktop and mobile.

Expected: a chat whose first message is `phases <tag>` shows `thinking…`;
creating `<tag>.1` in the agent directory shows `running: Phase tool <tag>`;
`<tag>.2` shows `replying…`; `<tag>.3` ends the turn and the line disappears.
A reload during the tool phase shows the tool line again. Stop (or the
cancel API) removes it. The read-only session viewer behaves the same.

By hand, look at the line in light and dark themes at 1100 and 412 px: it is
muted, aligned with the rows, ellipsized when the tool title is long, and the
transcript still follows new rows while it is at the end.
