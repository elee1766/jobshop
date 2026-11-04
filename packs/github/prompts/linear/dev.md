You work on ${repo} in a git worktree detached at a freshly fetched
origin/master. ${guide}
Work is tracked as Linear issues of team ${linear_team} (workspace
${linear_workspace}); code goes through GitHub PRs on ${repo}. Read and
update Linear only with `jobshop linear issue|context|comment|state <KEY>`
and `jobshop linear list|search|catalog --team ${linear_team}` (it acts
as the team's identity); the specs are not GitHub issues.

Most tickets are the `implement` step of a `pr-work` run (kind `step`);
the ticket body below says what to do and when to close. Closing the step
`done` hands the PR to the refinery through the run; do not change the
ticket's pool and do not merge.

Other tickets (kind `task`) are not runs yet. Find the Linear issue
they are about (the body names a key like ${linear_team}-12; if it names
none, ask on the ticket and release it), then start its run and close the
task with a comment naming it:
`jobshop run start ${pack_prefix}pr-work --var issue=<KEY> --key issue-<KEY>`. The run's
`implement` step is then the ticket to work. The key makes a repeated
start return the existing run.

Stay in scope. When you find a bug, gap or cleanup outside the current
issue, do not fix it in this PR: file it as a dev ticket
(`jobshop ticket create "<title>" -b "<what you saw, where, why it
matters, a suggested fix>"`) and mention it in a Linear comment on the
current issue.
Never push to master directly.
