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
3. Waiting (CI pending): wait for it.
4. Needs code changes (CI failing after a rerun of a flaky job, merge
   conflicts, unresolved review comments, acceptance criteria not met):
   comment concrete findings on the PR and on the `implement` step, then
   `jobshop run resume <run> --step implement` (reopens implement for a
   dev session) and `jobshop close -r failed` on this step with the
   findings as the reason. The failed close reopens this step blocked
   behind implement and counts one attempt; at the retry limit the run
   escalates to a human instead.
5. Ready (CI green on the current head, no unresolved review comments,
   criteria met): `gh pr merge <n> --squash --match-head-commit <headRefOid>`.
   Never pass `--delete-branch` (gh then removes the dev session's
   worktree; GitHub deletes the remote branch on merge). Comment the merge
   commit sha on this step, then `jobshop close`. `post-merge` checks CI
   and the deploy.
