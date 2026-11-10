# Self-dev pack cutover

2026-11-08. Scope: the `jobshop` user's self-development instance on
gasserver. Jobcity is out of scope.

The self-dev instance loads the selfdev, github (`acphost-`) and operator
packs as CUE from the installed checkout; there is no YAML pack left to
fall back to. Follow [migration](../configuration/migration.md) for the
config changes.

## Before activation

1. Drain open runs keyed `stall:...` and close their steps: after the
   upgrade they no longer hold a subject, and the `scan` trigger would
   start a second run for the same finding.
   `jobshop run list | jq -r '.[] | select(.title | test("stall")) | .id'`.
2. `jobshop config check --file <host.yaml>` with the candidate checkout
   (copy the host CUE root to a scratch dir and point its local module
   replacement at the candidate). It must load every `.star` file named by
   a workflow, trigger, scan signal, hook or desk command.
3. `jobshop identity check gfx-labs/jobshop` and `jobshop identity check
   elee1766/acphost` name the GitHub App identity and return a token.

## After activation

```sh readonly
jobshop version --server
jobshop trigger list | jq -r '.[].name' | sort | uniq -d   # empty: no duplicate triggers
jobshop trigger list | jq '.[] | select(.name == "scan" or .name == "acphost-scan") | {name, why, findings}'
jobshop automation list --limit 20 | jq -r '.[] | [.Key, .State] | @tsv'
```

Pass: the server runs the intended commit, each trigger appears once,
the scan triggers show findings or a reason, and recent handler, scan
script and hook invocations end `done` or `sync_done`, not `failed`.
`jobshop guard push --target master` from an operator session is denied.
Then the scan drill (`scan-drill.md`) and a live-check run, whose report
lands on its issue as an App comment.

Rollback is a binary and checkout revert; do not reset the ticket
database.

## Result (2026-11-08)

Deployed 0.9.0 (9c4f9bd) by hand: self-deploy's `config check` ran the old
host config (`stall_model`, `starlark_workflows`) against the new packs and
refused, as expected. All 13 triggers loaded once, `identity check` named the
`elee-agent` App for both repos, scan scripts ended `sync_done`,
`deploy-state` answered through the desk handler, and the next self-deploy
run was a no-op. Two pr-work runs open across the switch failed with `hook
manifest changed` and were cancelled (#1012).
