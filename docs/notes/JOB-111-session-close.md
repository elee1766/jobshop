# JOB-111: session close

2026-11-06. `session.close {id, handoff, reason, dry_run}` ends one open
session now and frees its pool slot, for a finished operator chat or a
session nobody needs. `session drain` waits for idle; this does not.

- Order: the runner and host stop first (a running turn is cancelled), so
  the session claims nothing after its held list is read. Each claimed
  ticket is handed back to its own pool, or to `handoff`, with a comment
  naming who closed the session and why. That is a handoff, so a step does
  not count a failed attempt. Pre-assigned tickets are unassigned (or
  handed to `handoff`). The session ticket closes with resolution
  `closed`, then `endSession` runs as for any ended session: socket,
  worktree, slot home, leftover claims.
- A claimed ticket with no pool is released by `endSession`, which counts
  a failed attempt for a step.
- Kept: the session ticket and its comments, `journal.ndjson` (Desk's
  transcript and `session.journal`), the run and its steps. `session reap
  --closed` removes the dir later, as before.
- `dry_run` returns `held: [{id, title, kind, status, pool}]` and changes
  nothing. Desk shows it in the confirm.
- Access: admin, or a pool with the new grant `session-close` for sessions
  in its own project. A session cannot close itself (the call would wait
  on its own turn); it closes or releases its ticket instead. No pack
  grants it yet.
- CLI: `jobshop session close <id> [--handoff <pool>] [-r <reason>]
  [--dry-run]`.
- Desk: Close on each open session row (Sessions view) and in the session
  pane header. Both run the dry run, confirm with the held list, then call
  `session.close` through `/api/rpc` with reason "closed in Desk".
- Tests: `TestSessionClose` (pkg/control), `ui/e2e/session-close.spec.js`
  (mocked API).
