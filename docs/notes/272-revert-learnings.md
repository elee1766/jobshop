# Revert learnings (#272)

2026-11-04. Follow-up to #264 (target #247, item Capture).

- Detector: the activity feed poll (#255). For each newly seen GitHub
  `merge` item the controller reads the PR. Chosen over subscriptions
  because the feed already polls every worked repo with ETags, sees
  merges by anyone (subscriptions only follow a ticket's own threads),
  and an idle repo costs nothing. Needs `activity:` configured.
- A revert: title starts `Revert "`, or the body has `This reverts commit
  <sha>` or GitHub's `Reverts owner/name#n`. The original is confirmed
  merged: for a sha, the commit's PRs (`commits/{sha}/pulls`) must
  include one whose merge commit starts with that sha; for `#n`, PR n
  must be merged.
- Only PRs jobshop merged count: the original must be linked (meta or a
  `PR <url>` comment) from a `pr-work` run or its steps in that repo.
- Records one `correction` on the run, source `controller`, scope the
  run's repo. Summary: `revert <repo>#<revert>: PR #<orig> for issue #<n>
  (run <id>) was reverted by PR #<revert> <url>`; `fix` is the revert
  body. Deduped on the `revert <repo>#<n>` prefix among the run's
  corrections, so a lost cache or restart does not record it twice.
