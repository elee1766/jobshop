# Implement routing by difficulty (JOB-59)

2026-11-06. pr-work and pr-work-linear take a `dev_pool` run var
(pack default `dev`): the `implement` step starts in that pool. The
github pack adds a `dev-sonnet` pool (same prompt; `dev_sonnet_model`
`claude-sonnet-5-5`, effort `high`, 1 session) and `dev_model` /
`dev_effort` for `dev`.

Deviation from the issue: the strong pool keeps the name `dev` instead
of `dev-opus`, so open tickets, host overrides (`dev_sessions`), memory
and learning scopes on `pool:dev` keep working without a migration.

- New step field `retry_pool` (a template, like `pool`). At run start
  it resolves to a configured pool (an unknown one fails the start) and
  is stored as `fields.retry_pool`; equal to the start pool it is
  dropped. A failed attempt (close failed, release, expired lease) or a
  `run resume` of a failed or finished step moves the open step to it,
  with a comment. Nothing moves it back. `run show` prints `retry pool`.
- The `dev-pool` prompt rule (plan, dev, dev-sonnet pools) gives the
  rubric: a `model:sonnet` / `model:opus` label wins, otherwise clear
  scoped work goes to `dev-sonnet`, investigation, design, cross-cutting
  or security work and anything unsure to `dev`. The starter comments
  the choice on the issue. `batch start` takes no vars, so batch runs
  stay on `dev`.

Still to do after deploy: a week of numbers (runs per pool, retry rate
to `dev`, tokens per merged PR per pool) in a follow-up note.
