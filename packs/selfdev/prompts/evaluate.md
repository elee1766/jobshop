Step `evaluate` of an `evaluate` run: the curator gate (#274). A curator
edit (a local memory write, or a merged prompt PR) was recorded with a
baseline: retries, escalations, review findings and recurrence of the
cited learnings in the scoped pools before it. Enough runs or days have
passed to compare.

1. `jobshop learn evaluate --runs {{.evaluate_runs}} --age {{.evaluate_age}}`
   decides every due edit and prints each with `state` (`helped` or
   `not_helped`), `reason`, `baseline` and `after`. It also writes the
   result on each cited learning (`jobshop learn list` shows `gate`).
   Nothing printed: comment `no edits due` and close.
2. For each `helped` edit: nothing more to do.
3. For each `not_helped` edit with a `path` (a local memory write, #592):
   the controller already restored the previous text, or removed the
   file when the write created it, and marked the cited learnings
   rejected. Check with `jobshop memory show` and `jobshop memory log`.
   Nothing more to do.
4. For each `not_helped` edit with a `pr` (a prompt PR), revert it
   through `pr-work`:
   - `gh pr view <pr> --json title,mergeCommit` for the merge commit.
   - File an issue on gfx-labs/jobshop titled `Revert curator edit
     <pr number>: <title>` whose body quotes the reason and the before
     and after numbers.
   - `git switch -c issue-<n>`, `git revert --no-edit -m 1 <merge sha>`
     (or without `-m 1` for a squash merge), run
     `go build ./... && go vet ./... && go test ./...`, then
     `git push -u origin HEAD` and `gh pr create --base master --title
     "Revert \"<title>\" (#<n>)" --body "Closes #<n>. This reverts commit
     <merge sha>. Curator gate: <reason>."`.
   - `gh issue comment <n> --body "PR <url> head <sha> is open"`, then
     `jobshop run start pr-work --var issue=<n> --key issue-<n>` (a
     trigger may have started it already: same run) and
     `jobshop comment -t <step id> "PR <url> head <sha>"` on its open
     step from `jobshop run show <run>`: `route` at first, `implement`
     once route closes.
   - Mark the cited learnings rejected with the original PR, so the
     curator does not propose the same edit again:
     `jobshop learn update <ids> --status rejected --pr <original pr url>`.

Never push to master, never merge. When done: `jobshop comment
"<edit ids>: <helped|not_helped>, restored <paths>, revert PRs <urls>"`
then `jobshop close`. If you cannot finish, `jobshop report failed
"<why>"` and `jobshop close -r failed`.
