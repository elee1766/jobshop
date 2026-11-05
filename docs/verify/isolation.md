# Per-pool isolation (#388)

Needs `services.jobshop.isolation.enable = true` and
`services.jobshop.sandbox.enable = true` deployed on gasserver.

- `systemctl show jobshop -p Delegate` lists cpu memory pids, and
  `grep jobshop /etc/subuid /etc/subgid` shows the range.
- With no pool setting `isolation:`, `journalctl -u jobshop` has no
  "isolation" lines and sessions start as before.
- As the jobshop user in a checkout:
  `go test ./pkg/control -run TestIsolatedPool -v` passes, not skips
  (mapped uid on the probe file, denied CIDR unreachable, session `oom`).
- Add a pool with `isolation: {user: {uid: 1000, map_uid: <subuid start>,
  map_gid: <subgid start>, groups: [jobshop]}, limits: {memory: 4G}}` and
  `sandbox: {backend: bubblewrap, network: pasta, deny: [10.0.0.0/8]}`,
  reload and file a ticket for it:
  - `jobshop session sandbox <id>` shows the uid/gid map, the cgroup
    `<own>/sessions/<id>` with memory.max 4294967296 and swap.max 0, and
    network pasta with the deny list.
  - `stat -c %u <worktree>/<file the agent wrote>` is the subuid start.
  - `curl localhost:<metrics>/metrics | grep jobshop_pool_cgroup` has
    memory, cpu and oom_kills for the pool.
- Set `map_uid` outside the subuid range and reload: that pool is
  quarantined with one escalation naming the range; other pools keep
  working.
