# Session hosts in a systemd scope (#992)

2026-11-07.

With `limits.cgroup_driver: systemd`, #944 left session hosts in the
controller's unit cgroup (`KillMode=process` kept them across a restart),
and with them every agent without a pool scope. A restart of the unit
then met a non-empty unit tree, the case acphost note 226 moved agents
out of.

acphost v0.22.0 adds `acphost.InScope(slice)` (acphost#240). With the
systemd driver `startDetachedHost` now starts each host in a transient
scope `acphost-host-<session>-<pid>.scope` in `limits.sessions_slice`
(default `jobshop-sessions.slice`). The manager owns the scope, so the
host and an agent without its own scope count against the slice and are
outside jobshop.service. systemd collects the scope once the host and
its children exit. No limits, no `Delegate`.

cgroupfs is unchanged: `InCgroup(<own>/sessions/hosts)` when
`limits.cgroup` asks for the split, else the controller's cgroup.

A scope that cannot be started (no manager, a caller outside its
manager's tree) logs "session host stays in the controller cgroup" and
the host starts without it, as the cgroupfs fallback does. A
`*acphost.SpawnError` (the agent's setup failed) is returned as is: a
start without the scope would fail the same way.

Hosts already running when the controller moves to this build keep
their cgroup until their session respawns.

Tests: `TestHostPlacement` (option per driver),
`TestDetachedHostScopeFallback` (no manager), and
`TestSystemdScopeRestart` now checks the host pid's cgroup is
`<slice>/acphost-host-*.scope` (real user manager, skips without one).
