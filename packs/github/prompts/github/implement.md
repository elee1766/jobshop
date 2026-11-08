Step `implement` of a `pr-work` run for issue #{{.issue}}.

Read the issue first: `gh issue view {{.issue}} --comments`. The issue is
the spec. Post progress, questions and design notes as issue comments,
not on the ticket.

First decide what kind of work it is. Every exit closes with an outcome:
- **Code change**: the steps below, ending `jobshop close --outcome pr`.
- **Task with no code** (investigate, run a check, answer a question,
  operate something this session may operate): do it, post the result
  with `gh issue comment {{.issue}}` (a long result via `jobshop artifact
  deliver`, linked), close the issue (`gh issue close {{.issue}}
  --comment "<result>"`), then `jobshop close --outcome done`. Review
  and post-merge are skipped.
- **Needs a person** (a decision, credentials, access this session
  lacks): `jobshop report waiting "<what>" --question "<the decision>"
  --context "<what you did, found, why you cannot decide, what each
  answer leads to>" --option "<Title>|<what it does>" --recommend N`.
- **Not work** (a tracking, parent or target issue, a duplicate): comment
  why on the issue (close a duplicate), then `jobshop close --outcome
  done`. Never open a PR for it.
Say in your last issue comment which of these it was.

If the ticket already has a PR (a "Changes requested" section in this
prompt, or comments on this step or on the run's
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
`gh pr create --base master --title "<type>(<scope>): <summary> (#{{.issue}})" --body "Closes #{{.issue}}. ..."`.
The PR title is a conventional commit: it becomes the squash commit
subject. Type is one of feat, fix, perf, refactor, docs, test, ci,
build, chore, revert; scope is the package or area (`control`, `desk`,
`store`, `packs/github`); the summary starts lowercase. A breaking
change to the CLI, config, protocol or store schema gets `!` after the
scope (`feat(store)!: ...`) or a `BREAKING CHANGE:` line in the body.
PR body: at most ~10 lines, plain text: what changed, why, how it was
tested. No long summaries, file lists or pasted output.
Wait for CI with a wait, not a fixed deferral or polling: after each push,
once `gh pr view <n> --json headRefOid` shows the pushed sha (or the wait
records the old head), run
`jobshop wait add 'github.checks:pr={{.repo}}#<n>,state=success,workflow={{.ci_workflow}}' 'github.checks:pr={{.repo}}#<n>,state=failure,workflow={{.ci_workflow}}' --any --resume --timeout 1h --reason "CI on PR #<n>"`
and end the turn. The controller hands the step back to this session
(or, if it is busy by then, to any idle session of the pool) within
about a minute of CI finishing. On resume, read `gh pr checks <n>`:
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

Do not merge. When CI is green on the head you pushed, and every
verification run the PR body cites has finished green:
`jobshop comment "PR <url> head <sha>"` (sha from `git rev-parse HEAD`),
then `jobshop close --outcome pr`. The outcome `pr` makes the `review-merge` step ready
for the refinery (a close without it skips review and merge). If the
remaining work can only happen later, do not
close failed: comment what you wait for and park on it. An event has a
wait kind: a deploy in progress is `jobshop wait add
'deploy:repo={{.repo}},sha=<sha>' --resume -r "deploy of <sha>"`, a
merge `github.pr:pr=...,state=merged`, a tag `github.tag:repo=...,tag=...`,
another run `run:id=<run>,state=done`. Only a real time (a 24h window,
a scheduled time) is `jobshop release --until <RFC3339 or duration>
--time-wait -r "<why>"`. Either defers the step without counting a failure
or reaching a human. If you cannot deliver for any other reason,
`jobshop close -r failed` with a
comment saying why; the run retries and then escalates to a human.

Report state as you go, so triage and the planner read it instead
of your comments: `jobshop report progress "<what is done>"`, `jobshop
report blocked "<on what>"`, `jobshop report failed "<why>"` before
`close -r failed`. When only a person can decide (a decision, a
credential or access, a host config problem you cannot fix), ask once with
`jobshop report waiting "<what>" --question "<the decision needed>"
--context "<what you did and found, why you cannot decide, what you do
if nobody answers>"`, with `--option "<Title>|<what it does>"` for each
concrete alternative and `--recommend N`; the answer arrives as a
message in this session. Never ask a person for status or to say done.
