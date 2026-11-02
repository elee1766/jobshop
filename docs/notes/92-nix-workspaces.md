# Workspaces in the NixOS module (#92)

`services.jobshop.workspaces.<name>` has `url`, `dir` (default
`<dataDir>/workspaces/<name>`), `sshCommand`, `livePacks`,
`livePackVars`, `syncInterval` and `selfDeploy`. Each workspace with a
url gets `jobshop-workspace-<name>` (clone before the controller) and
`jobshop-workspace-<name>-sync` (fast-forward, reload on pack changes;
timer only with live packs). Live packs of all workspaces go into the
generated `packs:` with their own vars. At most one workspace sets
`selfDeploy` (assertion); `selfDeploy.enable` needs one. That workspace
gives the repo for the green-commit pick, and its packs follow the
deployed commit. The old `workspace` option still evaluates: it maps to
`workspaces.default` with the old dir (`<dataDir>/src/<repo>`) and
`selfDeploy = selfDeploy.enable`, so gasserver keeps its checkout; only
the unit names change (`jobshop-workspace-default[-sync]`). `checks.module`
in the flake evaluates a two-workspace config, the two-selfDeploy
assertion and the old option.
