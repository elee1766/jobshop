# Self-deploy out of jobshop (JOB-74, 2026-11-06)

- Deploying jobshop is the host's job. Removed from the module: the
  `services.jobshop.selfDeploy.*` options (now `mkRemovedOptionModule`
  errors pointing here), the per-workspace `selfDeploy` flag (an assertion
  with the same pointer), the `jobshop-self-deploy` unit and timer, the
  polkit rule for the trigger start (#826), the `nixos-upgrade` override
  flags, `packs/deploy` (Deploys view, `deploy-state` desk command,
  `self-deploy` trigger) and their tests and runbooks.
- New `workspaces.<name>.followInstalled` (#840): the sync unit holds the
  checkout at the installed binary's commit (#227) instead of origin. A host
  that deploys jobshop sets it.
- Kept, as generic pieces a deployer uses: `jobshop ticket upsert` (#577),
  event triggers with `command:` (JOB-62), `jobshop config check`, version
  and health RPCs, and the github pack's `deploy-check` step, which only
  watches for a deploy.
- gasserver's deploy now lives in gasserver.nix `home/jobshop-deploy.nix`
  with the Deploys view and trigger in `home/jobshop-deploy/`: same logic
  (green-commit pick, config check, home-manager switch, health check,
  rollback, drill, `deploy:*` tickets), owned by the operator. Jobshop
  agents do not change it.
