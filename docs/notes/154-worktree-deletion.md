# Session worktree deletion (#154)

2026-11-02.

Findings, by suspect:
- `reap` (`session reap --closed`): only closed tickets older than the
  cutoff, and skips a dir whose host is alive. Not the cause.
- `removeWorktree`: called from `endSession` (ticket close of a session,
  runner finish, `shutdown`) and from reap. All paths act on a closed
  session after its host was stopped. Not the cause, but nothing enforced it.
- `git fetch --prune` prunes remote-tracking refs only; `git worktree
  prune` drops registrations whose dir is already gone. Neither deletes a
  live worktree.
- No gasserver timer or tmpfiles rule touches `sessions/` (the module's
  tmpfiles rules are the data dir, config link and gcroots).
- Cause: the refinery's `gh pr merge --delete-branch` (gh 2.101+) deletes
  the local branch and runs `git worktree remove` on the worktree holding
  it, which is the dev session that pushed the PR. Prompts stopped passing
  `--delete-branch` in bd917e2 (#67), but any agent or operator running
  gh or `git worktree remove` in the shared repo can still do it.

Fix:
- `ensureWorktree` adds the worktree with `--lock` (reason names the
  session). `git worktree remove --force` and `gh` then refuse it;
  `worktree prune` skips it. Before re-adding, the controller unlocks the
  stale entry so prune can drop it.
- `removeWorktree` takes a caller, WARNs `removing session worktree
  session=<id> caller=<path>`, refuses (WARN `refusing to remove worktree
  of live session`) while the ticket is open or the host is alive, and
  removes with `--force --force` to override its own lock.
- The respawn workaround WARN now carries `session` and
  `caller=runner.worktreeMissing`.
- `reap` WARNs before deleting a session dir.

Regression: `TestWorktreeDeleted` checks an outside `git worktree remove
--force` fails on a live session's worktree. Runbook:
`docs/verify/worktree-deletion.md`.
