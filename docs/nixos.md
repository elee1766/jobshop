# NixOS

Moved from the README (2026-11-05, #551).

The flake exports `nixosModules.default`. It gives `services.jobshop` with
`settings` (the yaml), `workspaces.<name>` (repos cloned before start,
with live packs that follow the checkout), `gateway.<name>`,
`identities`, telemetry, and `selfDeploy`. Self-deploy promotes the
newest green master commit and rolls back on a failed health check. See
[`nix/module.nix`](../nix/module.nix), and [layout.md](layout.md) for
deployment. The sandbox option is in [isolation.md](isolation.md).

## Running from your own repo

Keep `jobshop.yaml` and `packs/` in a git repo and point a workspace at it:

```nix
services.jobshop.workspaces.config = {
  url = "https://github.com/me/jobshop-config.git";
  config = "jobshop.yaml";
};
```

The host links `dataDir/jobshop.yaml` to the checkout, so relative
`packs[].path` resolve in the repo. The file lists
`include: [jobshop.module.yaml]`, where the module puts identities and
secret paths: secrets stay out of the repo. Each push is fast-forwarded,
checked with `jobshop config check`, then reloaded. A commit that fails
is reset to the last good one and shows in `jobshop attention` as
`config-repo`, and so does a checkout a local edit keeps from moving.
See [notes/532-config-repo.md](notes/532-config-repo.md).
