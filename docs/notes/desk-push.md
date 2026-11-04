# Desk push notifications (PWA)

2026-11-04. `pkg/desk/push.go`, `pkg/desk/webpush.go`, `ui/public/`.

- Desk is installable: `manifest.webmanifest`, icons and `sw.js` are in
  `ui/public` and land in the embedded `dist/`. The service worker only
  shows pushes and focuses Desk on click; no offline cache.
- `jobshop desk --push` turns on Web Push (RFC 8030, aes128gcm RFC 8291,
  VAPID RFC 8292), written on the standard library: no vendor SDK or
  account, and it works with the FCM, Mozilla and Apple push services.
- State is `<root>/desk/push.json` (0600): the VAPID key made on first
  start, the browser subscriptions, and the attention keys already sent.
  Desk stays an admin-socket client and does not use the database.
- What is pushed: new `/api/attention` items (needs-human runs and
  tickets, quarantined pools, rolled-back deploys), once per item while
  it stays listed, keyed `kind:id`. Items waiting when push is first
  turned on are not sent. Read every minute and 2s after ticket, run or
  pool events. A 404 or 410 from the push service drops the
  subscription.
- API (CSRF header on POSTs, like the rest): `GET /api/push/key`,
  `POST /api/push/subscribe|unsubscribe|test`. Routes exist only with
  `--push`.
- UI: a bell in the nav footer subscribes and unsubscribes, plus a test
  button. Hidden when the browser has no Web Push or Desk runs without
  `--push`.
- Browsers allow service workers and push only on HTTPS or localhost.
  On the tailnet, put Desk behind `tailscale serve` (HTTPS with a
  ts.net cert) and pass that name with `--host`. iOS delivers Web Push
  only to a Desk added to the home screen.
- Checked: `TestEncryptRFC8291` (the RFC's vector, byte for byte),
  `TestPushAttention` (a fake push service verifies the VAPID JWT and
  decrypts the payload; once per item; gone subscriptions dropped;
  state survives restart), `TestPushRoutes`. In a browser: the
  manifest and worker are served, the worker registers, and permission
  prompts. Playwright's browser builds cannot subscribe to a real push
  service, so delivery to a device needs one manual check.
