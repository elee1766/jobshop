# NixOS module: config from your own repo (#532)

2026-11-05.

`services.jobshop.workspaces.<name>.config = "jobshop.yaml"` runs the
controller from a file in that workspace's checkout. A field on the
existing workspace, not a new `configRepo` option: workspaces already
clone, authenticate (`sshCommand`) and sync on a timer, so only the
check-and-reset step is new. At most one workspace, and not with
`configFile` or `settings` (assertions).

- `dataDir/jobshop.yaml` links to `<dir>/<config>`; the module's keys
  (identities, secret paths, gateway, sandbox) go to
  `jobshop.module.yaml` as with `configFile` (#531). The repo file
  lists it under `include:`, so secrets stay out of the repo.
- **Core change:** pack and `*_file` paths resolve against the config
  file's real directory (links followed), so `packs: [{path: packs/x}]`
  finds the checkout's packs although the root is `dataDir`. `include:`
  still resolves beside the root's file, where the module writes
  `jobshop.module.yaml`. A link into `/nix/store` (the generated file)
  keeps the link's directory, so managed mode is unchanged.
- **Sync** (`jobshop-workspace-<name>-sync`, timer every
  `syncInterval`): fetch, fast-forward, then `jobshop config check
  --file dataDir/jobshop.yaml`. Pass: write the commit to
  `dataDir/config-repo.last-good`, remove `config-repo.failed`, reload.
  Fail: `git reset --hard` to last-good, write the short commit and the
  error to `config-repo.failed`, exit 1, no reload. The same bad commit
  is not checked again until origin moves. A checkout that cannot
  fast-forward (a local edit) also writes `config-repo.failed`.
- **Attention:** the sync files a needs-human `config-repo:<name>`
  ticket with the commit and error, and closes it when a commit passes
  (#577; was `attention.config_repo_file`, now a load error).
- `jobshop-config.path` watches only `jobshop.module.yaml` in this mode;
  the sync unit owns reloads of the repo file.

Checked by `nix/module-test.nix` against a local bare repo: a good push
reloads, a broken pack resets and raises attention, a later good push
recovers, a local edit raises attention.
