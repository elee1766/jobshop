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
- Before each spawn the controller checks that all of
  `[map_uid, map_uid+range)` and `[map_gid, map_gid+range)` lie in one
  subuid and one subgid range of its user (acphost checks only the id in
  use). If not, the spawn fails permanently and the pool is
  quarantined, as in #388.
- NixOS: with `isolation.enable`, an assertion fails the build when a
  pool's per-session ids in `settings` run past `subIdStart +
  subIdCount`.
- Tests: `TestSlotAllocation`, `TestSlotRestart`, `TestCheckSlotRange`,
  and `TestPerSessionIsolation` (two sessions, distinct host uids,
  cross-read denied; skipped without the #388 host setup).
