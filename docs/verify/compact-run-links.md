# Compact run references

2026-11-07

Run only against an isolated controller. Define a disposable workflow accepting
`repo`, `pr` and `issue` variables that stays open long enough to inspect. For
example, a command step with `run: 'false'` reaches its human-review failure state
without contacting an outside service. Do not use a workflow that changes a repo
or performs external actions.

After building the frontend, rebuilding the Go binary and restarting Desk:

```sh
DESK_URL=http://127.0.0.1:7492 DESK_MUTATE=1 \
  DESK_LINK_WORKFLOW=link-acceptance \
  npm --prefix ui run test:e2e -- run-links.spec.js --workers=1
```

The spec starts and cancels a real run. It records explicit Linear workspace
metadata and checks the Home issue and PR destinations on desktop and mobile.
When the compact layout hides the PR column, the same PR is linked beneath the
ticket title. The test does not navigate to external sites.
