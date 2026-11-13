# #1246: callback failures after the last step

2026-11-13

Run js-0034pvnqaw002 (pr-work-linear) went needs_human the minute its
review-merge step closed done, and its escalation said "a step failed at
its retry limit". The run had started on 2026-11-09, so its recorded
inputs (pack vars plus run vars, taken at start) had no `pool_prefix`.
#1239 then added that var and used it in `prompts/linear/post-merge.md`.
`pr_work_linear_completed` returned the post-merge `start_step`, and
rendering its `prompt_file` from the old inputs failed with
`var "pool_prefix" is not set`. `run resume` retried the same
invocation, which failed the same way.

As built:

- A callback step's `prompt_file` renders from the workflow's live pack
  vars, then the run's recorded inputs, then the action's inputs. A pack
  var added after a run started is available; what the run accepted
  still wins.
- A failed workflow callback sets `fields.hook_error` on the run and
  starts a new escalation round. The escalation body, the escalation
  chat's reason and the controller learning say "its workflow callback
  failed: <error>". Moving the run out of needs_human clears the field.
- `run resume` on a needs_human callback run retries the failed
  invocation as before. With none to retry and every step finished, it
  forgets the last step's completion and sets the run running, so
  `on_step_completed` runs again for that step.

Tests: `pkg/orchestration/control/hookescalate_test.go`.
