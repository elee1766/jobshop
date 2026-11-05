# Linear targets (JOB-4)

2026-11-05.

- Targets move from GitHub `target` issues to Linear parent issues with
  sub-issues (operator decision, 2026-11-05). This applies to the
  packs/github planner (`prompts/github/plan.md`) when `linear_team` is
  set.
- The planner reads targets with `jobshop linear list --team <team>
  --label target`, and reads each target's sub-issues with `jobshop
  linear context`. An item is done when its sub-issue is Done or
  Canceled. When every item is done, the planner comments a summary on
  the target and moves it to Done.
- Missing work becomes a new sub-issue: `jobshop linear create --parent
  <TARGET> --state Todo` (with the repo label when
  `linear_intake_labels` is set), so intake starts it. Scope changes go
  in a sub-issue labeled `needs-decision`. Progress comments go on the
  Linear target.
- Proposals are a Linear issue labeled `proposed`, with sub-issues in
  Backlog, and `jobshop linear ask` to request approval. A human
  approves by relabeling it `target` and moving the sub-issues to Todo.
- Intake skips parents: the `linear_skip_labels` default is now
  `Blocked,target,proposed`. An instance that overrides it must keep
  those two labels.
- GitHub `target` issues are still read while any is open, so there is
  no cutover. The operator closes them once their items have moved.
- `prompts/linear/implement.md` makes the PR body's
  `Fixes [KEY](https://linear.app/<workspace>/issue/KEY)` a link.

Tests: `TestLinearIntakeSkipsTargets` (pkg/config, fixture pack) and
`TestLinearIntakeMatch` (pkg/control).
