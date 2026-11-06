# Reserve CPU for the controller (JOB-91, 2026-11-06)

A session's node source build pegged `jobshop.slice` and the admission
path, which runs in the controller, timed out (19 events denied). The
controller and every session shared one cgroup tree with no weights.

## What changed

- `limits.cgroup` in jobshop.yaml splits the delegated cgroup:
  `<own>/controller` (the controller, its hooks, git) and
  `<own>/sessions` (every session host and agent). Session hosts now start
  in `<own>/sessions/hosts` through `CgroupFD`, so they count against the
  sessions share even when their pool has no `isolation.limits`. Agents of
  pools with limits keep their own child of `sessions` as before.
- Keys: `sessions_cpu` (CPUs, written to `sessions/cpu.max`),
  `sessions_cpu_weight` (default 100), `controller_cpu_weight` (default
  1000). Written at start and on every reload; an unset cap lifts the old
  one. Absent block: nothing changes, no cgroup is touched.
- With the block set, the controller sets up its cgroup at start even when
  no pool has limits. Without delegation it logs a warning and runs as
  before. Hosts left in `controller` by a restart (or by the deploy that
  turned this on) are moved to `hosts` once at start.
- Module (needs `isolation.enable`, for `Delegate=`): `sessionsCpu`
  (default: `cpuQuota` minus one CPU, at least one, none if `cpuQuota` is
  not a percentage), `sessionsCpuWeight`, `controllerCpuWeight`. On the
  default 600% quota sessions get 5 CPUs and the controller can always use
  the sixth, with 10 times a session's weight above that.

## Not covered

- Hooks run through `runIsolated` for an isolated pool land in that pool's
  session cgroup, so a heavy command step shares sessions' cap.
- Memory is not split: `memoryMax` stays on the slice.
- Weight only matters when the slice is at its quota or the host is busy;
  the cap is what guarantees the spare CPU.

Check: `docs/verify/controller-cgroup.md`. Tests: `TestHostCgroup`,
`TestTuneCgroupAt`, `TestMoveHosts`, `TestSetupCgroupAt`, and the module
check (`cgDefault`, `cgSet`, `cgOff`).
