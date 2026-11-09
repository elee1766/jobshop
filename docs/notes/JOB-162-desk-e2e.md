# Desk browser checks in CI (JOB-162)

2026-11-09. `ui/e2e/run-isolated.mjs` runs the Playwright specs that need a
controller of their own: `ticket-menu`, `message-headers`, `embeds` and
`embed-reload`. The CI `e2e` job runs it on every PR and master push, after the
`ui` job.

- The script makes a root under the temp dir: a `type: chat` pool whose agent
  is the `pkg/control` test binary (`fake-agent`), a `link-acceptance` workflow
  that stays open (a command step exiting 75 with a 24h `retry_after`) and five
  `desk.embeds`. It starts `serve --tick=1s`, Desk on `127.0.0.1:7461` and
  `embed-fixture.mjs`, runs the specs and removes everything.
- `JOBSHOP_SOCKET`, `JOBSHOP_SESSION` and `JOBSHOP_ROOT` are dropped from the
  environment of every child, so a run from an agent session cannot reach a
  live controller.
- The job builds `jobshop` and `control.test` from the PR, so the Desk under
  test is the one in `pkg/desk/dist` from the `ui` job.
- The other specs keep the `DESK_MUTATE` and fixture guards and still skip
  when run against a plain Desk.

Manual check: [`docs/verify/desk-after-react.md`](../verify/desk-after-react.md).
