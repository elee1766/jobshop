# Desk embedded apps (JOB-151, 2026-11-09)

Operators can add iframe tabs to Desk with `desk.embeds` in the service
config. Reference, schema and trust boundaries: [desk-embeds.md](../desk-embeds.md).
Browser checks: [verify/desk-embeds.md](../verify/desk-embeds.md).

## How it is built

- `pkg/config/embed.go`: `DeskEmbed` and its checks, run by `Desk.validate`
  on every load and reload, so a bad entry keeps the old config.
- `pkg/desk/embed.go`: `GET /api/embeds` and the live `embeds` section give
  the browser the validated list with the final `sandbox` and `allow`
  strings. Desk drops `allow-same-origin` for an embed on its own host name.
  The index page adds `frame-src` for the embed origins and a
  `jobshop-embed-origins` meta tag. Assets and the API keep the old CSP.
- `pkg/control/reload.go`: a successful reload publishes a `state` event, so
  open Desk streams re-read the list at once.
- `ui/src/screens/Embed.tsx`: route `?view=embed&embed=<id>`. The id is looked
  up in the list, so no URL from the address bar reaches an iframe. One
  iframe at a time. It is keyed on its address and permissions, so a change
  replaces it.
- Header tabs: the 760 to 899 px rule that hides `sys` and `catch-up` is off
  while embeds exist, and the nav wraps instead.

## Decisions

- Query-string route, not `/embed/<id>`: Desk serves one page at any proxy
  prefix and has no path routes (JOB-150). The id is still stable.
- No server-side URL fetch, proxy, token forwarding or message bridge.
- Secrets in URLs are rejected by key name. This is a guard against the
  common mistake, not a guarantee.
- A new origin needs a page reload because the CSP is sent with the page.
  Desk detects that and says so instead of showing a blocked frame.
