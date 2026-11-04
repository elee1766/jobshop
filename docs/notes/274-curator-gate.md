# Curator gate (#274)

2026-11-04. Item "Gate" of target #247 B. Follows the curator (#273).

- Record: when the curator sets learnings `applied` (`jobshop learn
  update <ids> --status applied [--merged-at <t>]`), the controller
  records one row per PR in a new `memory_edits` table (migration
  00006): the PR, the cited learning ids, the scoped pools and a
  baseline. A store row, not run meta: the PR merged long after the
  curate run closed, and the gate queries pending edits by state.
  Recording again for the same PR is a no-op.
- Scoped pools: a learning's `pool:` scope, else the pool of its ticket
  (for a run, its steps' pools). None found: every pool.
- Metrics (`PoolMetrics`) over steps of those pools created in a window,
  and learnings recorded in it: runs, steps, retries (summed
  `fields.failures`), escalations (steps closed failed at the limit),
  review findings (`review resumed` corrections on those steps) and
  recurrence (learnings of a cited learning's kind whose summary shares
  at least half its words of 4+ letters). Baseline: the 7 days before
  the merge. After: merge to now.
- Due: `evaluate_runs` runs (default 20) in the pools since the merge,
  or `evaluate_age` (default 72h). Trigger condition `when.edits: {runs,
  age}` holds while any pending edit is due; the selfdev `evaluate`
  trigger starts the `evaluate` workflow (one step in `learn`, key
  `evaluate`).
- Decision (`jobshop learn evaluate`): not helped when the cited problem
  still recurs at least as often per step as before, or retries plus
  escalations plus findings per step rose by more than 25% (and more
  than one). No steps since the merge counts as helped. The result goes
  on the edit (`state`, `after`, `reason`) and on each cited learning
  (`gate`), shown by `jobshop learn list` and `jobshop learn edits`.
- Not helped: the evaluate step files an issue, opens a revert PR, starts
  `pr-work` for it and marks the cited learnings `rejected` with the
  original PR, so the curator does not propose the edit again.
- Tests: `TestGateMetrics`, `TestGateDecide`, `TestCuratorGate`,
  `TestCuratorGateAge`, `TestSelfdevPackLoads`.
