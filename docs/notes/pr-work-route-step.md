# pr-work routes in the workflow

2026-11-10. Replaces the planner-side routing of JOB-59
(`JOB-59-dev-pool-routing.md`).

- `pr-work` and `pr-work-linear` start with a `route` step in the
  `triage` pool (`prompts/<tracker>/route.md`, 1 retry). It reads the
  issue and closes with outcome `scoped` or `primary`.
- `on_step_completed` then starts the one `implement` step in
  `dev-sonnet` (`scoped`) or `dev` (`primary`). No step is created for
  the route not taken. `retry_pool` is still `dev`.
- Any other outcome fails the callback, so the run goes to `needs_human`
  with only the route step.
- Removed: the `dev_pool` pack var and run var, and the `dev-pool` prompt
  rule that had the planner pick a pool and pass `--var dev_pool=`. A
  start that still passes `--var dev_pool` is refused as an unknown var.
- Batch runs are routed too. Before, `batch start` took no vars and its
  runs stayed on `dev`.
- The rubric and the `model:sonnet` / `model:opus` label override moved
  into the route prompt unchanged.

Deploy: this edits `packs/github/workflows.star` in place, so a callback
run open across the deploy fails with `hook manifest changed`. Cancel
and restart it, or `jobshop run resume <run> --repin` on the admin
socket. A repinned `pr-work` run that already has `implement` continues
from there: the new callbacks accept the old step ids.

Test: `packs/packs_test.go` `TestPRWorkRoutes`.
