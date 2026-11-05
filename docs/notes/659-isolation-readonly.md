# Isolation runbook readonly checks (#659)

2026-11-05. `docs/verify/isolation.md` now has three `sh readonly`
blocks that `live-check` (#635) runs unattended, each exiting non-zero
on a failed check:

- host setup: `DelegateControllers` lists cpu, memory and pids; jobshop has
  subuid and subgid ranges;
- running sessions: at least one open session has a mapped user
  (`jobshop session sandbox`), its `map_uid` is inside the subuid range,
  its cgroup dir exists, and `memory.max` matches the pool limit; the
  user, cgroup and network (pasta, deny list) are printed;
- metrics: `jobshop_pool_cgroup_{memory,cpu,oom_kills}` are exported.

Left for the operator: journal check with no isolated pool,
`TestIsolatedPool`, adding the pool, the worktree file owner, and the
out-of-range quarantine. `per-session-isolation.md` is unchanged: every
step changes config or ends a session.
