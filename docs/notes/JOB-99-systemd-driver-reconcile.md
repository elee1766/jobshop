# JOB-99: systemd cgroup driver, reconciled

2026-11-08.

JOB-99 asked for session hosts and agents in transient scopes under a
sessions slice, with no `Delegate=` on jobshop.service. That landed
under #944 / JOB-97:

- `limits.cgroup_driver` and `limits.sessions_slice`, agents in
  `acphost-<session>-<pid>.scope`, pool limits as scope properties, and
  the module's systemd mode (no `Delegate=`/`DelegateSubgroup=`,
  `jobshop-sessions.slice` with `CPUQuota=`/`CPUWeight=`): #946, see
  `944-systemd-cgroup-driver.md`.
- Session hosts in `acphost-host-<session>-<pid>.scope`: #993, see
  `992-host-scope.md`.

This change adds what the acceptance asked for and was missing: a
second restart and a kill -9 of the controller in
`docs/verify/systemd-cgroup-driver.md` (with `MemoryMax`/`TasksMax`
read from the scope unit), and a second restart in
`TestSystemdScopeRestart`. `docs/verify/controller-cgroup.md` stays the
cgroupfs check and points at it.

Turning the driver on for gasserver is its host config
(`services.jobshop.isolation.cgroupDriver`, the home-manager default
once `isolation.enable` is set), not this repo.
