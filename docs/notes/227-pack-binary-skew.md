# Binary and live packs cannot diverge (#227)

2026-11-03. A manual switch installed an older jobshop than the live
packs needed; `jobshop serve` failed config load and crash-looped.

Built:
- After every good load or reload the controller writes the merged
  config to `<root>/config.last-good.yaml`. On start, an error inside a
  pack (`config.ErrPack`: a pack file failed to load, or validation
  names a pack file) falls back to that file: the controller serves,
  logs ERROR, and `jobshop attention` lists a `config` item until a
  reload succeeds. Host config errors, and a pack error with no last
  good file, still exit.
- The workspace sync (selfDeploy workspace) reads the commit from the
  generation's installed package (`${jobshop} version`), not
  `/proc/<MainPID>`, so a stopped or crash-looping controller still
  holds the checkout back. An unknown commit, or one missing from the
  checkout, leaves the checkout unmoved.
- Self-deploy logs a WARN when the running commit differs from
  `$state/deployed` (a switch outside self-deploy, a rollback if older).

Tests: `TestLastGood`, `TestStartWithBadPack`, module-test checks "sync
follows the installed binary" and "self-deploy logs a switch outside it".
Live: `docs/verify/pack-binary-skew.md`.
