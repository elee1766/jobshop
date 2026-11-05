# Live check (#635)

2026-11-05.

- Runbook steps are fenced blocks. `sh readonly` (the word `readonly` in
  the info string) marks a block safe to run unattended. Format:
  `docs/verify/README.md`. Marked so far: deploy-state, evidence,
  respawn-env (second block), telemetry.
- `jobshop runbook run <file> [--json] [--timeout 2m]` (pkg/runbook) runs
  the marked blocks with `sh -e` in order, records output (8KiB tail) and
  pass/fail per block, lists unmarked blocks as left for the operator,
  and exits non-zero if any marked block failed. It does not talk to the
  controller.
- `packs/selfdev` workflow `live-check` (vars `runbook`, `issue`): one
  command step on the dev pool with `worktree: true`, so it runs under
  the pool sandbox in a worktree of master. It prints the report (the
  step output on the run) and comments it on the issue; a failing block
  fails the step, and the run escalates to plan.
- Planner: prompt rule `selfdev-live-check` tells the plan pool to start
  `live-check` for target items waiting only on a runbook.
- Tests: `pkg/runbook` (`TestRunOnlyReadOnly` on a fixture with marked and
  unmarked blocks, `TestVerifyDocsParse` over docs/verify).
