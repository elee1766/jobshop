# Per-pool isolation (#388)

Needs `services.jobshop.isolation.enable = true` and
`services.jobshop.sandbox.enable = true` deployed on gasserver.
The `sh readonly` blocks only read state and fail when the check fails
(`jobshop runbook run docs/verify/isolation.md`, #635); the rest is for
the operator.

Host setup: cgroup delegation and a subuid/subgid range for jobshop.

```sh readonly
d=$(systemctl show jobshop -p DelegateControllers --value)
echo "DelegateControllers: $d"
for c in cpu memory pids; do echo "$d" | grep -qw "$c" || { echo "FAIL: $c not delegated"; exit 1; }; done
grep '^jobshop:' /etc/subuid || { echo "FAIL: no subuid range"; exit 1; }
grep '^jobshop:' /etc/subgid || { echo "FAIL: no subgid range"; exit 1; }
```

Running sessions of an isolated pool: at least one open session has a
mapped user, a session cgroup and, when its pool sets `network: pasta`,
the pasta network. Each mapped uid is inside the subuid range.
`session sandbox` needs the admin socket (#677): run this as the
jobshop user on the host. Without it the block stops with one
"cannot read sandbox" line instead of reporting no isolated session.

```sh readonly
start=$(grep '^jobshop:' /etc/subuid | cut -d: -f2)
count=$(grep '^jobshop:' /etc/subuid | cut -d: -f3)
n=0
ids=$(JOBSHOP_SOCKET= jobshop session list 2>&1) || { echo "FAIL: cannot list sessions: $ids"; exit 1; }
for id in $(echo "$ids" | jq -r '.[] | select(.alive) | .id'); do
  s=$(JOBSHOP_SOCKET= jobshop session sandbox --json "$id" 2>&1) || { echo "FAIL: cannot read sandbox: $s"; exit 1; }
  echo "$s" | jq -e . >/dev/null 2>&1 || { echo "FAIL: cannot read sandbox: $s"; exit 1; }
  echo "$s" | jq -e '.user' >/dev/null || continue
  n=$((n + 1))
  echo "$s" | jq -c '{session, pool, user, cgroup_path, cgroup, network: .sandbox.network, deny: .sandbox.deny}'
  mu=$(echo "$s" | jq -r '.user.map_uid')
  [ "$mu" -ge "$start" ] && [ "$mu" -lt $((start + count)) ] || { echo "FAIL: $id map_uid $mu outside $start+$count"; exit 1; }
  cg=$(echo "$s" | jq -r '.cgroup_path // empty')
  [ -n "$cg" ] && [ -d "$cg" ] || { echo "FAIL: $id cgroup '$cg' missing"; exit 1; }
  mm=$(echo "$s" | jq -r '.cgroup.memory_max // empty')
  if [ -n "$mm" ]; then
    [ "$(cat "$cg/memory.max")" = "$mm" ] || { echo "FAIL: $id memory.max $(cat "$cg/memory.max") != $mm"; exit 1; }
  fi
done
[ "$n" -gt 0 ] || { echo "FAIL: no open session of an isolated pool"; exit 1; }
echo "ok: $n isolated sessions"
```

Pool cgroup metrics are exported:

```sh readonly
m=$(curl -s 127.0.0.1:9464/metrics | grep '^jobshop_pool_cgroup' || true)
echo "$m"
for k in memory cpu oom_kills; do echo "$m" | grep -q "jobshop_pool_cgroup_$k" || { echo "FAIL: no jobshop_pool_cgroup_$k"; exit 1; }; done
```

Operator steps (change config, run tests or files a ticket):

- With no pool setting `isolation:`, `journalctl -u jobshop` has no
  "isolation" lines and sessions start as before.
- As the jobshop user in a checkout, `TestIsolatedPool` passes, not
  skips (mapped uid on the probe file, denied CIDR unreachable, session
  `oom`):

  ```sh
  go test ./pkg/orchestration/control -run TestIsolatedPool -v
  ```

- Add a pool with `isolation: {user: {uid: 1000, map_uid: <subuid start>,
  map_gid: <subgid start>, groups: [jobshop]}, limits: {memory: 4G}}` and
  `sandbox: {backend: bubblewrap, network: pasta, deny: [10.0.0.0/8]}`,
  and `credentials: ["~/.jcode/auth.json"]` on its agent (JOB-19). The
  module defaults (`sandbox.rw` under `~`, a 0700 dataDir) work as is:
  the controller binds `~/` paths from `<dataDir>/homes/<pool>`.
  Reload and file a ticket for it. The readonly blocks above then check
  its session; also the file the agent wrote is owned by the subuid start:

  ```sh
  JOBSHOP_SOCKET= jobshop reload
  JOBSHOP_SOCKET= jobshop ticket create "isolation probe" --pool <pool>
  stat -c %u <worktree>/<file the agent wrote>
  ```

- Set `map_uid` outside the subuid range and reload: that pool is
  quarantined with one escalation naming the range; other pools keep
  working.
