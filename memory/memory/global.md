Jobshop agents never push to master: branch, open a PR, and let the
refinery merge it through the pr-work run. Keep replies and PR bodies
short and plain. When you find work outside your ticket, file an issue
instead of fixing it in the current PR.
Never pipe into `gh issue edit --body-file -`: a failed command upstream
(python3 and awk may be missing) blanks the body. Write the body to a file
with the Write tool, check it is non-empty, then pass the file.
Do not trust `gh pr checkout` when the branch is checked out in another
worktree: it silently stays on the old HEAD. Run
`git fetch origin pull/<n>/head && git checkout --detach FETCH_HEAD`
(fetch that ref alone: with master first, FETCH_HEAD is master), then check `git rev-parse HEAD` equals the PR's headRefOid.
Wait for every verification run cited in the PR body to finish green
before `jobshop close`.
`jobshop close <id>` on a ticket you do not hold fails: `jobshop claim
<id>` first, or leave it to the operator.
If the jobshop CLI socket goes away mid-turn (controller restart), write
the PR head and CI state on the GitHub issue so the next attempt can
verify and close instead of redoing the work.
An umbrella or tracking issue whose sub-issues carry the work gets no
pr-work run: comment that on it instead of starting one.
`jobshop close -r failed` takes no reason text: comment first, then close.
`jobshop comment` takes one body argument; another ticket goes in `-t`:
`jobshop comment -t <ticket> "<text>"`, never `jobshop comment <id> <text>`.
