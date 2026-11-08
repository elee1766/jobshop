# Scan rename and Starlark-only packs

2026-11-08. Prerequisite of JOB-123.

- "Stall" is renamed with no compatibility layer: `when.scan`,
  `ScanSignal`, a detected item is a finding, run keys
  `scan:<signal>:<subject>`, `meta.trigger.finding`, `event.source
  "scan"`, `trigger list` field `findings`, signal `step_idle` (was
  `step_stalled`). The github pack's workflow, pool and prompt are
  `triage`; vars `triage_model`, `triage_effort`, `scan_after`,
  `scan_step_after`, `scan_human_after`, `pr_review_skip_authors`,
  `pr_review_skip_labels`. Runs keyed `stall:` no longer hold a subject.
- `packs/github`, `packs/selfdev` and `packs/operator` are CUE only. The
  YAML packs, `starlark_workflows` / `starlark_triggers` and the parity
  test are gone. Triggers are handlers and scan scripts in
  `triggers.star` that read GitHub through the `github` built-in; no pack
  shells out to `gh`. The operator push guard is `guard.star:push`.
- `packs/agent` (jobcity) is still YAML with shell checks.
- `packs/packs_test.go` loads the three packs together against a fake
  GitHub API and checks each trigger's runs and the push guard.
