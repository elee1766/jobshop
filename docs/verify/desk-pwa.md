# Verify: Desk PWA, push and attention chat (#535)

On a device that reaches Desk over HTTPS (`tailscale serve`, Desk with
`--push --host <name>.ts.net`):

1. Open Desk, add it to the home screen (iOS) or install it (Android,
   desktop Chrome). It opens standalone at Desk's base path.
2. Attention: Notify me, allow notifications, then Test. A notification
   arrives; tapping it opens Desk.
3. Make an attention item (`jobshop ticket create t -l needs-human`).
   Within a minute a notification arrives; tapping it opens the item's
   chat with the item above the composer.
4. On the phone, focus the composer: it stays above the keyboard, and
   the transcript does not scroll sideways.
5. Click a stale or question item twice: the second click opens the
   same session (`jobshop session list` shows one `attention:` key).
6. Mocked UI checks, any build: `cd ui && npx vite preview --port 7481`,
   then `DESK_URL=http://127.0.0.1:7481 npx playwright test
   e2e/attention-chat.spec.js` (desktop and Pixel 7). With a live Desk,
   also `DESK_PREFIX=1` (#516).
