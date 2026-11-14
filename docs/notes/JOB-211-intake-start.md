# Intake starts the issue's run itself (JOB-211)

2026-11-14. A `linear_issue_no_run` finding used to open a `triage` run
whose agent ran `jobshop linear eligible` and then `jobshop run start`
(61 such runs in three days, all to start work the controller could
start). The signal entry now takes `start: <workflow>`; with it the
controller starts that workflow for each finding and no triage run opens.

- `start` names a workflow with `tracker: linear` and an issue var (any
  other var must be a pack var). Config load rejects an unknown or
  non-Linear workflow, a disabled one, and `start` on a handler trigger.
  A pack prefix renames it with the workflows.
- The run is started as `jobshop run start` would: key `linear-<KEY>`,
  var = the issue key, so the planner's start and the scan's return each
  other's open run. The run also carries `meta.trigger` for the finding,
  so `max_runs`, `cooldown` and the escalation hold work as for triage.
  Intake hooks, the run title read and `meta.linear` are as for any
  Linear run.
- The gate is unchanged and runs in the signal's read: open issue in
  `states`, one of `labels`, none of `skip_labels`, no open Linear
  blocker or open `Depends on`, no open run, past `grace`. The read is
  cached for `poll` (5m; an Issue webhook drops the cache), so a label
  added in that window is seen at the next read. The fire-time store
  recheck for an open run stays.
- No learning is recorded for these runs: they are intake, not stuck work.
- packs/github sets `start` to `pr-work-linear` (github tracker) or
  `pr-work` (linear tracker). With `start` empty (a hand-written config)
  the trigger's `run` still takes the finding, and the triage prompt
  keeps its `linear_issue_no_run` section for that.

Tests: `TestLinearScanStart` (pkg/control: one run from the scan, skip
label and planner-started issues get none, a second pass starts nothing,
same key returns the same run), `TestLinearIntakeStart` (pkg/config),
`TestGithubPackLinearIntakeStart` (packs).
