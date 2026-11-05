# #728: isolation preflight at config load (2026-11-05)

At start, after adopting running sessions, and after every good reload,
the controller runs acphost `Preflight` for each pool that uses a
bubblewrap sandbox, `isolation.user` or `isolation.limits`. The spec has
the sandbox network, the mapped user (with the per_session range as
MapCount) and the cgroup limits under the controller's sessions cgroup.

A failing report quarantines the pool with `isolation preflight:` and
the report text as the reason, so `jobshop pool list` and the
needs-human escalation list each failed check with its fix line. No
session spawns until a reload passes or `jobshop pool unquarantine`.
A reload clears every quarantine first and then runs preflight again.

Pools without isolation are not checked. When the controller's own
cgroup setup failed, the cgroup is left out of the spec, and the spawn
reports that error as before (applyIsolation).
