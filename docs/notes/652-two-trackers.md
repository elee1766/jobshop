# Two trackers in one instance (#652)

2026-11-05. A `packs/github` instance with `tracker: github` takes work
from GitHub issues and, with `linear_team` and `linear_workspace` set,
from that Linear team too. PRs, CI and review stay on GitHub.

- `pr-work-linear`: same steps and pools as `pr-work` (dev, refinery,
  deploy-check), `tracker: linear`, the `prompts/linear/` step prompts.
  On a `tracker: linear` instance `pr-work` is unchanged.
- `pr-work` sets `linear_workspace` only when `tracker` is linear, so a
  github instance can set the var without failing `linear_workspace needs
  tracker: linear`.
- Config load no longer requires `linear_workspace` on a linear workflow
  (a github instance with empty Linear vars loads); starting a run of
  such a workflow fails with a message naming the var.
- `linear_issue_no_run` is off while `linear_team` is empty, as before.
  `prompts/stall.md` starts `pr-work-linear` for it on a github instance,
  `pr-work` on a linear one. The github planner lists the team's intake
  issues (same `linear list` filters and `eligible` gate as the linear
  planner) and starts `pr-work-linear` with key `linear-<KEY>`; targets,
  proposals and the decision log stay on GitHub.
- `packs/selfdev` passes `linear_workspace`, `linear_team`,
  `linear_intake_labels` and `linear_human` through (empty defaults).

Tests: `TestSelfdevTwoTrackers` (config, both ways),
`TestTwoTrackersLinearIntake` (fake Linear: intake on a github instance
starts `pr-work-linear` keyed `linear-JOB-1`), `TestPacksGolden`.
