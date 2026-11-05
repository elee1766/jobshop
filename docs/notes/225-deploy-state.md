# Self-deploy state follows the running binary (#225)

2026-11-03.

- Seen on gasserver: `deployed` said af9c010 while jobshop.service ran
  53b6b27. The file is written only after a healthy switch, so a
  rollback, a manual rebuild or auto-upgrade can leave it stale.
- Each self-deploy run now reads the running controller's version first.
  If `deployed` names another commit, it logs one WARN (the #227
  message, `<4>jobshop switched outside self-deploy (a rollback if
  older): running Y, last deployed X; deployed rewritten`) and rewrites
  the file with the running commit.
- The rewrite keeps the file's old mtime: the Deploys view (packs/deploy
  since #577) marks `failed` as rolled back while it is newer than
  `deployed`, and a fix-up is not a new good deploy. Attention no longer
  reads these files (#577): rollbacks are `deploy:rollback` tickets.
- Test: the rollback script test in `nix/module-test.nix`.
- Live check: `docs/verify/deploy-state.md`.
