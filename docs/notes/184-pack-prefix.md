# Several repos on one host: pack prefixes (#184)

2026-11-03. A `packs:` entry in jobshop.yaml takes `prefix:` (lowercase,
up to 16 chars, e.g. `gz-`). Every pool, agent, trigger and workflow the
pack and its `extends:` bases declare is renamed `<prefix><name>`, and
the references between them are rewritten: pool `agent`, step `pool`,
`escalate_pool`, trigger `when.pool`, `ticket.pool` and `run.workflow`.
Names the pack does not declare (host agents and pools) stay. Packs see
`${pack_prefix}` ("" without a prefix); packs/github prompts use it in
`jobshop run start ${pack_prefix}pr-work` and `--pool ${pack_prefix}dev`.

So a second repo is a second packs/github entry:

    packs:
      - path: packs/selfdev
        vars: {...}
      - path: packs/github
        prefix: gz-
        vars: {repo: acme/gizmo, checkout: /src/gizmo, agent: a, human: wile, check_cmd: make test}

Each repo gets its own dev, plan, refinery and deploy-check pools and
its own planner, master-red and stall triggers. `issue_no_run` is
already repo-scoped (#183), and run keys are per workflow, so
`issue-5` in both repos does not collide. A run in one repo waits on a
run in another with the existing `jobshop edge add <implement step>
<other run's post-merge step>`; edges do not care about repos.

Not built here: shared pools keyed by repo, one planner across repos,
and the live check with a second throwaway repo.

Test: `TestGithubPackTwoRepos`.
