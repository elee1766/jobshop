Step `implement` of a `pr-work` run for Linear issue {{.issue}}.

Read the issue first: `jobshop linear issue {{.issue}}` (JSON with
description and comments). The issue is the spec; `jobshop linear
context {{.issue}}` adds labels, blockers, parent and children. Post progress,
questions and design notes as Linear comments
(`jobshop linear comment {{.issue}} "<text>"`), not on the ticket.

If the ticket already has a PR (comments on this step or on the run's
`review-merge` step list refinery findings; `jobshop run show <run>` from
the ticket's `parent_id`): do not start a new branch. `gh pr checkout <n>`,
address each finding, reply to each review comment you resolve, push and
wait for green CI.

Otherwise: move the issue to In Progress
(`jobshop linear state {{.issue}} "In Progress"`; if that name is not one
of the team's states, the error lists them, pick the started one),
`git switch -c issue-{{.issue}}`, make the change, run
`{{.check_cmd}}`, commit with
`({{.issue}})` in the subject, `git push -u origin HEAD`, then
`gh pr create --base master --title "<summary> ({{.issue}})" --body "Fixes {{.issue}}. ..."`.
PR body: at most ~10 lines, plain text: what changed, why, how it was
tested. No long summaries, file lists or pasted output.
Comment the PR link on the issue:
`jobshop linear comment {{.issue}} "PR <url>"`.
Wait for CI with a wait, not a fixed deferral or polling: after each push
run
`jobshop wait add 'github.checks:pr={{.repo}}#<n>,state=success' 'github.checks:pr={{.repo}}#<n>,state=failure' --any --resume --timeout 1h --reason "CI on PR #<n>"`
and end the turn. The controller hands the step back to this session
within about a minute of CI finishing. On resume, read `gh pr checks <n>`:
green, comment and close as below; red, fix, push and set the wait again
(a new head clears the old wait). Do not use `release --until` for CI.

Docs: {{.docs_rule}}

Do not merge. When CI is green on the head you pushed:
`jobshop comment "PR <url> head <sha>"` (sha from `git rev-parse HEAD`),
then `jobshop close`. Closing `done` makes the `review-merge` step ready
for the refinery. If the remaining work can only happen later (a 24h
window, a scheduled time, a deploy you saw is in progress), do not close
failed: comment what you wait for, then `jobshop release --until
<RFC3339 or duration>`. That defers the step without counting a failure
or reaching a human. If you cannot deliver for any other reason,
`jobshop close -r failed` with a
comment saying why; the run retries and then escalates to a human.
