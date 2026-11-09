# Desk embedded apps: browser acceptance (JOB-151)

2026-11-09. Reference: [desk-embeds.md](../desk-embeds.md). Run against an
isolated controller, never a live one: the reload spec rewrites its
`jobshop.yaml`.

## Setup

```sh
go run ./pkg/rpc/rpcschema/cmd/deskgen
npm --prefix ui ci && npm --prefix ui run check && npm --prefix ui run build
go build -o /tmp/embeds/jobshop ./cmd/jobshop
env -u JOBSHOP_SOCKET -u JOBSHOP_SESSION -u JOBSHOP_ROOT /tmp/embeds/jobshop init /tmp/embeds/root
```

Replace `/tmp/embeds/root/jobshop.yaml` with:

```yaml
pack: pack.yaml
desk:
  embeds:
    - {id: cross, label: Cross app, url: "http://localhost:7497/app"}
    - {id: same, label: Same host, url: "http://127.0.0.1:7498/app", sandbox: [allow-same-origin]}
    - {id: denied, label: Denied XFO, url: "http://localhost:7497/deny"}
    - {id: ancestor, label: Denied CSP, url: "http://localhost:7497/ancestor"}
    - {id: extra-long-label-for-wrap, label: A rather long embed label, url: "http://localhost:7497/app"}
```

Start the controller (`serve --tick=1s`), Desk
(`desk --listen 127.0.0.1:7461`, and a second with `--listen 127.0.0.1:7462
--trusted-proxy 127.0.0.1` for the prefix run) and the fixture
(`node ui/e2e/embed-fixture.mjs`, ports 7497 and 7498). Each in its own
background process.

## Run

```sh
cd ui
DESK_URL=http://127.0.0.1:7461 npx playwright test embeds.spec.js
DESK_PREFIX=1 DESK_URL=http://127.0.0.1:7462 npx playwright test embeds.spec.js
EMBED_ROOT=/tmp/embeds/root EMBED_JOBSHOP=/tmp/embeds/jobshop \
  DESK_URL=http://127.0.0.1:7461 npx playwright test embed-reload.spec.js --project=desktop --workers=1
DESK_URL=http://127.0.0.1:7461 npx playwright test shell.spec.js desk.spec.js
```

Expected, desktop and mobile projects:

- top bar: built-in views, then the embeds in config order, then `catch-up`;
  at 1000, 899, 790, 700 and 600 px every tab is visible and inside the
  viewport, with no page overflow
- an embed opens in the content area with the header, status and chat rail,
  an iframe titled `<label> (embedded app)` and an `open externally` link
  (`target=_blank`, `rel=noopener`)
- deep link, refresh, Back and Forward restore the embed, also under `/desk/`
- `?view=embed&embed=<unknown or a URL>` shows a message and no iframe
- only one iframe exists. Leaving the embed ends the app (the fixture's tick
  count stops rising)
- in the frame: `window.origin` is `null`, parent document, location and
  storage, cookies, top navigation, popups and Desk's API are blocked. The
  request that loaded it carries no cookie, authorization, CSRF header or
  referrer
- the embed on Desk's own host loses `allow-same-origin`
- `X-Frame-Options: DENY` and `frame-ancestors 'none'` fixtures render no
  content, and Desk stays usable with its `open externally` link
- the page CSP keeps `default-src`, `script-src` and `frame-ancestors 'none'`,
  adds `frame-src` for the configured origins only, and an unconfigured frame
  origin raises a `frame-src` violation
- reload spec: removing the active embed drops the frame and tab with no
  page reload, a changed address replaces the frame, an invalid config is
  refused and changes nothing, an embed with a new origin says Desk must
  reload

Without `desk.embeds`, `embeds.spec.js` skips and the page and CSP match
the old ones.

## Rejected configuration

```sh
go test ./pkg/config -run DeskEmbeds
go test ./pkg/desk -run Embeds
go test ./pkg/control -run TestReloadPublishesStateForDesk
```

These cover the rejected cases (scheme, userinfo, secret keys, plain http,
sandbox and allow values, ids, labels, duplicates, packs), the CSP, the
same-host sandbox rule and the reload event.
