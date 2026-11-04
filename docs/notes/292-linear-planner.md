# Linear planner (#292)

2026-11-04. Part of #258. With write commands (#291) and intake (#290)
the `tracker: linear` prompts no longer depend on a human filing issues.

- Planner (`prompts/linear/plan.md`): lists the team with `jobshop
  linear list --team` per state in `linear_intake_states`, with
  `linear_intake_labels` and the new `linear_intake_assignee` (empty is
  any), plus the issues the target names. Each goes through `linear
  eligible`; eligible ones start `pr-work` with key `linear-<KEY>`, the
  key the intake stall-check uses, so the two cannot double-start.
- Work with no issue: `linear create --team ... --key plan-<slug>`, then
  `linear relate <NEW> --related <target>`. Decisions go through
  `linear ask`.
- Dev and review-merge: out-of-scope findings become `linear create`
  issues in the team (keys `found-<KEY>-<slug>`, `review-<KEY>-<slug>`),
  linked with `linear relate --related` to the source issue, after a
  `linear search` for duplicates. Dev tickets are no longer used for them.
- Task tickets in dev also start runs with `--key linear-<KEY>`.

Tests: `TestGithubPackLinearTracker` (vars and prompt text),
`TestLinearPlannerIntakeKey` (fake Linear server: planner and intake
starts give one run and the stall clears). Live check:
`docs/verify/linear-pack.md`.
