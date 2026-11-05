# Release PRs go to the refinery through a trigger (#698)

2026-11-05.

release-please PRs have no pr-work run, so nothing merged them until the
`pr_unmerged` stall fired after 2h and a stall-check filed a refinery
ticket by hand (acphost #93 and #103, jobshop #657).

packs/github now has:

- var `release_pr_label`, default `autorelease: pending`. Empty turns
  the trigger off.
- trigger `release-pr`: a `gh pr list` check prints the lowest open,
  non-draft PR with that label whose checks are all green. It starts a
  `release-merge` run with `pr=<n>` and key `release-pr-<n>`. The
  trigger key (`release-pr/<n>`) dedups across ticks and after the run
  closes, so one PR gets one run. A trigger holds one open run, so a
  second release PR waits for the first.
- workflow `release-merge`: one `merge` step in `refinery`
  (`prompts/release-merge.md`), `meta.github.pr` set. No review, by
  policy (#569, #644/#646).
- `pr_unmerged` skips PRs with the label.

Test: `TestReleasePRTrigger` in pkg/control loads the real pack with a
fake `gh` (needs `jq` on PATH; it skips without).
