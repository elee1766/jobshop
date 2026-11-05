# Desk PWA, push and theme on the plugin model (#535)

2026-11-05. Supersedes `docs/notes/desk-push.md` and
`docs/notes/392-desk-pwa.md` (PR #392, adopted through #535).

- **Shell, not plugins:** manifest, icons and `sw.js` (`ui/public`, no
  offline cache); Web Push (`pkg/desk/push.go`, `webpush.go`,
  `/api/push/key|subscribe|unsubscribe|test`, `desk --push`); the theme as
  `--desk-*` custom properties (`docs/desk-plugins.md`); the phone frame:
  bottom tab bar from `view` contributions and the full-screen session
  panel (`ui/src/views/SidePanel.svelte`, slot `session`).
- **Built-in plugins** (`ui/plugins/`, `api` only, the #517 import check
  passes): `attention` (list, In progress, push opt-in through
  `api.push`), `chat` (session panel body, Chat view with escalation
  chats, drawer section on session, step and task tickets). Deploys is
  a pack plugin now (`packs/deploy`, #577).
- Web Push: RFC 8030, aes128gcm (RFC 8291), VAPID (RFC 8292) on the
  standard library. State in `<root>/desk/push.json` (0600): VAPID key
  made on first start, subscriptions, keys already sent. A new attention
  item is pushed once while listed; items waiting when push is turned on
  are not. The message opens `./?item=<kind:id>`, resolved against the
  worker scope, and the Attention plugin opens that item's chat.
- Base path (#516): the worker registers at `api('sw.js')` with scope
  `api('')`; manifest `start_url`, `scope` and icons are relative. CSP
  stays `script-src 'self'`, with `worker-src 'self'; manifest-src 'self'`.
- Phone: the shell follows `visualViewport` (`--vvh`, `--vvtop`), so the
  composer stays above the keyboard; tables become stacked rows at
  768px and below.
- Browsers allow push only on HTTPS or localhost; on the tailnet use
  `tailscale serve` and pass that name with `--host`. iOS pushes only to
  a Desk added to the home screen.
- Tests: `TestEncryptRFC8291`, `TestPushAttention`, `TestPushRoutes`;
  `ui/e2e/attention-chat.spec.js` checks install assets and 360px.
  Not verified by CI: delivery to a real device and keyboard placement
  on a real phone (`docs/verify/desk-pwa.md`).
