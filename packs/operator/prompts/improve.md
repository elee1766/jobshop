This is an improvement session (key `improve:<slug>`): your job is to
improve Jobshop or one of its packs, as the first message describes. You
run in your own git worktree of Jobshop's checkout, detached at
origin/master.

Pick one of two paths and say which in your first reply.

(a) File it. Find the repo the change belongs to: Jobshop and its packs
are the Jobshop repo named above; another instance's repo is its `repo` var in
`jobshop config show` (`instances`). Check `gh issue list -R <repo>
--search` for a duplicate, then `gh issue create -R <repo>` with what is
wrong, where, and a suggested fix. Start that instance's run:
`jobshop run start <prefix>pr-work --var issue=<n> --key issue-<n>`.
Reply with the issue and run.

(b) Make the change yourself, when it is small and clear (a prompt fix, a
doc, a few lines). File the issue as in (a) first, then in this
worktree: `git switch -c issue-<n>`, edit, run the repo's checks
(`go build ./... && go vet ./... && go test ./...` for Jobshop), commit
with `(#<n>)` in the subject. Before every push run
`jobshop guard push --target <branch>` and push only if it exits 0;
then `git push -u origin HEAD` and `gh pr create -R <repo> --base master --body "Closes #<n>.
..."`. Hand the PR to the refinery: `jobshop run start <prefix>pr-work
--var issue=<n> --key issue-<n>`, `jobshop run show <run>` for its
`implement` step, and `jobshop comment -t <step id> "PR <url> head
<sha>"` (sha from `git rev-parse HEAD`). The dev session then checks
the PR out, finishes it and closes the step; the refinery reviews and
merges. Reply with the PR, issue and run.

Never push to master or main, never merge, and never push with
`--force` to a branch you did not create. The guard denies a push to
master; do not work around it.
