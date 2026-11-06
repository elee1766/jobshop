Step `review-merge` of a `pr-review` run for PR #{{.pr}} in {{.repo}}.

This PR has no `pr-work` run of this instance behind it (JOB-33): another
instance's `implement` step opened it in this repo, or `pr-unmerged`
found it green and unowned, or someone started this run by hand. Review
and merge it like a `pr-work` review, reading the PR instead of an
implement step.

Read:
- `gh pr view {{.pr}} --repo {{.repo}} --json title,body,author,state,headRefOid,mergeable,mergeStateStatus,reviewDecision,statusCheckRollup,labels,comments,reviews,closingIssuesReferences`
- the linked issue, if any, and its acceptance criteria: a GitHub issue
  in `closingIssuesReferences` or a `Fixes #<n>` / `Fixes owner/name#<n>`
  line (`gh issue view <n> --repo <repo> --comments`), or a Linear key
  (`ABC-12`) in the title or body (`jobshop linear issue <KEY>`). No
  linked issue: the PR body is the spec.
- the diff (`gh pr diff {{.pr}} --repo {{.repo}}`)
- master CI (`gh run list --repo {{.repo}} --branch master --limit 3`)
- the run that opened it, if one did: `jobshop run list` and a
  `pr-work` run whose implement step names this PR (its comments say
  `PR <url> head <sha>`). Its review-merge step waits on this run.

First, and on every resume:

1. Already merged: comment the merge sha on this step and `jobshop close
   --outcome merged` (post-merge checks CI and the deploy). Closed
   unmerged: comment that and `jobshop close --outcome closed`; the run
   ends without post-merge.
2. Held on purpose (draft, author {{.human}} or in
   `{{.stall_pr_skip_authors}}`, a label in `{{.stall_pr_skip_labels}}`,
   or the `{{.release_pr_label}}` label, which `release-merge` owns):
   comment why and `jobshop close --outcome held`. Nothing is merged.
3. Owned by an open `pr-work` run of this instance (its implement step
   names this PR in this repo): that run's review-merge merges it, not
   this one. Comment on that review-merge step that the PR is green and
   waiting, then `jobshop close --outcome owned`. A stuck step there is
   the stall trigger's (`step_stalled`).

Run the repo's checks on the PR head yourself before deciding: `git fetch
origin pull/{{.pr}}/head && git checkout --detach FETCH_HEAD`, then, as a
separate command, confirm `git rev-parse HEAD` equals `headRefOid`. Then
run `{{.check_cmd}}`. A failure, a lint finding, or a new `//nolint` or
relaxed linter config without a reason in the PR body is a finding (case
4 below). You do not fix code yourself.

A PR that removes or renames a pack var breaks every host config that
still sets it (#261). It must keep the old name with `renamed_vars: {old:
new}` in pack.yaml, or list in the PR body the hosts and the config
change each needs. Doing neither is a finding.

Before any merge, nothing on the linked issue may wait on a human (an
open `needs-human` message for it in `jobshop ticket list --label
needs-human --status open --kind message`), and the PR and issue must
not carry `needs-decision`. Otherwise it is case 5.

Then do exactly one of these. A plain release counts as a failed
attempt. Wait for CI on the PR head with
`jobshop wait add 'github.checks:pr={{.repo}}#{{.pr}},state=success' 'github.checks:pr={{.repo}}#{{.pr}},state=failure' --any --resume --timeout 1h --reason "CI on PR #{{.pr}}"`
and end the turn. Do not poll with `sleep` or defer with a fixed
`release --until` for CI.

1. Master CI red: merge nothing. Make sure a fix is under way (an open
   fix issue or a `ci-repair` run); wait for master CI with
   `github.checks:repo={{.repo}},sha=<sha>`, then start over.
2. Behind master: `gh pr update-branch {{.pr}} --repo {{.repo}}`, then
   wait for CI on the new head. Right before merging, `git fetch origin`
   and `git merge-tree --write-tree origin/master <head>`; a conflict is
   case 4 (docs-only conflicts: resolve them here, keep both sides,
   `git push --force-with-lease`, and wait for CI).
3. CI pending: wait for it.
4. Needs code changes: comment concrete findings on the PR. If a run
   opened it, comment them on that run's implement step too. File one
   dev ticket for the fix: `jobshop ticket create "Address review of PR
   #{{.pr}}" --pool dev -b "<PR url, findings, linked issue>"`. Then
   park on a new head:
   `jobshop wait add 'github.pr:pr={{.repo}}#{{.pr}},state=head_changed' 'github.pr:pr={{.repo}}#{{.pr}},state=closed' --any --resume --timeout 24h --reason "fixes on PR #{{.pr}}"`
   and end the turn; on resume start over from the top.
5. Waiting for approval or a human: comment which, then `jobshop close
   -r failed`. At the retry limit the run escalates.
6. Ready (required checks green on the current head, approval present,
   no open question, no unresolved review comments, criteria met): the
   title must be a conventional commit with a lowercase summary that
   names the linked issue (`(#<n>)` or `(<KEY>)`) when there is one; fix
   it with `gh pr edit {{.pr}} --repo {{.repo}} --title "..."`. Then
   `gh pr merge {{.pr}} --repo {{.repo}} --{{.merge_method}} --match-head-commit <headRefOid>`
   (never `--delete-branch`). Comment the merge commit sha on this step,
   then `jobshop close --outcome merged`. Only that outcome runs
   `post-merge`.

Merge policy for this repo:
- Merge method: `{{.merge_method}}`.
- Required checks: `{{.required_checks}}`. Empty means every check in
  `statusCheckRollup` must pass; otherwise each listed check must be
  SUCCESS on `headRefOid`.
- Required approver: `{{.required_approver}}`. Empty means none.
  Otherwise that login needs an APPROVED review on the current head:
  `gh api --paginate repos/{{.repo}}/pulls/{{.pr}}/reviews --jq '[.[] | select(.user.login == "{{.required_approver}}" and .state == "APPROVED" and .commit_id == "<headRefOid>")] | length'`
  must print a number above 0.
