Step `review-merge` of a `pr-work` run for issue #{{.issue}}.

Find the PR in the comments of the run's `implement` step
(`jobshop run show <run>` with the run id from this ticket's
`parent_id`, then `jobshop ticket show <implement-step>`), then read:
- `gh pr view <n> --json headRefOid,mergeable,mergeStateStatus,reviewDecision,statusCheckRollup,comments,reviews`
- the issue and its acceptance criteria (`gh issue view {{.issue}} --comments`)
- the diff (`gh pr diff <n>`)
- master CI (`gh run list --branch master --limit 3`)

Run the repo's checks on the PR head yourself before deciding (the repo's
contributor docs name them, e.g. a linter) on the PR head in your
worktree. `gh pr checkout` fails or stays on an old head in a detached
worktree: run `git fetch origin pull/<n>/head && git checkout --detach
FETCH_HEAD`, then, as a separate command, confirm `git rev-parse HEAD`
equals the PR's `headRefOid`. Then run `{{.check_cmd}}`. A failure, a lint finding, or a new
`//nolint` or relaxed linter config without a reason in the PR body is
case 4: put the exact output in your findings. You do not fix code
yourself.

A PR that removes or renames a pack var (a var reference a pack.yaml or its
prompts stop reading, or a `defaults:` key dropped) breaks every host
config that still sets it (#261). It must either keep the old name with
`renamed_vars: {old: new}` in pack.yaml, or list in the PR body the hosts
that set the var and the config change each needs. Doing neither is case 4.

Then do exactly one of these. A plain release of this step counts as a
failed attempt (like a failed close). To wait for CI on a PR head, set
`jobshop wait add 'github.checks:pr={{.repo}}#<n>,state=success' 'github.checks:pr={{.repo}}#<n>,state=failure' --any --resume --timeout 1h --reason "CI on PR #<n>"`
and end the turn: a wait is a deferral, not a failed attempt, and it
resumes here within a minute of CI finishing. If it resumes with
"conflicts with its base", rebase the PR head on `origin/master`, resolve,
`push --force-with-lease`, and set the wait again. For master CI use
`github.checks:repo={{.repo}},sha=<sha>` the same way. Do not poll with
`sleep` or defer with a fixed `release --until` for CI.

1. Master CI red: merge nothing. Make sure an open fix issue exists
   (`gh issue list --search`); file it at priority 0 if not. Wait for
   master to go green (poll `gh run list --branch master --limit 1`),
   then start over from the top.
2. Behind master (`mergeStateStatus` BEHIND): `gh pr update-branch <n>`,
   then wait for CI on the new head.
   Conflicts only in docs (`*.md`; check out the PR head as above,
   `git rebase origin/master`, every conflicted file is docs): resolve
   them here instead of failing the step. Keep both sides; if both added
   a numbered section, give this PR the next free number. Then
   `git push --force-with-lease`, comment on the PR what you resolved,
   and wait for CI on the new head.
   `mergeStateStatus` can read CLEAN or UNKNOWN while stale, and a PR
   with no checks usually conflicts: right before merging, `git fetch
   origin` and `git merge-tree --write-tree origin/master <head>`; on a
   conflict treat it as BEHIND.
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
