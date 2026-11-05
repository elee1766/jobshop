Jobshop repo notes (packs/selfdev, #592). These hold on every
elee1766/jobshop turn, alongside AGENTS.md.

- pkg/config and pkg/control change fast: `git fetch` and rebase onto
  origin/master before the first push, and reuse helpers master added.
- After every rebase, rerun the generators even if GitHub shows the PR
  clean: `go generate ./pkg/store/sqlite/` (sqlc is not on PATH),
  `go run ./pkg/rpc/rpcschema/cmd/deskgen`, then `go test ./...`.
- Parallel PRs take the same migration number: before pushing a new
  migration, fetch master, renumber past its highest file in
  `pkg/store/sqlite/migrations` (and `pkg/store/postgres/migrations`),
  then regenerate sqlc.
- With `sqlc.slice`, named params after the slice bind into the list: put
  named params first or use bare `?`, and check the `?N` in sqlcdb.
- Deferred SQLite transactions get SQLITE_BUSY on the read-to-write
  upgrade (busy_timeout does not wait on it): retry the whole
  transaction, bounded.
- Tests do not assert pack contents (JOB-2): editing packs/ needs no
  test change unless the pack stops loading (`TestPacksLoad`). A new
  pack gets minimal vars in `TestPacksLoad`.
- A PR that removes or renames a config field the NixOS module sets must
  keep it as a load warning (ignored) for one release, not an error: the
  self-deploy config check runs the new binary on the jobshop.yaml the
  old module wrote (#597). Turn it into an error in a later PR.
- `nix/module-test.nix` greps module output: after changing
  nix/module.nix, grep it and run
  `nix build .#checks.x86_64-linux.module` (CI shows only "builder failed").
- Never commit `go test -c` binaries (`*.test`); check `git status`.
  Several sessions lint on one host: `golangci-lint run
  --allow-parallel-runners ./...`.
- Local `-race` failures (dial timeouts) while /proc/loadavg is far above
  the core count are load: rerun the test alone before fixing. In
  pkg/control tests, wait for every asserted condition in one
  pass/passUntil, and run `pass(t, dir, 1)` before creating state that
  startup (reapOrphans) must not touch.
- Sandbox and isolation tests skip silently without bwrap, newuidmap,
  pasta or nft on PATH: run them under `PATH=$PATH:/run/wrappers/bin nix
  shell nixpkgs#bubblewrap nixpkgs#passt nixpkgs#nftables -c go test -v
  ...` and check the output for SKIP.
- npm is not on agent PATH: in ui/ run `nix shell nixpkgs#nodejs_22 -c sh
  -c 'npm ci && npm run check && npm run build'`.
- Bash calls background after about 3s, and redirects into the work tree,
  scratch and session dirs are blocked: write files with the Write tool,
  send logs to /tmp/<dir>/ or pipe to tail.
