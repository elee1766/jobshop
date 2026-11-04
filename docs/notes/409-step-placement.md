# Step placement: weighted pool options (#409)

2026-11-04. Follow-up to #400 (resource slots) and #384 (pool vars).

```yaml
steps:
  - id: review
    prompt_file: prompts/review.md
    title: "Review #${issue}"
    placement:
      required: ["pr:${repo}#${issue}"]   # every option needs these
      prefer_for: 10m
      options:
        - {pool: review-flagship, resources: [flagship], weight: 100}
        - {pool: review-workhorse, resources: [workhorse], weight: 10,
           when: "${risk} != 'high'"}
```

- `placement` replaces `pool` and `needs_resources` on agent steps.
- At run start each option's pool, resources and `when` are filled from
  run vars. `when` is `A == B` or `A != B` (sides trimmed, optional
  quotes). An option whose `when` is false is dropped for that run; none
  left fails the start. Options must share a project.
- Options are kept by weight, highest first. The step ticket starts on
  the top one: its pool, and `fields.resources` = required plus the
  option's resources.
- Each supervisor pass (before sessions are sized), for each open,
  unclaimed step with two or more options:
  - `placement.since` is set the first pass its blockers are finished.
  - Until `prefer_for` has passed since then, the step stays on the top
    option and waits there, as with plain `needs_resources`.
  - After that, it moves to the highest-weight option whose resource
    slots are all free (pool and `fields.resources` change together, with
    a comment). With none free it stays put. If a higher option frees
    before a session claims it, it moves back up.
- A claim pins the option: placement only moves unclaimed steps.
- `fields.placement` keeps the options, `chosen` and `since`. `run show`
  prints `placement option i/n pool P weight W (fallback)`.

Feasibility is resource slots only. Pool session caps are not checked:
a step on a full pool waits for a session like any other step.
