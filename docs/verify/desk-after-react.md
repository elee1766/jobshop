# Desk after the React port: context menu, message headers, iframe tabs

2026-11-09. Checks [JOB-136](ticket-menu.md), [JOB-137](message-headers.md)
and [JOB-151](desk-embeds.md) together on one isolated controller. CI runs the
automated part as the `e2e` job; this file is the same check by hand plus what
a script cannot judge.

## Automated

`ui/e2e/run-isolated.mjs` makes a throwaway root in the temp dir with the ACP
test agent as a chat pool, a held `link-acceptance` workflow and five
`desk.embeds`. It starts the controller (`--tick=1s`), Desk on
`127.0.0.1:7461` (`DESK_PORT` changes it) and the embed fixture, runs the specs
and removes everything. It drops `JOBSHOP_SOCKET`, `JOBSHOP_SESSION` and
`JOBSHOP_ROOT`, so it cannot reach a live controller even from an agent session.

```sh
npm --prefix ui ci && npm --prefix ui run build
go build -o /tmp/dar/jobshop ./cmd/jobshop
go test -c -o /tmp/dar/control.test ./pkg/orchestration/control
JOBSHOP_BIN=/tmp/dar/jobshop FAKE_AGENT_BIN=/tmp/dar/control.test \
  npm --prefix ui run test:e2e:isolated
```

Use the binary built from the branch: the one on `PATH` embeds the deployed
Desk. Browsers come from Playwright (`npx playwright install chromium`) or, on
gasserver, from nixpkgs as in `desk.md`.

Expected, desktop and mobile projects, 30 passed and 6 skipped (the five
width checks and the corner placement check are desktop only), then the three
reload checks:

- `ticket-menu.spec.js`: the menu opens from a home row (right click and the
  `⋯` button), a trace card and a relationship entry. `copy ticket id` copies
  the exact id and says `copied <id>`. `open in trace` goes to that ticket. A
  link keeps the browser's menu. Escape and an outside click close it and
  restore focus. It stays inside a 700x420 viewport.
- `message-headers.spec.js`: each human message has a `you` bar, an agent turn
  has one agent bar however many tool rows it has, and the read-only session
  viewer shows the same bars.
- `embeds.spec.js` and `embed-reload.spec.js`: tabs follow the built-in views
  in configured order, an embed opens in the content area with the header,
  status and chat rail, deep links, refresh and Back and Forward restore it,
  an unknown id or a URL makes no frame, the sandbox and CSP hold, and a
  config reload updates the tabs. Details: [desk-embeds.md](desk-embeds.md).

## By hand

The script removes its stack when it ends. For these, start an isolated
controller and Desk yourself with the root and `jobshop.yaml` the script
writes (copy them from `ui/e2e/run-isolated.mjs`), the embed fixture
(`node ui/e2e/embed-fixture.mjs`) and a few chats and runs, then look at what
assertions do not judge:

1. Home at 1100 and 412 px, light and dark theme: the `⋯` button is a usable
   touch target and the open menu is legible, with its id footer selectable.
2. Open a chat and send two messages: the `you` bar reads clearly with the
   gutter colors ignored, and a streaming reply still carries one agent bar.
3. Click through each embed tab: the label is readable at 600 px, the frame
   fills the content area and `open externally` opens a new tab.
4. Reload Desk on an embed route, then press Back: the previous view returns.

Record any difference as a ticket; do not change the specs to match it.
