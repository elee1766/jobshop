Parallel PRs often conflict in docs/plan.md lists, and mergeStateStatus can
show CLEAN or UNKNOWN while stale. Right before merging, fetch master and
run `git merge-tree`; on conflict rebase, keep both list entries, push
with --force-with-lease and wait for CI again.
pkg/config and pkg/control change fast: `git fetch` and rebase onto
origin/master before the first push and reuse helpers master added. After
changing packs or a config default, run `go test ./pkg/config -run
TestPacksGolden -update` and commit the golden.
Parallel PRs often take the same migration number: before pushing a new
migration, fetch master and renumber past its highest file, then regenerate sqlc.
Deferred SQLite transactions get SQLITE_BUSY on the read-to-write upgrade
(busy_timeout does not wait on it): retry the whole transaction, bounded.
Bash calls auto-background after about 3s; redirects into the work tree,
scratch and session dirs are blocked: write files with the Write tool,
send logs to /tmp/<dir>/ or pipe to tail. For CI or deploy waits,
check once, comment the state, then `jobshop release --until 30m`.
Never commit `go test -c` binaries (`*.test`): check `git status`. Several sessions lint on one host: run
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
sqlc is not on PATH: regenerate with `go generate ./pkg/store/sqlite/`.
With `sqlc.slice`, named params after the slice bind into the list: put
named params first or use bare `?`, and check the `?N` in sqlcdb.
Old runs leave local `issue-<n>` branches: `git branch --list issue-<n>`
before `git switch -c`; if it exists use `issue-<n>-<slug>`. Do not
chain the switch with build/test, a failed switch skips them.
Local `-race` failures (dial timeouts) when /proc/loadavg is far above
the core count are load, not bugs: rerun the test alone before fixing.
A flaky test: `gh issue list --search <TestName>` first and link the
existing issue. In pkg/control tests, wait for every asserted condition
in one pass/passUntil, and run `pass(t, dir, 1)` before creating state
that startup (reapOrphans) must not touch.
