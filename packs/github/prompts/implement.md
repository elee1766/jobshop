Step `implement` of a `pr-work` run for issue #${issue}.

Read the issue first: `gh issue view ${issue} --comments`. The issue is
the spec. Post progress, questions and design notes as issue comments,
not on the ticket.

If the ticket already has a PR (comments on this step or on the run's
`review-merge` step list refinery findings; `jobshop run show <run>` from
the ticket's `parent_id`): do not start a new branch. `gh pr checkout <n>`,
address each finding, reply to each review comment you resolve, push and
wait for green CI.

Otherwise: `git switch -c issue-${issue}`, make the change, run
`${check_cmd}`, commit with
`(#${issue})` in the subject, `git push -u origin HEAD`, then
`gh pr create --base master --title "<summary> (#${issue})" --body "Closes #${issue}. ..."`.
Wait for CI with `gh pr checks --watch`; fix failures and push again.

Do not merge. When CI is green on the head you pushed:
`jobshop comment "PR <url> head <sha>"` (sha from `git rev-parse HEAD`),
then `jobshop close`. Closing `done` makes the `review-merge` step ready
for the refinery. If you cannot deliver, `jobshop close -r failed` with a
comment saying why; the run retries and then escalates to a human.
