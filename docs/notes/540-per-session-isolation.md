# Per-session isolation (#540)

Added 2026-11-05.

```yaml
isolation:
  user:
    uid: 1000
    map_uid: 231072
    map_gid: 231072
    groups: [jobshop]
    per_session: {range: 16}
```

- Each session gets a slot in `[0, range)` when it is created: the
  lowest not held by an open session of the pool. It is stored as
  `slot` in the session fields, and the agent runs as host
  `map_uid+slot` / `map_gid+slot`. `per_session` needs `groups`.
- The agent has no supplementary groups. Before spawn the controller
  sets POSIX ACLs (access and default) on the session dir, its sockets
  and the whole worktree: owner, the slot's uid and the first group get
  rwx, the owning group and others nothing. Another session's uid falls
  to "other". A filesystem without ACLs is a permanent spawn error
  that quarantines the pool.
- Scale up starts at most as many sessions as there are free slots.
  Session create fails once the range is full.
- A session with no slot (created before `per_session` was set) or a
  slot outside a range that shrank fails permanently at spawn.
- `session sandbox` shows the slot and the mapped ids.
