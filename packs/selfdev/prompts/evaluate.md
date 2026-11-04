Step `evaluate` of an `evaluate` run: the curator gate (#274). A curator
PR merged a memory, skill or prompt edit; the controller recorded its
baseline (retries, escalations, review findings and recurrence of the
cited learnings in the scoped pools before the merge). Enough runs or
days have passed to compare.

1. `jobshop learn evaluate --runs {{.evaluate_runs}} --age {{.evaluate_age}}`
   decides every due edit and prints each with `state` (`helped` or
   `not_helped`), `reason`, `baseline` and `after`. It also writes the
   result on each cited learning (`jobshop learn list` shows `gate`).
   Nothing printed: comment `no edits due` and close.
2. For each `helped` edit: nothing more to do.
3. For each `not_helped` edit, revert it through `pr-work`:
   - `gh pr view <pr> --json title,mergeCommit` for the merge commit.
   - File an issue on elee1766/jobshop titled `Revert curator edit
     <pr number>: <title>` whose body quotes the reason and the before
     and after numbers.
   - `git switch -c issue-<n>`, `git revert --no-edit -m 1 <merge sha>`
     (or without `-m 1` for a squash merge), run
     `go build ./... && go vet ./... && go test ./...` and
     `jobshop memory check memory/memory`, then `git push -u origin HEAD`
     and `gh pr create --base master --title "Revert \"<title>\" (#<n>)"
     --body "Closes #<n>. This reverts commit <merge sha>. Curator gate:
     <reason>."`.
   - `jobshop run start pr-work --var issue=<n> --key issue-<n>`, find its
     `implement` step with `jobshop run show <run>` and
     `jobshop comment -t <step id> "PR <url> head <sha>"`.
   - Mark the cited learnings rejected with the original PR, so the
     curator does not propose the same edit again:
     `jobshop learn update <ids> --status rejected --pr <original pr url>`.

Never push to master, never merge. When done: `jobshop comment
"<edit ids>: <helped|not_helped>, revert PRs <urls>"` then `jobshop
close`. If you cannot finish, `jobshop report failed "<why>"` and
`jobshop close -r failed`.
