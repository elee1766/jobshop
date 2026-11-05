# Per-pool isolation (#388)

2026-11-05. Needs acphost v0.3.0.

Optional. With no `isolation:` key and no `network: pasta`, nothing below
runs and sessions start as before.

Config, under `defaults`, agents and pools (`user` and `limits` each taken
whole from the most specific level; `deny` appended like sandbox paths):

```yaml
isolation:
  user: {uid: 1000, map_uid: 231072, map_gid: 231072, groups: [jobshop]}
  limits: {cpu: "4", memory: 8G, memory_high: 6G, pids: 4096}
sandbox:
  network: pasta
  deny: [10.0.0.0/8, 100.64.0.0/10, 192.168.0.0/16]
```

- `user`: the agent runs as uid/gid (gid defaults to uid) in a user
  namespace mapped to `map_uid`/`map_gid`, which must be in the controller
  user's /etc/subuid and /etc/subgid. `groups` (names or gids the
  controller holds) stay supplementary groups. The first group owns the
  session dir, cli.sock and the worktree, made group rw and setgid, so the
  controller and refinery can still read and remove them.
- `limits`: when some pool sets them the controller finds its cgroup at
  start (`acphost.DelegatedCgroup`), moves itself into `<own>/controller`,
  enables cpu, memory and pids, and puts each session in
  `<own>/sessions/<session>`. `cpu` is a CPU count, memory takes K/M/G/T.
  With `memory` set and no `swap`, `memory.swap.max` is 0, so the limit
  OOM kills instead of swapping.
- `network: pasta` and `deny` pass to acphost's sandbox; they need
  `backend: bubblewrap`.
- Problems are permanent spawn errors and quarantine only that pool with
  one escalation (#146): no delegated cgroup, ids outside subuid/subgid,
  missing newuidmap or pasta, an unknown group name.
- A session whose agent the memory limit killed ends with resolution
  `oom` instead of restarting.
- `jobshop session sandbox <id>` shows the uid/gid map, the cgroup path
  and limits, the network mode and deny list.
- Metrics per pool, from the live session cgroups each tick:
  `jobshop.pool.cgroup.memory`, `.cpu` (usage_usec) and `.oom_kills`.
- Changing isolation config is hashed into EnvHash, so idle sessions
  respawn on reload.
- NixOS: `services.jobshop.isolation.enable` sets `Delegate=cpu memory
  pids`, a subuid/subgid range (`subIdStart`, `subIdCount`), pasta and
  /run/wrappers (newuidmap) on PATH.

Tests: `TestIsolationMerge`, `TestIsolationErrors`, `TestSetupCgroupAt`,
`TestApplyIsolationNoDelegation`; `TestIsolatedPool` runs a mapped uid,
memory limit and pasta deny end to end and skips without the host setup
(docs/verify/isolation.md).
