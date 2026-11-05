Step `implement` of a `pr-work` run for issue #{{.issue}}.

Read the issue first: `gh issue view {{.issue}} --comments`. The issue is
the spec. Post progress, questions and design notes as issue comments,
not on the ticket.

If the ticket already has a PR (comments on this step or on the run's
`review-merge` step list refinery findings; `jobshop run show <run>` from
the ticket's `parent_id`): do not start a new branch. Check out the PR
head with `git fetch origin pull/<n>/head && git checkout --detach
FETCH_HEAD` (`gh pr checkout` fails in a detached worktree), then, as a
separate command, confirm `git rev-parse HEAD` equals the PR's
`headRefOid`. To push, `git push origin HEAD:<headRefName>`. Then
address each finding, reply to each review comment you resolve, push and
wait for green CI.

Otherwise: `git switch -c issue-{{.issue}}`, make the change, run
`{{.check_cmd}}`, commit with
`(#{{.issue}})` in the subject, `git push -u origin HEAD`, then
`gh pr create --base master --title "<summary> (#{{.issue}})" --body "Closes #{{.issue}}. ..."`.
PR body: at most ~10 lines, plain text: what changed, why, how it was
tested. No long summaries, file lists or pasted output.
Wait for CI with a wait, not a fixed deferral or polling: after each push,
once `gh pr view <n> --json headRefOid` shows the pushed sha (or the wait
records the old head), run
`jobshop wait add 'github.checks:pr={{.repo}}#<n>,state=success' 'github.checks:pr={{.repo}}#<n>,state=failure' --any --resume --timeout 1h --reason "CI on PR #<n>"`
and end the turn. The controller hands the step back to this session
within about a minute of CI finishing. On resume, read `gh pr checks <n>`:
green, comment and close as below; red, fix, push and set the wait again
(a new head clears the old wait). Do not use `release --until` for CI.
If the wait was cleared because the PR conflicts with its base, rebase on
`origin/master`, resolve, `push --force-with-lease`, and set the wait again.

Docs: {{.docs_rule}}

Do not merge. When CI is green on the head you pushed, and every
verification run the PR body cites has finished green:
`jobshop comment "PR <url> head <sha>"` (sha from `git rev-parse HEAD`),
then `jobshop close`. Closing `done` makes the `review-merge` step ready
for the refinery. If the remaining work can only happen later (a 24h
window, a scheduled time, a deploy you saw is in progress), do not close
failed: comment what you wait for, then `jobshop release --until
<RFC3339 or duration>`. That defers the step without counting a failure
or reaching a human. If you cannot deliver for any other reason,
`jobshop close -r failed` with a
comment saying why; the run retries and then escalates to a human.

Report state as you go, so stall-check and the planner read it instead
of your comments: `jobshop report progress "<what is done>"`, `jobshop
report blocked "<on what>"`, `jobshop report failed "<why>"` before
`close -r failed`. When only a person can decide (a decision, a
credential or access, a host config problem you cannot fix), ask once with
`jobshop report waiting "<what>" --question "<the decision needed, the
options, what you do if nobody answers>"`; the answer arrives as a
message in this session. Never ask a person for status or to say done.
