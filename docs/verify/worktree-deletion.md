# Session worktrees survive outside removal (#154)

After deploy, with at least one dev session running:

```sh
git -C <pool repo> worktree list --porcelain | grep -A1 'sessions/'
git -C <pool repo> worktree remove --force /var/lib/jobshop/sessions/<id>/work
journalctl -u jobshop --since -24h -o cat \
  | grep -E 'worktree missing, respawning|removing session worktree|refusing to remove worktree'
```

Pass:
- Each session worktree shows `locked jobshop session <id>`.
- The `worktree remove --force` fails with "locked"; the dir stays.
- Over 24h: no `session worktree missing, respawning host` line, and every
  `removing session worktree` line names a session that had ended.
