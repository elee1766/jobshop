# Jobshop

Multi-agent work orchestration on a SQLite ticket store. Go 1.27, module
`github.com/elee1766/jobshop`. Design and decisions: `docs/plan.md`,
`docs/layout.md`, `docs/data-model.md`.

- Packages live in `pkg/`; `cmd/jobshop` only calls `cli.Main`.
- Every write goes through the controller; the CLI never opens the database.
- Check: `go build ./... && go vet ./... && go test ./... && golangci-lint run ./...`.
- Lint: golangci-lint v2.15.0 with `.golangci.yml`, the same version CI runs
  (on PATH for agents on gasserver; elsewhere `nix develop` or
  `nix run .#golangci-lint -- run ./...`). Fix every finding in the code you
  touched before pushing; do not add `//nolint` or relax `.golangci.yml`
  to pass. A finding in code you did not touch that blocks CI: fix it in
  the same PR and say so in the PR body.
- Store queries are sqlc (`pkg/store/sqlite/sqlc.yaml`); regenerate after
  editing `pkg/store/sqlite/queries/` or `migrations/`.
- Record design decisions in `docs/` with the date. In `docs/plan.md`: a
  decision goes in section 6, behavior as built in a new file
  `docs/notes/<issue>-<slug>.md` (short, dated), a new runbook check in a
  new file `docs/verify/<slug>.md`, and update the lists in sections 4
  and 5. Never add a numbered section to plan.md or gasserver-verify.md:
  parallel PRs take the same number and conflict (#131).
  Do not append long paragraphs to other sections.
- Plain names, short comments explaining why, no marketing language.

## Nix

CI also runs `nix build .#jobshop` (`nix/package.nix`, `buildGo127Module`).

- Any change to `go.mod` or `go.sum` changes the module vendor hash. Update
  `vendorHash` in `nix/package.nix` in the same commit, or the nix CI job
  fails with `hash mismatch in fixed-output derivation`.
- Get the new hash locally: set `vendorHash = lib.fakeHash;`, run
  `nix build .#jobshop 2>&1 | grep got:`, paste the `got:` value back. Then
  `nix build .#jobshop && ./result/bin/jobshop version` must pass.
- The package source is a fileset of `go.mod`, `go.sum`, `cmd/` and `pkg/`.
  A new top-level directory the binary needs (embedded files, testdata used
  by the build) must be added to that fileset.
- The build has `CGO_ENABLED=0`; do not add cgo dependencies.
- `nix/module.nix` is the NixOS module. If you add a config field the host
  sets through Nix, check the module still evaluates:
  `nix flake check --no-build`.
