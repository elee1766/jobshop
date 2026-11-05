# #392 Desk PWA under a path prefix

2026-11-05, adopted through #535. Push itself: `docs/notes/desk-push.md`;
session panel: `docs/notes/392-desk-session-panel.md`.

- PR #392 was merged with master after #508 (Artifacts), #515 (proxy
  user) and #516 (base path). Both sides kept: the Linear-style nav gains
  Artifacts (`g r`, unread badge) and the proxy user line.
- Every URL the PWA builds is base-relative, as #516 requires: API calls
  go through `api()`, the service worker registers at `api('sw.js')` with
  scope `api('')`, and the manifest's `start_url`, `scope` and icons are
  relative to it. Push messages carry `./?ticket=<id>`, which the worker
  resolves against its scope, so a notification opens Desk under
  `/desk/` too.
- CSP keeps `worker-src 'self'; manifest-src 'self'` and takes master's
  `base-uri 'self'`, which the `<base href>` tag needs.
- The VAPID key is made on first start and kept in `<root>/desk/push.json`
  (0600), never in config.
- e2e specs use relative URLs, and the mocked escalation spec strips the
  prefix, so `DESK_PREFIX=1` runs them under `/desk/`.
