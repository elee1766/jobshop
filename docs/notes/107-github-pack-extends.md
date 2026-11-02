# Generic github pack and pack extends (#107)

`packs/github` holds the plan, dev and refinery pools, the `replan` and
`master-red` triggers and the `pr-work`, `ci-repair` and `blocker-check`
workflows with their prompts. Pack vars: `repo` (owner/repo), `checkout`,
`agent`, `human`, `check_cmd`, and defaulted `ci_workflow`,
`dev_sessions`, `guide`, `deploy_check`. `pack.yaml` may set
`defaults:` (var values the including file did not set) and `extends:`
(a list like `packs:` whose `vars` are templates over the extending
pack's vars). Base layers merge before the extending pack, each recorded
in `sources` under its own file; nesting deeper than 8 fails as a cycle.
Pool `prompt_file` contents in a pack now get pack var substitution
(host prompt files stay literal). `packs/selfdev` is `extends: ../github`
with Jobshop's values; its own vars (`repo` = checkout path, `agent`,
`dev_sessions`) are unchanged, so hosts and the NixOS module need no
change. The merged selfdev config has the same pools, triggers,
workflows, steps and retries; prompt text differs only in wording that
named Jobshop or self-deploy directly.
