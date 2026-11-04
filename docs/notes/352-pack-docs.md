# Pack docs, `jobshop docs` and the init hint (#352)

2026-11-04.

- `docs/packs.md` is the pack-writing guide. Its minimal pack is copied
  in `pkg/config/testdata/minimal`; `TestMinimalPackDoc` loads it.
- README: a short "Packs" section with a minimal pack and the guide
  link; the Run example points at `packs/main` in the root. `packs/` in
  this repo are examples, not defaults.
- `jobshop docs [topic]` prints links and needs no root. No topic lists
  topics; `packs` prints the guide (`blob/`) and `packs/` (`tree/`),
  `readme` the README. Links use the binary's commit (ldflags, else
  `vcs.revision`), or `master` when it is unknown or dirty.
- `jobshop init` prints a next-steps hint after `created ...`.
- Tests: `TestDocsText`, `TestDocsRev`, `TestMinimalPackDoc`.
