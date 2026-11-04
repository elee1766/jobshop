# Step needs and when (#385)

2026-11-04.

A workflow step may set:

- `needs: [a, b]`: the step is blocked by those steps instead of the one
  before it. Steps with the same needs become ready together. The needs
  must be earlier steps, so the graph has no cycles. Without `needs:` a
  step needs the previous step, as before.
- `when: {step: review, outcome: approve}`: the step runs only if
  `review` closed done with that outcome. `when.step` must be an
  ancestor through needs, so it is finished before the gate is checked.

An agent sets an outcome with `jobshop close <step> --outcome approve`.
The outcome must match `^[a-z][a-z0-9_-]{0,31}$`. It is stored in the
step's fields (`outcome`) and shown by `run show`.

Gated steps are created deferred to year 9999, so they never show as
ready. The run check (`gateSteps`, under the run lock) looks at each
gated open step whose needs are all finished. It clears the defer when
the outcome matches. Otherwise it closes the step with resolution
`skipped` and a comment. A skipped step counts as finished:
- for blocker edges (store queries),
- for the current step,
- for run completion.
So steps after it still run, and the run can close done.
