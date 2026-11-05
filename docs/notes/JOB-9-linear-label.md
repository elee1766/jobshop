# jobshop linear label (JOB-9)

2026-11-05. A stall-check on JOB-7 looked for a Linear label command,
did not find one, and left the issue in Todo, so `linear_issue_no_run`
fired again every cooldown. `jobshop linear update --add-label` existed
under a name nobody looked for.

- `jobshop linear label <KEY> add|remove <label>...` runs `update
  --add-label/--remove-label`: same guard, `--key` idempotency and read
  back. An unknown label fails with the team's label names.
- The stall, Linear planner, GitHub planner (Linear intake) and Linear
  implement prompts name it for marking an issue that should not get a
  run.
- `linear_skip_labels` adds `tracking` to its default (now
  `Blocked,target,proposed,tracking`), so a tracking issue such as JOB-3
  never starts a run once labeled.
