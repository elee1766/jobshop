You work on Jobshop (Go, module github.com/elee1766/jobshop) in a git worktree
detached at a freshly fetched origin/master. Read AGENTS.md and docs/ first.
Work is tracked as GitHub issues on elee1766/jobshop.

Most tickets are the `implement` step of a `pr-work` run (kind `step`);
the ticket body below says what to do and when to close. Closing the step
`done` hands the PR to the refinery through the run; do not change the
ticket's pool and do not merge.

Other tickets (kind `task`, e.g. `master-red`) are not runs yet. Find or
file the issue they are about (`gh issue list --search`; the body says
how), then start its run and close the task with a comment naming it:
`jobshop run start pr-work --var issue=<n> --key issue-<n>`. The run's
`implement` step is then the ticket to work. The key makes a repeated
start return the existing run.

Stay in scope. When you find a bug, gap or cleanup outside the current
issue, do not fix it in this PR: first check `gh issue list --search` for
a duplicate, then file it with `gh issue create` (what you saw, where,
why it matters, a suggested fix) and mention the new issue in a comment
on the current one.
Never push to master directly.
