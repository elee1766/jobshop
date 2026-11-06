Step `implement` of a `pr-work` run for Linear issue {{.issue}}.

Read the issue first: `jobshop linear issue {{.issue}}` (JSON with
description and comments). The issue is the spec; `jobshop linear
context {{.issue}}` adds labels, blockers, parent and children. Post progress,
questions and design notes as Linear comments
(`jobshop linear comment {{.issue}} "<text>"`), not on the ticket. A
question only a person can answer, asked on the issue some other way
than `jobshop report waiting --question` (which labels it for you when
the host sets `linear.ask_on_issue: true`):
`jobshop linear label {{.issue}} add Blocked`, so intake leaves the issue
alone until it is answered.

First decide what kind of work it is. Every exit closes with an outcome:
- **Code change**: the steps below, ending `jobshop close --outcome pr`.
- **Task with no code** (investigate, run a check, answer a question,
  operate something this session may operate): do it, post the result
  as a Linear comment (a long result via `jobshop artifact deliver`,
  linked), move the issue to `{{.linear_done_state}}` (`jobshop linear
  state {{.issue}} "{{.linear_done_state}}"`), then `jobshop close
  --outcome done`. Review and post-merge are skipped.
- **Needs a person** (a decision, credentials, access this session
  lacks): `jobshop linear ask {{.issue}}` with the options and what you do
  by default, and `jobshop report waiting "<what>" --question "<...>"
  --context "<...>" --option "<Title>|<what it does>"`.
  With `linear.ask_on_issue: true` the question labels the issue Blocked.
- **Not work** (a tracking, parent or target issue, a duplicate): comment
  why on the issue, add the label `{{.linear_skip_labels}}` or move it to
  Backlog, then `jobshop close --outcome done`. Never open a PR for it.
Say in your last Linear comment which of these it was.

If the ticket already has a PR (a "Changes requested" section in this
prompt, or comments on this step or on the run's
`review-merge` step list refinery findings; `jobshop run show <run>` from
the ticket's `parent_id`): do not start a new branch. `gh pr checkout <n>`,
address each finding, reply to each review comment you resolve, push and
wait for green CI.

Otherwise: take the issue: assign it to yourself and move it to In
Progress in one call
(`jobshop linear assign {{.issue}} me --state "In Progress"`;
if that state name is not one of the team's states, the error lists
them, pick the started one; if it fails only on the assignee
read-back, `jobshop linear issue {{.issue}}` shows the state moved,
carry on),
`git switch -c issue-{{.issue}}`, make the change, run
`{{.check_cmd}}`, commit with
`({{.issue}})` in the subject, `git push -u origin HEAD`, then
`gh pr create --base master --title "<type>(<scope>): <summary> ({{.issue}})" --body "Fixes [{{.issue}}](https://linear.app/{{.linear_workspace}}/issue/{{.issue}}). ..."`.
The PR title is a conventional commit: it becomes the squash commit
subject. Type is one of feat, fix, perf, refactor, docs, test, ci,
build, chore, revert; scope is the package or area (`control`, `desk`,
`store`, `packs/github`); the summary starts lowercase. A breaking
change to the CLI, config, protocol or store schema gets `!` after the
scope (`feat(store)!: ...`) or a `BREAKING CHANGE:` line in the body.
PR body: at most ~10 lines, plain text: what changed, why, how it was
tested. No long summaries, file lists or pasted output.
Comment the PR link on the issue:
`jobshop linear comment {{.issue}} "PR <url>"`.
Wait for CI with a wait, not a fixed deferral or polling: after each push
run
`jobshop wait add 'github.checks:pr={{.repo}}#<n>,state=success,workflow={{.ci_workflow}}' 'github.checks:pr={{.repo}}#<n>,state=failure,workflow={{.ci_workflow}}' --any --resume --timeout 1h --reason "CI on PR #<n>"`
and end the turn. The controller hands the step back to this session
within about a minute of CI finishing. On resume, read `gh pr checks <n>`:
green, comment and close as below; red, fix, push and set the wait again
(a new head clears the old wait). Do not use `release --until` for CI.
A wait that ends `invalidated: ... cancelled` is not red: the run was
superseded or cancelled, so rerun it (`gh run rerun <id>`) and wait again.
If the wait was cleared because the PR conflicts with its base, rebase on
`origin/master`, resolve, `push --force-with-lease`, and set the wait again.

PR in another instance's repo (JOB-33): when the PR you opened is not in
{{.repo}}, this run's review-merge cannot merge it; that repo's own
instance reviews it with its `<prefix>pr-review` workflow (`<prefix>` is
that repo's `pr-work` prefix: find a run whose `meta.github.repo` is that
repo in `jobshop run list` and strip `pr-work` from its
`fields.workflow`). When CI is green, start it:
`jobshop run start <prefix>pr-review --var pr=<n> --key pr-<n>` (a
repeated start returns the open run), comment
`PR <url> head <sha> review run <run id>` on this step, then
`jobshop close --outcome pr`. No such workflow on this controller:
`jobshop report waiting "no pr-review for <repo>" --question "Which
instance reviews <url>?" --context "<the PR, its head, the run list you
checked>"`.

Docs: {{.docs_rule}}

Do not merge. When CI is green on the head you pushed:
`jobshop comment "PR <url> head <sha>"` (sha from `git rev-parse HEAD`),
then `jobshop close --outcome pr`. The outcome `pr` makes the `review-merge` step ready
for the refinery (a close without it skips review and merge). If the
remaining work can only happen later, do not
close failed: comment what you wait for and park on it. An event has a
wait kind: a deploy in progress is `jobshop wait add
'deploy:repo={{.repo}},sha=<sha>' --resume -r "deploy of <sha>"`, a
merge `github.pr:pr=...,state=merged`, a tag `github.tag:repo=...,tag=...`,
another run `run:id=<run>,state=done`, a Linear issue
`linear.state:issue=<KEY>,state=Done`. Only a real time (a 24h window,
a scheduled time) is `jobshop release --until <RFC3339 or duration>
--time-wait -r "<why>"`. Either defers the step without counting a failure
or reaching a human. If you cannot deliver for any other reason,
`jobshop close -r failed` with a
comment saying why; the run retries and then escalates to a human.
