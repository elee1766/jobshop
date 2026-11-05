# Delegated cgroup survives a controller restart (JOB-21)

2026-11-05.

- With isolation, the controller enables cpu, memory and pids in the unit
  cgroup's `subtree_control` and puts sessions in `<unit>/sessions/<id>`.
  Session hosts outlive a controller stop (KillMode=process).
- Before: on the next start systemd placed the new main process in the
  unit cgroup itself. A cgroup with controllers enabled for children
  cannot hold processes, so the spawn failed with EBUSY ("Failed to spawn
  executor: Device or resource busy") and restarted in a loop.
- Now the module sets `DelegateSubgroup=controller` (systemd 254+):
  systemd starts the controller in `<unit>/controller`, the leaf the
  controller used to move itself into. `setupCgroupAt` already treats a
  start inside `controller` as a restart and works on the parent.
- Sessions in `<unit>/sessions/*` are untouched and re-adopted as without
  isolation.
- A host already stuck in the loop recovers on the switch that adds
  `DelegateSubgroup`: no root cleanup is needed, since systemd no longer
  writes to the unit cgroup's `cgroup.procs`.
- Test: the module check "isolation starts the controller in the
  controller subgroup". Runtime check: enable isolation, start a session,
  `systemctl restart jobshop`, then `systemctl is-active jobshop` and the
  session is still listed open.
