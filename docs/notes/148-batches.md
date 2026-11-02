# Batches (#148)

Design note, 2026-11-02. Part of #145. No code yet.

A batch is the Gas City "convoy": an ordered set of pr-work runs the
planner starts together and tracks as one thing.

## What happened in the chat track (#121-#125)

Times are UTC on 2026-11-02, from `jobshop ticket show` history of the
runs and their steps, and the issue and PR timelines.

| Issue | Run | Run created | PR | Run done |
|---|---|---|---|---|
| #121 send/subscribe | js-003283hvrr002 | 06:18 | #128, merged 07:01 | 07:05 |
| #124 operator pack | js-003283jx04002 | 06:18 | #127, merged 07:12 | 07:27 |
| #122 chat pools | js-0032a8wvfw002 | 11:21:34 | #137, merged 12:30 | 13:30 |
| #123 gateway | js-0032a8wvjw002 | 11:21:34 | #138, merged 13:34 | 14:36 |
| #125 escalation threads | js-0032a8wvng002 | 11:21:34 | #142, merged 15:48 | 16:47 |

- Order wanted: #121, then #122 -> #123 -> #125; #124 in parallel. The
  issue bodies said so in prose ("Depends on ...").
- #121 and #124 started at 06:18 with no edges. #124 shipped with a
  workaround (`max_sessions: 0`, started by hand) because #122 was not
  built yet.
- #122, #123, #125 were started at 11:21:34. The planner then added
  cross-run edges with `jobshop edge add`: #123 `implement` blocked by
  #122 `post-merge`, and #125 `implement` blocked by #123 `post-merge`,
  at 11:21:53.
- The edges came too late. The dev pool had claimed all three
  `implement` steps at 11:21:45-46, seven seconds earlier. Edges do not
  affect a ticket already claimed.
  - #123 was built in parallel with #122 anyway. It worked because the
    gateway only needed the send/subscribe API from #121.
  - #125's session found no gateway on master, closed `failed` at
    11:22:00 (attempt 1 of 3) and commented why. The step reopened, the
    edge now held, and it was claimed again at 14:36:42, right after #123
    `post-merge` closed. One retry burned on ordering, not on a real
    failure.
- Status was tracked by hand: a checklist on target #82, ticked by the
  planner in its pass comments with PR and CI evidence ("Ticked #122:
  PR #137 merged ..."). Each pass re-listed which runs were active. There
  was no single object answering "is the chat track done, and what is
  it waiting on".
- No run in the track escalated. Total wall time for the chained three:
  11:21 to 16:47.

## Proposal

A batch is a ticket of kind `batch` whose children are pr-work runs.
No new store table: tickets, `parent_id`, `fields`, edges and graph apply
already carry everything.

- Ticket: kind `batch`, `fields.key` (unique per project like run keys),
  `fields.issues` (ordered list), `fields.state`
  (`open` / `running` / `needs_human` / `done`, plus `failed` if a run
  is closed failed for good). No pool, so it is never claimed.
- Members: one pr-work run per issue, `parent_id` = the batch, run key
  `issue-<n>` as today. If a run with that key exists it is adopted
  (re-parented) instead of started again.
- Order: `--order` takes `a>b` pairs (or `a>b>c` chains). Each pair
  becomes a `blocks` edge from b's first step to a's last step
  (`post-merge`), so b starts only after a is merged and deployed.
  Without `--order` the issues are independent.
- Creation is one graph apply: batch ticket, runs, steps and the
  cross-run edges in a single transaction. This fixes the race above:
  no step is claimable before its edges exist.
- State is derived by the controller in the same place runs derive
  theirs: `running` once any member step is claimed, `needs_human` if
  any member run is `needs_human`, `done` when all member runs are
  `done`. The batch closes `done` with the last run.
- Escalation stays per run. A batch in `needs_human` only points at
  the member run; it files no ticket of its own.

CLI:

- `jobshop batch start --key chat-track --issues 122,123,125 --order 122>123>125`
  starts or returns the existing batch for the key.
- `jobshop batch show <id|key>`: state, then one line per member
  (issue, run id, run state, current step, blocked by which issue).
- `jobshop batch list`: open batches.
- The planner prompt uses `batch start` for any group of issues that a
  target lists with dependencies, and quotes `batch show` in its pass
  comments instead of re-listing runs.

Data model change: none in the schema. New ticket kind `batch` and its
`fields` (key, issues, order, state); new admin RPCs `batch.start`,
`batch.show`, `batch.list` built on graph apply and ticket queries.

## Out of scope

- Batches across repos or projects.
- Workflows other than pr-work as members (the shape allows it later).
- Merging a batch as one PR, or holding merges until all members pass.
- Changing the target checklist: the planner still ticks boxes with
  evidence. A batch is tracking, not acceptance.

## Open questions

- Should adding an issue to a running batch be allowed
  (`batch add`), or is a new batch simpler?
- When a member run fails for good, do dependents stay blocked (current
  edge semantics) or does the batch cancel them?
- Should a step whose edges change after claim be released back
  (general fix for the race), independent of batches? Filed as #149.
- Is `post-merge` the right edge target, or is `review-merge` (merged,
  not yet deployed) enough for code-only dependencies?

## Follow-up implementation issues

1. Batch ticket kind, `batch.start` via graph apply with cross-run edges,
   adopting existing runs by key.
2. Batch state derivation in the controller and close on last run.
3. `jobshop batch show` and `batch list`.
4. Planner prompt: start batches for dependent issue groups and report
   with `batch show`.
