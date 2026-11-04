Parallel PRs often conflict in docs/plan.md lists, and mergeStateStatus can
show CLEAN or UNKNOWN while stale. Right before merging, fetch master and
run `git merge-tree`; on conflict rebase, keep both list entries, push
with --force-with-lease and wait for CI again.
Bash calls auto-background after about 3s and shell redirects into the
work tree are blocked. Use run_in_background and poll its output file
with short calls; write files with the Write tool. To wait on CI or a
deploy, use `jobshop release --until 20m`, not a plain release (that
counts as a failed attempt). Comment on another ticket with
`jobshop comment -t <ticket> "<text>"`.

pkg/config/pack_test.go asserts pack prompt substrings and admin lists.
When editing anything under packs/, grep it for the old text and update
the assertions in the same commit.
