# Session cgroups as systemd scopes (#944, JOB-97)

2026-11-06.

## Problem

With the cgroupfs driver the controller unit has `Delegate=cpu memory
pids` and enables those controllers in its cgroup for session cgroups
under `<unit>/sessions`. Sessions outlive the controller
(`KillMode=process`), so on a restart the unit cgroup still has
controllers enabled for a populated child, and systemd cannot clone3 the
next controller into it (EBUSY, `Result=resources`). Disabling the
controllers on stop resets live sessions' limits to `max`.

## What changed

- `limits.cgroup_driver: systemd` (default `cgroupfs`, unchanged) and
  `limits.sessions_slice` (default `jobshop-sessions.slice`). With
  systemd a limited pool's agent gets `cgroup.DriverSystemd` with the
  slice as `Parent`: acphost asks the manager the controller runs under
  for a transient scope `acphost-<session>-<pid>.scope` with the limits
  as unit properties (acphost v0.21.0, its note 226).
- Nothing is delegated or set up: no `<own>/controller`, no
  `<own>/sessions`, no move of the controller or of hosts. The unit
  cgroup enables nothing, so a restart works with live limited sessions,
  and the scopes keep their limits (systemd owns them, not the unit).
- Session hosts stay in the controller's unit, as without isolation:
  `KillMode=process` keeps them across a restart. Their agents are in
  the scopes.
- `limits.cgroup` (JOB-91) with systemd: its presence gives every
  agent a scope in the slice, also in pools without limits, so all agent
  load counts against the slice. `sessions_cpu` and the weights are not
  written by the controller (a load warning says so): they are
  `CPUQuota=`/`CPUWeight=` of the slice and `CPUWeight=` of
  jobshop.service, set by the module.
- Preflight sends the systemd cgroup, so acphost checks `cgroup.slice`
  and `cgroup.systemd` instead of the cgroupfs parent checks.
- Pool metrics read each session's scope dir from its `isolation.json`
  (the scope name holds a pid). `session sandbox` shows the live scope.
- A driver change is in the isolation part of the env hash, so idle
  sessions respawn into the new kind of cgroup on reload.

## Module

- `services.jobshop.isolation.cgroupDriver`: `systemd` by default under
  home-manager, `cgroupfs` on NixOS. With systemd: no `Delegate=` or
  `DelegateSubgroup=` on jobshop.service, `CPUWeight=` (controller
  weight, default 1000) on it, and `jobshop-sessions.slice` (nested in
  `jobshop.slice` by name) with `CPUQuota=` from `sessionsCpu` and
  `CPUWeight=` (sessions weight, default 100). The slice is
  `X-SwitchMethod=keep-old` like `jobshop.slice`.
- NixOS asserts cgroupfs: the controller runs as the jobshop user in a
  system unit, and the system manager's private socket is root only, so
  it has no manager to ask. Running the NixOS module with systemd would
  need a user manager for jobshop (the home-manager setup).

## Not covered

- Memory is still capped on `jobshop.slice` only, as before.
- Moving a live host from cgroupfs to systemd: a host adopted after the
  switch keeps its old cgroup until its session respawns (env hash).
  The old `<unit>/sessions` tree is left in the unit cgroup until those
  sessions end.

Check: `docs/verify/systemd-cgroup-driver.md`. Tests:
`TestSystemdScopeRestart` (real user manager, skips without one),
`TestPoolCgroupSystemd`, `TestCgroupDriver`, and the module checks
(`hmIso`, "home-manager isolation uses the systemd cgroup driver").
