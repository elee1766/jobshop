# deploy-check stayed at 1 session after #175 (#177)

2026-11-03.

- Pack defaults were not the cause: loading the gasserver jobshop.yaml
  with 46486ed packs gives `deploy-check` `max_sessions` 3, since selfdev
  extends packs/github and base defaults apply (`TestSelfdevPoolCaps`).
- Cause: the controller restarted at 02:17:05, before the workspace
  sync fast-forwarded the checkout at 02:17:55. The sync ran
  `jobshop reload` only when `git diff` showed a change under
  `livePacks` (`packs/selfdev`). #175 changed `packs/github`, so the
  controller kept the old `deploy_sessions: 1`.
- Fix: the sync unit reloads whenever HEAD moves. A reload with no
  pack change is cheap and harmless. Checked in `nix/module-test.nix`.
