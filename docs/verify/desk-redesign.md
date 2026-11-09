# Desk live delivery

2026-11-07. Run against an isolated controller before changing live work.

## Build and checks

```sh
npm --prefix ui ci
npm --prefix ui run check
npm --prefix ui run build
go build ./... && go vet ./... && go test ./...
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run ./...
node --test packages/jobshop-client/index.test.js
git diff --exit-code -- packages/jobshop-client/rpc.d.ts
nix fmt -- --check $(git ls-files '*.nix')
nix build .#jobshop
```

Regenerate the client types with `go run ./pkg/transport/rpc/rpcschema/cmd/deskgen`
before checking the generated diff. Nix CI checks formatting, packaging and
module evaluation when Nix is unavailable on the development host.

## Browser acceptance

Start a separate controller root with a project and no production credentials.
Serve its Desk on loopback, for example port 7488. `DESK_MUTATE` creates and
closes test tickets, so do not set it against a live work instance.

```sh
DESK_URL=http://127.0.0.1:7488 DESK_MUTATE=1 npm --prefix ui run test:e2e -- desk.spec.js live.spec.js --project=desktop --workers=2
DESK_URL=http://127.0.0.1:7488 DESK_PREFIX=1 DESK_MUTATE=1 npm --prefix ui run test:e2e -- desk.spec.js live.spec.js --project=desktop --workers=2
DESK_URL=http://127.0.0.1:7488 node ui/e2e/shots.mjs "$JCODE_SCRATCH_DIR/desk-shots"
```

Expected: layout at 1000, 790, 700, 412 and 360 pixels, direct view navigation,
blocked-storage rendering, no page overflow or browser errors. A real commit
updates Waiting and watched Trace without state GETs. A disconnected client
recovers a fresh snapshot. Ordered events replay only sequences after the
saved cursor, with matching SSE IDs and no duplicates. Inspect screenshots,
including representative work, not just empty states.

`TestNamedDeskChatLifecycle` checks actual controller restart, archive/restore,
retained journal and ACP identity, and successful send after restore.
`TestSessionSendSubscribe` owns journal replay, interrupt, cancel and close.
`TestPushControllerEvents` checks real commits produce notification and clear
messages without waiting for a periodic timer.

For a configured operator pool, opt into the real conversation browser check:
`DESK_CHAT=1 DESK_URL=<isolated-desk> npm --prefix ui run test:e2e -- chat.spec.js`.
This starts an agent and may consume provider credits. Without it, do not claim
browser/model round-trip acceptance.

## Deployed instance

Confirm the installed binary hash and active controller/Desk services before
browser checks. Do not enable mutating layout tests on selfdev. Load Home,
Trace, Improvements and System, inspect real Waiting/Inflight data, open a real
work Trace and verify connection recovery. Check DevTools shows persistent
`api/live`, not repeated overview/attention/chat requests. Recheck controller
and Desk services and logs after the deployment.

Real Pixel Fold/Gboard keyboard behavior and notification dismissal on an
actual subscribed phone remain separate device checks.
