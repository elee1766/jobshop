Step `assess` of a `blocker-check` run for ticket ${ticket} (issue #${issue}).

That ticket or issue says it is blocked or waiting on an operator. Decide
whether the blocker is still real, using current state, not what the
ticket says. Read the claim first:
- `jobshop ticket show ${ticket}` (comments and history; `none` means the
  claim is only on the issue)
- `gh issue view ${issue} --comments`

Then check each thing the claim depends on:
- Self-deploy: is the fix it waits for live? Compare master
  (`git fetch origin && git rev-parse origin/master`) with the running
  build (`jobshop version` prints `jobshop <version> (<commit>)`;
  `git merge-base --is-ancestor <needed sha> <commit>`).
- Pools: does the pool it needs exist and run sessions?
  `jobshop ticket list --orphaned` lists tickets whose pool is gone or
  has no sessions.
- Permissions and credentials: does the token, label, secret or access it
  needs exist now? Try the read-only form of the command that failed
  (e.g. `gh auth status`, `gh api repos/elee1766/jobshop`, `gh label list`).
- Any other dependency it names: a merged PR, a closed issue, a run
  (`jobshop run show <run>`).

Comment the evidence (commands and their output, short) on the issue
with `gh issue comment ${issue}` and on this step with `jobshop comment`.
Then do exactly one:

1. Stale (every check passes): route the work back. A ticket parked in a
   pool without sessions: `jobshop ticket update ${ticket} --pool <pool>`.
   A run in `needs_human`: `jobshop run resume <run>`. Remove the
   `needs-human` label if set (`gh issue edit ${issue} --remove-label
   needs-human`). Then `jobshop close`.
2. Real (a check fails): say which one and what would unblock it, then
   `jobshop close -r failed`. This step has one attempt, so the run
   escalates at once and the planner asks @elee1766.

Do not fix the blocker yourself and do not release this step (a release
counts as the failed attempt).
