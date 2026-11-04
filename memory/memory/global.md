Jobshop agents never push to master: branch, open a PR, and let the
refinery merge it through the pr-work run. Keep replies and PR bodies
short and plain. When you find work outside your ticket, file an issue
instead of fixing it in the current PR.
Never pipe into `gh issue edit --body-file -`: a failed command upstream
(python3 and awk may be missing) blanks the body. Write the body to a file
with the Write tool, check it is non-empty, then pass the file.
Do not trust `gh pr checkout` when the branch is checked out in another
worktree: it silently stays on the old HEAD. Run
`git fetch origin pull/<n>/head && git checkout --detach FETCH_HEAD`,
then check `git rev-parse HEAD` equals the PR's headRefOid.
`jobshop close <id>` on a ticket you do not hold fails: `jobshop claim
<id>` first, or leave it to the operator.
If the jobshop CLI socket goes away mid-turn (controller restart), write
the PR head and CI state on the GitHub issue so the next attempt can
verify and close instead of redoing the work.
On an issue_no_run stall, run `gh pr list --state all --search <n>` first:
if a merged PR delivered it, close the issue instead of starting pr-work.
