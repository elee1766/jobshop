Step `review-merge` of a `pr-work` run for Linear issue {{.issue}}.

Find the PR in the comments of the run's `implement` step
(`jobshop run show <run>` with the run id from this ticket's
`parent_id`, then `jobshop ticket show <implement-step>`), then read:
- `gh pr view <n> --json headRefOid,mergeable,mergeStateStatus,reviewDecision,statusCheckRollup,comments,reviews`
- the issue and its acceptance criteria (`jobshop linear issue {{.issue}}`)
- the diff (`gh pr diff <n>`)
- master CI (`gh run list --branch master --limit 3`)

Run the repo's checks on the PR head yourself before deciding (the repo's
contributor docs name them, e.g. a linter): `gh pr checkout <n>` in your
worktree, then `{{.check_cmd}}`. A failure, a lint finding, or a new
`//nolint` or relaxed linter config without a reason in the PR body is
case 4: put the exact output in your findings. You do not fix code
yourself.

A problem you find that is outside this PR's issue does not block it:
file it with `jobshop linear create --team {{.linear_team}} --title
"<title>" --description - --key review-{{.issue}}-<short slug>`, then
`jobshop linear relate <NEW> --related {{.issue}}`, after `jobshop linear
search` finds no duplicate.

PR in another repo (JOB-33): if the PR is not in {{.repo}}, do not
review or merge it here. The implement step's comment names the
`pr-review` run of that repo's instance (`review run <id>`); with no run
named, start it yourself (`jobshop run start <prefix>pr-review --var
pr=<n> --key pr-<n>`, `<prefix>` from a run of that repo in `jobshop run
list`). Then park on it:
`jobshop wait add 'run:id=<review run>,state=closed' --resume --timeout 24h --reason "review of <url> by its own instance"`
and end the turn. On resume: merged (`gh pr view <url> --json state`),
comment the merge sha here and `jobshop close`; post-merge then checks
the issue. Closed unmerged: comment that and `jobshop close -r failed`.

Then do exactly one of these. A plain release of this step counts as a
failed attempt (like a failed close). To wait for CI on a PR head, set
`jobshop wait add 'github.checks:pr={{.repo}}#<n>,state=success' 'github.checks:pr={{.repo}}#<n>,state=failure' --any --resume --timeout 1h --reason "CI on PR #<n>"`
and end the turn: a wait is a deferral, not a failed attempt, and it
resumes here within a minute of CI finishing. If it resumes with
"conflicts with its base", rebase the PR head on `origin/master`, resolve,
`push --force-with-lease`, and set the wait again. For master CI use
`github.checks:repo={{.repo}},sha=<sha>` the same way. Do not poll with
`sleep` or defer with a fixed `release --until` for CI.

1. Master CI red: merge nothing. Make sure a fix is under way (a
   `ci-repair` run in `jobshop run list`, or an open dev ticket); file a
   priority-0 dev ticket if not. Wait for
   master to go green (poll `gh run list --branch master --limit 1`),
   then start over from the top.
2. Behind master (`mergeStateStatus` BEHIND): `gh pr update-branch <n>`,
   then wait for CI on the new head.
   Conflicts only in docs (`*.md`; `gh pr checkout <n>`,
   `git rebase origin/master`, every conflicted file is docs): resolve
   them here instead of failing the step. Keep both sides; if both added
   a numbered section, give this PR the next free number. Then
   `git push --force-with-lease`, comment on the PR what you resolved,
   and wait for CI on the new head.
   Non-docs conflicts with master and nothing else wrong (no review
   finding): comment the conflicted files on the PR and on the `implement`
   step, `jobshop run resume <run> --step implement`, then
   `jobshop close -r rebase` on this step. It reopens behind implement
   without counting a failed attempt: master moving is not the PR's fault.
   A conflict plus any case 4 finding is case 4.
3. Waiting (CI pending): wait for it.
4. Needs code changes (required checks failing after a rerun of a flaky job, 
   unresolved review comments, acceptance criteria not met):
   comment concrete findings on the PR and on the `implement` step, then
   `jobshop run resume <run> --step implement` (reopens implement for a
   dev session) and `jobshop close -r failed` on this step (no reason
   text: the findings are already in the comments). The failed close
   reopens this step blocked
   behind implement and counts one attempt; at the retry limit the run
   escalates to a human instead.
5. Waiting for approval (step 6 needs an approval that is missing or is
   for an older head): comment on this step which approval is missing,
   then `jobshop close -r failed`. The step reopens and the next attempt
   checks again; at the retry limit the run escalates to a human.
6. Ready (required checks green on the current head, required approval
   present, no unresolved review comments, criteria met):
   the title is the squash commit subject and must be a conventional
   commit, `<type>(<scope>): <summary> ({{.issue}})` with type feat,
   fix, perf, refactor, docs, test, ci, build, chore or revert and a
   lowercase summary; fix a title that is not with
   `gh pr edit <n> --title "..."` first. Then
   `gh pr merge <n> --{{.merge_method}} --match-head-commit <headRefOid>`.
   Never pass `--delete-branch` (gh then removes the dev session's
   worktree; GitHub deletes the remote branch on merge). Comment the merge
   commit sha on this step, then `jobshop close`. `post-merge` checks CI
   and the deploy.

Merge policy for this repo:
- Merge method: `{{.merge_method}}`.
- Required checks: `{{.required_checks}}`. Empty means every check in
  `statusCheckRollup` must pass. Otherwise each listed check name
  (comma-separated) must be SUCCESS on `headRefOid`; others may fail:
  `gh pr checks <n> --required=false --json name,state`.
- Required approver: `{{.required_approver}}`. Empty means no approval is
  needed. Otherwise that login must have an APPROVED review on the
  current head; an approval of an older commit does not count:
  `gh api --paginate repos/{{.repo}}/pulls/<n>/reviews --jq '[.[] | select(.user.login == "{{.required_approver}}" and .state == "APPROVED" and .commit_id == "<headRefOid>")] | length'`
  must print a number above 0.
