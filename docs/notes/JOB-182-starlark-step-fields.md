# Starlark start_step fields (JOB-182)

2026-11-09. Closes the "not yet on start_step" list in
`JOB-124-script-steps.md`.

`start_step` takes five more keyword arguments. Each is checked with the
rule its workflow YAML counterpart has, and written to the step the same
way `buildRun` writes it, so the rest of the controller (resource
admission, placement, request_changes, `not_yet`) needs no new path.

| Argument | Steps | Rule |
| --- | --- | --- |
| `needs_resources=["emu", "pr:acme#1"]` | agent, script | List of nonempty strings. Each key must name a configured resource, keyed as `name:<key>` and counted bare (`CheckResourceKey`). Duplicates dropped. The step is ready only when every key has a free slot. |
| `placement={"required": [...], "prefer_for": "10m", "options": [{"pool": ..., "resources": [...], "weight": 100}]}` | agent | Replaces `pool`, `needs_resources` and `retry_pool` (setting them too is an error). At least one option, each with a pool in the run's project and the callback's reference scope. `prefer_for` is a Go duration >= 0. Unknown keys are rejected. The step starts on the highest weight, ties in call order. |
| `retry_after="30s"` | script | Positive Go duration. Used when the function returns `result(not_yet=True)`: the step is deferred by it without counting a failure. Default 1m. |
| `max_review_rounds=2` | agent | Integer >= 0. Omitted keeps the default (`DefaultReviewRounds`). 0 sends the first `request_changes` to a human. |
| `on_request_changes="impl"` | agent | Must be in `needs`. Without it a step that needs the previous step reopens that step, as before. |

Differences from YAML:

- A placement option has no `when`: the script decides in Starlark
  before building the dict.
- `pool` is optional only when `placement` is given.
- `result(not_yet=True)` is new. A command step says "not yet" with exit
  75 and an http step with a 503, a script step had no way to. It cannot
  be combined with `outcome`. `output` and `output_schema` are not read.
- `retry_after`, `max_review_rounds` and `on_request_changes` are for the
  step kind where YAML accepts them. Passing one to the other kind is an
  error, not ignored.

Placement pools are rechecked against the live configuration when a
planned batch is applied (`checkPlannedTargets`), like `pool` and
`retry_pool`, so a reload between planning and applying cannot route a
step to a pool outside the scope the run accepted. Resource names are
checked when the batch is planned, as a workflow start checks them.

Tests: `pkg/execution/automation/engine_test.go` (each field and every
rejection), `pkg/orchestration/control/hookfields_test.go` (a step with
`needs_resources` waits on the cap, placement falls back after
`prefer_for`, review fields reach the step, `not_yet` defers by
`retry_after`).
