# Deploy state out of core (#577)

2026-11-05. Supersedes #575.

- Removed from the core: `attention.rollback_file` and
  `attention.config_repo_file` (attention kinds `rollback` and
  `config-repo`), Desk's `--deploy-state`, `/api/deploys`, its deploy
  attention item and the built-in Deploys plugin. A config that still
  sets either attention key loads with a warning naming `ticket upsert`
  (2026-11-05: an error blocked self-deploy, whose pre-switch check runs
  the new binary against the old module's yaml; make it an error once
  hosts run the new module).
- Desk's `/api/attention` is now the controller's `attention` RPC, so
  Desk, `jobshop attention`, gateways and digests list the same items.
- New generic `jobshop ticket upsert --key K [title] [-b body] [-l
  label] [--close]` (admin, RPC `ticket.upsert`): at most one open task
  per key and project (`meta.upsert.key`); a repeat updates it in place,
  `--close` closes it (title/body become a comment), a later report
  opens a new one.
- NixOS self-deploy: a failed config check upserts `deploy:config-check`,
  a failed switch or unhealthy controller upserts `deploy:rollback`, both
  `-l needs-human` with the rev and reason. A good deploy closes both.
  A skipped repeat re-files the saved report (`failed.report`), so a
  report lost while the controller was down still lands, with no second
  ticket. The state files stay as the script's own memory.
- Config repo sync (#532): `config-repo:<name>` the same way, closed when
  a commit passes.
- `services.jobshop.reportProject` picks the project; empty uses
  `default_project`.
- Deploys view: `packs/deploy`, a plain-JS plugin plus a `deploy-state`
  desk command that prints the `deployed`/`failed` revs and mtimes. The
  module adds the pack when `selfDeploy.enable` is set.
- Tests: `TestTicketUpsert`, `TestDeployPack`,
  `TestAttentionDeployKeysRemoved`, and the rollback and config-repo
  script tests in `nix/module-test.nix` (failure upserts, a repeat makes
  no second ticket, a good deploy closes it).
- Live check: `docs/verify/rollback-drill.md`.
