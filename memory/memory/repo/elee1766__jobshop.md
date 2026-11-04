Parallel PRs often conflict in docs/plan.md lists, and mergeStateStatus can
show CLEAN or UNKNOWN while stale. Right before merging, fetch master and
run `git merge-tree`; on conflict rebase, keep both list entries, push
with --force-with-lease and wait for CI again.
pkg/config and pkg/control change fast: `git fetch` and rebase onto
origin/master before the first push and reuse helpers master added. After
changing packs or a config default, run `go test ./pkg/config -run
TestPacksGolden -update` and commit the golden.
Bash calls auto-background after about 3s and shell redirects into the
work tree are blocked. Use run_in_background and poll its output file
with short calls; write files with the Write tool. To wait on CI or a
deploy, use `jobshop release --until 20m`, not a plain release (that
counts as a failed attempt). Several sessions lint on one host: run
`golangci-lint run --allow-parallel-runners ./...`.

pkg/config/pack_test.go asserts pack prompt substrings and admin lists.
When editing anything under packs/, grep it for the old text and update
the assertions in the same commit. nix/module-test.nix greps module
output: when changing nix/module.nix or removing strings, grep it and run
`nix build .#checks.x86_64-linux.module` (CI shows only "builder failed").
`jobshop run resume <run>` reopens only the first unfinished step. If a
later step (review-merge) failed, pass `--step <that step>` or the run
escalates again.
A PR with no checks usually conflicts with master: run
`git merge-tree --write-tree origin/master HEAD` and rebase before
pushing an empty commit.
