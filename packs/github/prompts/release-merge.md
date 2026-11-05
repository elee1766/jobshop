Step `merge` of a `release-merge` run: release PR #{{.pr}} in {{.repo}}.

The `release-pr` trigger started this run because the PR carries the
`{{.release_pr_label}}` label and every check on it is green. Release PRs
are merged by policy without review: do not review the changelog.

Read `gh pr view {{.pr}} --repo {{.repo}} --json state,headRefOid,mergeable,mergeStateStatus,statusCheckRollup,labels`.

1. Already merged or closed: comment that on this step and `jobshop close`.
2. Behind master or checks not green on the current head (release-please
   rewrites the PR on every merge): set
   `jobshop wait add 'github.checks:pr={{.repo}}#{{.pr}},state=success' 'github.checks:pr={{.repo}}#{{.pr}},state=failure' --any --resume --timeout 1h --reason "CI on release PR #{{.pr}}"`
   and end the turn. Red checks: comment the failing check on this step
   and `jobshop close -r failed`; at the retry limit the run escalates.
3. Green: `gh pr merge {{.pr}} --repo {{.repo}} --{{.merge_method}} --match-head-commit <headRefOid>`
   (never `--delete-branch`). Comment the merge sha on this step, then
   `jobshop close`. The release workflow tags the release from master.
