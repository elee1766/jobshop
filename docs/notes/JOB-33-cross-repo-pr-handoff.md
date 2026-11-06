# JOB-33: cross-repo PR handoff (2026-11-06)

A `pr-work` run of one packs/github instance can open a PR in another
instance's repo (acphost issue-125 opened elee1766/jobshop#721). Its own
review-merge must not merge it, and before this nothing in the other
instance adopted an existing PR.

As built:

- `pr-review` workflow (vars `[pr]`, `pr: pr`, key `pr-<n>`): steps
  `review-merge` (refinery, `prompts/pr-review/merge.md`) and
  `post-merge` (deploy-check, only after `review-merge` closes with
  outcome `merged`). The prompts read the PR and its linked GitHub or
  Linear issue instead of an implement step. Outcomes `held`, `closed`
  and `owned` end the run without post-merge.
- implement (both trackers): a PR outside `{{.repo}}` starts
  `<prefix>pr-review --var pr=<n> --key pr-<n>` of that repo's instance,
  names the run in its `PR <url> head <sha> review run <id>` comment,
  and closes `pr`. The origin review-merge parks on
  `run:id=<review run>,state=closed` instead of closing done.
- `pr_unmerged` moved from the `stall` trigger to its own `pr-unmerged`
  trigger that starts `pr-review` instead of `stall-check`, same check
  and skips (human, `stall_pr_skip_authors`, `stall_pr_skip_labels`,
  `release_pr_label`).
- Stall triggers may set `run.key` from the stall vars (it was refused).
  The run key is that, while the trigger's own dedup stays
  `<trigger>/stall:<signal>:<subject>`. A fire that returns an open run
  (started by hand under the same key) records no stall learning.

Test: `TestPRReview` in pkg/control.
