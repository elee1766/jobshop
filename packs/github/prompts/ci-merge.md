Step `merge` of a `ci-repair` run: master CI failed at ${sha}.

Find the fix PR and issue in the comments of the run's `fix` step
(`jobshop run show <run>` with the run id from this ticket's
`parent_id`, then `jobshop ticket show <fix-step>`). If the fix step
says master was already green, comment that and `jobshop close`.

Read `gh pr view <n> --json headRefOid,mergeable,mergeStateStatus,statusCheckRollup,comments,reviews`
and `gh pr diff <n>`. Master is red, so this is the one PR allowed to
merge while it is: check that it fixes the failure in
`gh run view <id> --log-failed` and nothing else.

Releasing this step counts as a failed attempt, so wait inside your turn
(`gh pr checks <n> --watch`, `sleep 300`) instead of releasing.

1. Behind master: `gh pr update-branch <n>`, then wait for CI.
   Conflicts only in docs (`*.md`): rebase on `origin/master`, resolve
   them here keeping both sides, `git push --force-with-lease`, comment
   on the PR, and wait for CI. Code conflicts are case 2.
2. Needs changes (CI failing, conflicts, does not fix the failure):
   comment the findings on the PR and on the `fix` step,
   `jobshop run resume <run> --step fix`, then `jobshop close -r failed`.
   At the retry limit the run escalates to a human.
3. Waiting for approval: required approver `${required_approver}` is
   set and has no APPROVED review on the current head
   (`gh api --paginate repos/${repo}/pulls/<n>/reviews --jq '[.[] | select(.user.login == "${required_approver}" and .state == "APPROVED" and .commit_id == "<headRefOid>")] | length'`
   prints 0): comment that on this step and `jobshop close -r failed`;
   the next attempt checks again. An empty approver needs no approval.
4. Ready (required checks `${required_checks}` green on the current head,
   or every check when that is empty; approval present if required):
   `gh pr merge <n> --${merge_method} --match-head-commit <headRefOid>` (never
   `--delete-branch`). Wait for master CI on the merge commit
   (`gh run list --branch master --commit <merge sha>`, `gh run watch`).
   Green: comment the merge sha on this step and on the issue, then
   `jobshop close`. Red with a different failure: comment it and
   `jobshop close`; `master-red` starts a new run for the new SHA.
