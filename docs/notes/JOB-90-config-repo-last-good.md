# Config repo sync compares HEAD with last-good (JOB-90)

2026-11-06. The workspace sync unit of a config repo (`workspaces.<name>.config`)
checks and reloads when HEAD differs from `dataDir/config-repo.last-good`, not
only when its own `git merge --ff-only` moved HEAD. A checkout advanced outside
the unit (an operator's `git pull` before a switch) used to skip the check and
the reload, leaving the controller on the old packs until a manual
`jobshop reload`. Tested in `nix/module-test.nix` (`configRepoTest`).
