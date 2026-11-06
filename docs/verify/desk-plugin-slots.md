# Verify: plugin-defined Desk slots (JOB-87)

Passed 2026-11-06 against a throwaway root (`jobshop init`, `meta_packs:
[e2e]`, `desk_commands.e2e-echo: {command: cat}`), with `serve` and
`desk --listen 127.0.0.1:7487`: `plugins.spec.js`, `desk.spec.js` and
`chat.spec.js`, 26 passed on desktop and Pixel 7 (the rest skip without
runs, sessions or a chat pool).

1. `cp -r ui/e2e/plugins/* <root>/desk/plugins/`, restart Desk.
2. `curl -s 127.0.0.1:7487/api/plugins | jq -r '.[] | "\(.name) \(.contributes|map(.slot)|join(","))"'`
   lists `e2e-slots` with `work:toolbar,work:aside,chat:panel,chat:composer-actions`
   and `e2e-slot-throws` with `work:toolbar`. A slot without a colon
   (`"slot": "toolbar"`) is logged as `desk: plugin.json` and skipped.
3. Run `docs/verify/desk.md` step 3 with `DESK_PLUGINS=1` and
   `DESK_URL` set to that Desk.

Pass: Work shows an "E2E action" button in its header and a card in the
right column; the `e2e-slot-throws` box in the toolbar is an error card
with `e2e slot import failure` and it is the only error card on the
page; the state chips still filter. Chat shows "E2E chat panel" under the
session list. `no horizontal page overflow` passes on mobile with the
fixtures installed.
