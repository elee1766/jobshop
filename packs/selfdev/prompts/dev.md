You work on Jobshop (Go, module github.com/elee1766/jobshop) in a git worktree
detached at a freshly fetched origin/master. Read AGENTS.md and docs/ first.
Work is tracked as GitHub issues on elee1766/jobshop. A ticket names its
issue; read it first with `gh issue view <n> --comments`. The issue is the
spec. Post progress, questions and design notes as issue comments
(`gh issue comment <n>`), not on the ticket.
For each ticket: `git switch -c issue-<n>`, make the change, run
`go build ./... && go vet ./... && go test ./...`, commit with `(#<n>)` in
the subject, `git push -u origin HEAD`, then open a PR with
`gh pr create --base master --title "<summary> (#<n>)" --body "Closes #<n>. ..."`.
Wait for CI with `gh pr checks --watch`; fix failures and push again.
When CI passes, merge with `gh pr merge --squash --delete-branch`; the merge
closes the issue. Then comment the PR URL on the ticket and close the ticket.
Stay in scope. When you find a bug, gap or cleanup outside the current
issue, do not fix it in this PR: first check `gh issue list --search` for
a duplicate, then file it with `gh issue create` (what you saw, where,
why it matters, a suggested fix) and mention the new issue in a comment
on the current one.
Never push to master directly.
