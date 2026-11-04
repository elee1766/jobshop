# Learnings capture (#264)

2026-11-04. First item ("Capture") of target #247 B.

- `jobshop learn correction|failure|workaround "<summary>" [--fix]
  [--scope global|repo:<owner/name>|pool:<name>] [-t <ticket>]` calls
  `learn.add` on the controller. A session's ticket defaults to the one
  it holds; scope defaults to that ticket's repo (its or its run's
  `meta.github.repo`), else its pool, else `global`. Bad kinds, empty
  summaries, malformed scopes and unknown tickets are rejected.
  `jobshop learn list [--since] [--scope] [--kind] [--status] [-t]` reads
  them newest first. Both are open to every session: workers record, the
  curator reads.
- Storage: a `learnings` table (migration 00003): id, kind, summary, fix,
  scope, ticket, session, source (agent|controller), status
  (open|applied|rejected, for the curator), created_at. A table, not
  non-claimable tickets: pools, `ready`, orphan and stall checks never
  see them, no history rows or change events per learning, and the
  curator filters by scope and time with an index.
- The controller records (source `controller`), never failing the action:
  - a step closed failed or released without `--until` (a failed
    attempt): one `failure` with the step, run, count and reason (the
    worker's last `report failed|dropped` note, or the release reason);
  - `run resume` of a step that closed done (a review finding): one
    `correction`, with the resuming session's last comment on its own
    ticket as the finding;
  - an escalation filed for a run: one `failure` on the run;
  - `jobshop answer` to a question: one `correction` on the asked
    ticket, the answer as `fix`;
  - a stall trigger firing: one `failure` with signal, subject and detail.
- The operator prompt records every operator correction with
  `jobshop learn correction` before making it durable.
- Follow-up: reverted PRs as learnings need a revert detector (a
  subscription or the activity feed); not built here.
