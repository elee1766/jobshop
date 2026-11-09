# Desk embedded apps (JOB-151)

Desk can show your own web apps or dashboards as iframe tabs next to its
built-in views. This is a small embedding feature. It is not a plugin
system: Desk loads no code from an embed and gives it no data.

## Configure

Embeds are an ordered `desk.embeds` list in the service `jobshop.yaml`.
They are not part of a pack, and a pack that carries `desk:` does not load.
A config reload applies a change without restarting Desk.

```yaml
desk:
  embeds:
    - id: grafana          # route name, ^[a-z][a-z0-9_-]{0,31}$
      label: Grafana       # tab text, 1 to 32 characters
      url: https://grafana.example/d/abc/overview?kiosk
    - id: scratch
      label: Scratch app
      url: http://127.0.0.1:3000/
      sandbox: [allow-forms]       # optional, added to the baseline
      allow: [fullscreen]          # optional, browser features
```

| Field | Meaning | Default |
| --- | --- | --- |
| `id` | Stable route: `?view=embed&embed=<id>`. Unique. Not a built-in name (`home`, `trace`, `learn`, `improvements`, `system`, `sys`, `catch-up`, `embed`, `waiting`, `inflight`). | required |
| `label` | Text in the top bar. | required |
| `url` | Absolute `http` or `https` address. | required |
| `sandbox` | Extra iframe `sandbox` tokens. Allowed: `allow-downloads`, `allow-forms`, `allow-modals`, `allow-popups`, `allow-popups-to-escape-sandbox` (needs `allow-popups`), `allow-presentation`, `allow-same-origin`. | none: only `allow-scripts` |
| `allow` | Browser features delegated to the frame. Allowed: `autoplay`, `clipboard-read`, `clipboard-write`, `fullscreen`, `picture-in-picture`. | none |
| `allow_http` | Permit plain `http` on a host that is not loopback. | `false` |

At most 16 entries. Tabs follow the built-in navigation in list order.

A bad entry fails the load, so a reload keeps the old config and says why:

- a missing, duplicate or invalid `id`, or one that is a built-in name
- an empty, padded or over-long `label`
- a `url` that is not absolute `http(s)`: `javascript:`, `data:`, `file:` and
  relative or protocol-relative addresses are rejected
- credentials in the URL (`user:pass@host`), or a query or fragment key that
  looks like a secret (`token`, `secret`, `password`, `api_key`, ...)
- plain `http` to a host that is not loopback, unless `allow_http: true`
- a `sandbox` or `allow` value that is not in the lists above
- an unknown field

`allow-top-navigation`, `allow-pointer-lock`, `allow-orientation-lock`,
`allow-storage-access-by-user-activation`, camera, microphone, geolocation
and screen capture are not offered.

## What Desk does

- The iframe is in the normal content area. The header, status line and
  chat rail stay. It has a title (`<label> (embedded app)`) and an `open
  externally` link that opens the URL in a new tab with `noopener`.
- Only the selected embed is mounted. Leaving it unmounts the frame, so a
  hidden app keeps no timers, sockets or memory. Coming back reloads it.
  There are no background tabs.
- The route holds an id, never a URL. Desk looks the id up in the
  configuration the controller sent. `?view=embed&embed=nope`, or a URL in
  that parameter, shows a message and no frame.
- Direct links, refresh, Back and Forward work, and so do proxy prefixes,
  because the route is a query string like Desk's other views.
- When the config changes, the open embed follows it. A removed embed drops
  its frame and tab at once. A changed address or permission set replaces the
  frame. An embed whose origin was not configured when the page loaded cannot
  be framed until the page reloads, and Desk says so.
- With no `desk.embeds`, nothing changes: same page, same policy, same
  navigation.

## Trust boundaries

An embed URL is a trust decision by the operator who writes the service
config. Nothing a ticket, agent or payload carries can add or change one, and
Desk never fetches the URL on the server.

- **Sandbox.** Every frame gets `sandbox="allow-scripts"` plus only what
  `sandbox:` lists, and `referrerpolicy="no-referrer"`. Without
  `allow-same-origin` the app runs in an opaque origin: it cannot read Desk's
  DOM, cookies, storage or tokens, and cannot call Desk's API with Desk's
  credentials.
- **`allow-same-origin` with `allow-scripts`.** On the same origin as the
  page that frames it, this pair lets the app remove its own sandbox. Desk
  therefore drops `allow-same-origin` from any embed whose host name is the
  host the browser used to reach Desk (any port), and sends the reduced list.
  For an app on another host, `allow-same-origin` only gives that app its own
  origin back. Use it only for an app you trust with its own cookies.
- **Not a network filter.** The sandbox does not stop the app making network
  requests, and it does not replace the app's own authorization or CSRF
  defenses. An embedded app can call any server it can reach.
- **Desk's policy.** Desk's `Content-Security-Policy` stays as it was and
  gains `frame-src <origin> ...` on the page itself, listing only the
  configured origins. `frame-ancestors 'none'` and `X-Frame-Options: DENY`
  still apply, so Desk cannot be framed. The API and assets keep the old
  policy.
- **No bridge.** Desk sends no credentials, tokens, store data or messages
  into the frame and listens to none from it. There is no reverse proxy, no
  secret-in-URL support and no `postMessage` API.
- **Same-origin destinations.** Review any URL on Desk's own host, or on a
  host that shares its cookies. Prefer a separate host name. Desk's API
  trusts the browser session (the proxy's user header, or loopback access),
  so an app that runs as Desk's origin could call it. The sandbox's opaque
  origin is what prevents that, which is why Desk strips
  `allow-same-origin` there. Desk's CSRF header check also refuses writes
  from other origins. None of this makes an embed part of Desk.
- **HTTPS.** Prefer it. Plain `http` is allowed for loopback addresses and
  with `allow_http: true`. A page served over HTTPS cannot frame a remote
  `http` app: the browser blocks it as mixed content.

## Limits

- The destination must allow framing. Its own `X-Frame-Options` or
  `Content-Security-Policy: frame-ancestors` must permit Desk's origin.
  Otherwise the browser shows an empty or error frame.
- Cross-site cookies and sign-in may fail inside a frame, because browsers
  restrict third-party cookies. Fix that in the app or the browser.
- Desk cannot tell whether a frame loaded. `onload` fires for error pages and
  the parent cannot read a cross-origin load failure, so Desk shows a note and
  the `open externally` link and never claims success.
- Adding an embed with a new origin needs a page reload (the policy is sent
  with the page). Edits and removals apply live.

## Run the example

```sh
# 1. a fixture app and a Desk that frames it (isolated root, not live)
export ROOT=$(mktemp -d)
jobshop init "$ROOT"
cat >> "$ROOT/jobshop.yaml" <<'EOF'
desk:
  embeds:
    - id: demo
      label: Demo app
      url: http://localhost:7497/app
EOF
node ui/e2e/embed-fixture.mjs &     # serves http://localhost:7497/app
jobshop --root "$ROOT" serve &
jobshop --root "$ROOT" desk --listen 127.0.0.1:7480
# open http://127.0.0.1:7480/ and choose "Demo app"
```

See [the verification runbook](verify/desk-embeds.md) for the browser
acceptance run.
