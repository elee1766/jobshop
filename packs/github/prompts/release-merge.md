Step `merge` of a `release-merge` run: release PR #{{.pr}} in {{.repo}}.

The `release-pr` trigger started this run because the PR carries the
`{{.release_pr_label}}` label and every check on it is green. Release PRs
are merged by policy without review: do not review the changelog.

Read `gh pr view {{.pr}} --repo {{.repo}} --json state,headRefOid,mergeable,mergeStateStatus,statusCheckRollup,labels`.

1. Already merged or closed: comment that on this step and `jobshop close`.
2. A release workflow run on master not finished yet
   (`gh run list --repo {{.repo}} --workflow release --branch master --json status --jq '[.[] | select(.status != "completed")] | length'`
   is not 0): release-please may still rewrite this PR, or tag and
   supersede it. Merging now raced twice into a second PR for an
   already tagged version (v0.2.4, #761 and #763). Set
   `jobshop wait add 'time:at=2m' --resume --reason "release workflow running on master"`
   and end the turn.
3. The version in the PR title is already tagged
   (`gh release view v<version> --repo {{.repo}}` succeeds): the PR
   duplicates a published release. Do not merge. Comment that on this
   step and `jobshop close -r failed`.
4. Behind master or checks not green on the current head (release-please
   rewrites the PR on every merge): set
   `jobshop wait add 'github.checks:pr={{.repo}}#{{.pr}},state=success,workflow={{.ci_workflow}}' 'github.checks:pr={{.repo}}#{{.pr}},state=failure,workflow={{.ci_workflow}}' --any --resume --timeout 1h --reason "CI on release PR #{{.pr}}"`
   and end the turn. Without `workflow=` the wait can end green on
   `pr-title` alone while `{{.ci_workflow}}` is still queued. On resume,
   read `statusCheckRollup` and `mergeStateStatus` again before merging.
   Red checks: comment the failing check on this step and
   `jobshop close -r failed`; at the retry limit the run escalates.
5. Green: `gh pr merge {{.pr}} --repo {{.repo}} --{{.merge_method}} --match-head-commit <headRefOid>`
   (never `--delete-branch`). Comment the merge sha on this step, then
   `jobshop close`. The release workflow tags the release from master.
