# Batch state (#170)

2026-11-03. Follow-ups 2 and 3 of `docs/notes/148-batches.md`.

The controller derives the batch's `fields.state` wherever it derives
run state, from its member runs:

- `running` once any member step is claimed.
- `needs_human` while any member run is `needs_human`; it returns to
  `running` after `run resume`.
- `done` when every member run is done. The batch closes `done` with
  the last run.
- `failed` as soon as a member run closes without `done` (failed for
  good or cancelled). The batch closes `failed`.

A batch has no escalation ticket of its own. Its member runs escalate.

RPCs `batch.show` (id or key) and `batch.list` (open batches). CLI
`jobshop batch show <id|key>` prints the state, then one line per
member: issue, run id, run state, current step, and which issues block
it. `jobshop batch list` prints open batches.
