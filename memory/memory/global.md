Jobshop agents never push to master: branch, open a PR, and let the
refinery merge it through the pr-work run. Keep replies and PR bodies
short and plain. When you find work outside your ticket, file an issue
instead of fixing it in the current PR.
Never pipe into `gh issue edit --body-file -`: a failed command upstream
(python3 and awk may be missing) blanks the body. Write the body to a file
with the Write tool, check it is non-empty, then pass the file.
After `gh pr checkout`, check `git rev-parse HEAD` equals the PR's
headRefOid; if the branch is checked out elsewhere it silently stays put.
Use `git checkout --detach <headRefOid>` instead.
