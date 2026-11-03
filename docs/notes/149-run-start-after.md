# run start --after (#149)

2026-11-02. `jobshop run start ... --after <id>` (repeatable; RPC field
`after`) adds a `blocks` edge from the run's first step to each id in
the same graph apply that creates the steps. The step is never ready
before its outside blockers exist. Before, the planner started runs and
then ran `jobshop edge add`; the dev pool claimed the steps in the
seconds between, and edges do not affect a claimed ticket (chat track,
`docs/notes/148-batches.md`).

- An id may be a run: a run closes when its last step closes, so
  `--after <run>` waits for merge and deploy of that run.
- Unknown ids fail the start. Sessions can only name tickets in their
  project.
- A repeated start that returns an existing run by key adds no edges.
- The planner prompt (`packs/github/prompts/plan.md`) now uses `--after`
  instead of `edge add`.
- `edge add` onto an already-claimed ticket still does not release it.
  This was the issue's other option. It was not built, because releasing
  a live session's work is a judgment call.
