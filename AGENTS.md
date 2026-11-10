# Jobshop

Multi-agent work orchestration on a SQLite ticket store. Go 1.27, module
`github.com/gfx-labs/jobshop`. Design: `docs/plan.md` (frozen archive,
up to 2026-11-05), `docs/layout.md`, `docs/data-model.md`. Decisions:
the pinned decision log issue (#569). Progress: open `target` issues.

- Packages live in `pkg/`; `cmd/jobshop` only calls `cli.Main`.
- Fixed JSON envelopes use Go structs with JSON tags. Maps are for dynamic keys
  and arbitrary payload data.
- Every write goes through the controller; the CLI never opens the database.
- Generate Desk assets first: `npm --prefix ui ci && npm --prefix ui run build`.
  `pkg/transport/desk/dist` is ignored build output, not committed source. Rebuild it after
  frontend changes, then check:
  `go build ./... && go vet ./... && go test ./... && golangci-lint run ./...`,
  plus `nix fmt -- --check $(git ls-files '*.nix')` (CI runs it; `nix fmt` fixes it).
- Tests must not depend on wall time: no `time.Sleep` (lint bans it),
  run a supervisor pass (`pass` / `Options.Pass`) or wait on an event.
  New controller-logic tests use the in-process agent once #361 lands.
- Lint: golangci-lint v2.15.0 with `.golangci.yml`, the same version CI runs
  (on PATH for agents on gasserver, set in its host config; elsewhere
  `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run ./...`).
  Fix every finding in the code you
  touched before pushing; do not add `//nolint` or relax `.golangci.yml`
  to pass. A finding in code you did not touch that blocks CI: fix it in
  the same PR and say so in the PR body.
- Store queries are plain SQL in Go, no sqlc (#579): collections and
  `List` build theirs (bqb), the rest are hand-written in
  `pkg/storage/store/sqlite/sqldb` and `pkg/storage/store/postgres/pgdb`, one per engine.
  Goose migrations stay in `pkg/storage/store/{sqlite,postgres}/migrations`.
- Do not edit `docs/plan.md`: it is frozen (2026-11-05) because parallel
  PRs kept conflicting on it. Record a design decision as one dated
  comment on the decision log issue #569 (format in its body). Behavior
  as built goes in a new file `docs/notes/<issue>-<slug>.md` (short,
  dated), a new runbook check in a new file `docs/verify/<slug>.md`.
  Never add a numbered section to gasserver-verify.md (#131).
  Do not append long paragraphs to other docs.
- Plain names, short comments explaining why, no marketing language.
- Desk UI work: read `ui/AGENTS.md` and `docs/desk-design.md` before changing
  layout, styling, interaction or user-facing copy.

## Nix

CI also runs `nix build .#jobshop` (`nix/package.nix`). With go2nix's plugin
loaded (`plugin-files`, from `nix build .#go2nix-plugin`, Nix 2.34+) it builds
one derivation per Go package; without it, one `buildGo127Module` derivation.

- Any change to `go.mod` or `go.sum` changes the fallback builder's vendor
  hash. Update `vendorHash` in `nix/package.nix` in the same commit, or the
  nix CI job fails with `hash mismatch in fixed-output derivation`. go2nix
  needs nothing: it reads `go.sum`.
- Deleting or adding files that import a module (tests included) can change
  it too, with no `go.mod` change. A local build reuses a cached output
  for the old hash; `nix build .#jobshop.goModules --rebuild` shows the
  real `got:` value.
- Track new Go source files before calculating the hash: vendoring follows
  source imports, and Nix's Git source omits untracked files.
- Get the new hash locally, without the plugin: set `vendorHash =
  lib.fakeHash;`, run `nix build .#jobshop.goModules 2>&1 | grep got:`, paste the `got:` value back. Then
  `nix build .#jobshop && ./result/bin/jobshop version` must pass.
- The package source is a fileset of `go.mod`, `go.sum`, `cmd/` and `pkg/`.
  A new top-level directory the binary needs (embedded files, testdata used
  by the build) must be added to that fileset.
- Desk assets are built by `nix/desk-ui.nix`, then supplied through go2nix's
  `srcOverlay` or the fallback's `postPatch`. Changes to `ui/package-lock.json`
  require updating its `npmDepsHash`: set `lib.fakeHash`, run
  `nix build .#desk-ui`, and use the reported hash. Keep new frontend build inputs
  in that derivation's fileset.
- The build has `CGO_ENABLED=0`; do not add cgo dependencies.
- `nix/module.nix` is the NixOS module. If you add a config field the host
  sets through Nix, check the module still evaluates:
  `nix flake check --no-build`.
