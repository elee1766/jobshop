# pr_unmerged skips operator and hold PRs (#509)

2026-11-05. The github pack's `pr_unmerged` stall source no longer flags
PRs authored by `human` or a login in `stall_pr_skip_authors`, or labeled
with one of `stall_pr_skip_labels` (default `hold`). Both are
comma-separated and matched case-insensitively and exactly. Operator PRs
kept open on purpose (#421, #392) had opened stall-check runs that went
`needs_human` and sat in attention. `pkg/config` runs the rendered jq
filter on a fixture (when jq is on PATH): a hold-labeled PR and an
operator PR are skipped, other green PRs are still flagged.
