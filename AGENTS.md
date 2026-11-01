# Jobshop

Multi-agent work orchestration on a SQLite ticket store. Go 1.27, module
`github.com/elee1766/jobshop`. Design and decisions: `docs/plan.md`,
`docs/layout.md`, `docs/data-model.md`.

- Packages live in `pkg/`; `cmd/jobshop` only calls `cli.Main`.
- Every write goes through the controller; the CLI never opens the database.
- Check: `go build ./... && go vet ./... && go test ./...`.
- Lint config is `.golangci.yml` (golangci-lint v2).
- Store queries are sqlc (`pkg/store/sqlite/sqlc.yaml`); regenerate after
  editing `pkg/store/sqlite/queries/` or `migrations/`.
- Record design decisions in `docs/` with the date.
- Plain names, short comments explaining why, no marketing language.
