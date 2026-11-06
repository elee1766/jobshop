# Drop the host.lock/host.pid fallback (#918, 2026-11-06)

jobshop bumps acphost to v0.20.0, which removed `host.lock`, `host.pid`,
`acphost.LockFile` and `acphost.PidFile` (acphost#222). The pre-v0.17
fallback from #904 (`legacyHostAlive`, `legacyHostPid`, `stopLegacyHost`)
is gone from `pkg/control/hostproc.go`: with no `supervisor.pid` in a
session dir, `hostAlive` is false, `hostPid` is 0 and `stopDetachedHost`
returns false. `docs/notes/904-start-detached.md` describes the old
fallback; it no longer applies.

Before deploying this, check no pre-v0.17 host still runs: a session dir
with a flocked `host.lock` and no `supervisor.pid`. Stop any with the
previous build first, or the new controller will not see it.
