Step `review-merge` of a `pr-work` run for issue #${issue}.

Find the PR in the comments of the run's `implement` step
(`jobshop run show <run>` with the run id from this ticket's
`parent_id`, then `jobshop ticket show <implement-step>`), then read:
- `gh pr view <n> --json headRefOid,mergeable,mergeStateStatus,reviewDecision,statusCheckRollup,comments,reviews`
- the issue and its acceptance criteria (`gh issue view ${issue} --comments`)
- the diff (`gh pr diff <n>`)
- master CI (`gh run list --branch master --limit 3`)

Then do exactly one of these. Releasing this step counts as a failed
attempt (like a failed close), so wait inside your turn instead of
releasing: poll with `gh pr checks <n> --watch`, `gh run watch`, or
`sleep 300` between checks. Release only if a wait passes an hour.

1. Master CI red: merge nothing. Make sure an open fix issue exists
   (`gh issue list --search`); file it at priority 0 if not. Wait for
   master to go green (poll `gh run list --branch master --limit 1`),
   then start over from the top.
2. Behind master (`mergeStateStatus` BEHIND): `gh pr update-branch <n>`,
   then wait for CI on the new head.
   Conflicts only in docs (`*.md`; `gh pr checkout <n>`,
   `git rebase origin/master`, every conflicted file is docs): resolve
   them here instead of failing the step. Keep both sides; if both added
   a numbered section, give this PR the next free number. Then
   `git push --force-with-lease`, comment on the PR what you resolved,
   and wait for CI on the new head. Any non-docs conflict is case 4.
3. Waiting (CI pending): wait for it.
4. Needs code changes (required checks failing after a rerun of a flaky job, merge
   conflicts, unresolved review comments, acceptance criteria not met):
   comment concrete findings on the PR and on the `implement` step, then
   `jobshop run resume <run> --step implement` (reopens implement for a
   dev session) and `jobshop close -r failed` on this step with the
   findings as the reason. The failed close reopens this step blocked
   behind implement and counts one attempt; at the retry limit the run
   escalates to a human instead.
5. Waiting for approval (step 6 needs an approval that is missing or is
   for an older head): comment on this step which approval is missing,
   then `jobshop close -r failed`. The step reopens and the next attempt
   checks again; at the retry limit the run escalates to a human.
6. Ready (required checks green on the current head, required approval
   present, no unresolved review comments, criteria met):
   `gh pr merge <n> --${merge_method} --match-head-commit <headRefOid>`.
   Never pass `--delete-branch` (gh then removes the dev session's
   worktree; GitHub deletes the remote branch on merge). Comment the merge
   commit sha on this step, then `jobshop close`. `post-merge` checks CI
   and the deploy.

Merge policy for this repo:
- Merge method: `${merge_method}`.
- Required checks: `${required_checks}`. Empty means every check in
  `statusCheckRollup` must pass. Otherwise each listed check name
  (comma-separated) must be SUCCESS on `headRefOid`; others may fail:
  `gh pr checks <n> --required=false --json name,state`.
- Required approver: `${required_approver}`. Empty means no approval is
  needed. Otherwise that login must have an APPROVED review on the
  current head; an approval of an older commit does not count:
  `gh api --paginate repos/${repo}/pulls/<n>/reviews --jq '[.[] | select(.user.login == "${required_approver}" and .state == "APPROVED" and .commit_id == "<headRefOid>")] | length'`
  must print a number above 0.
