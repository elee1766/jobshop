# Pool quarantine on a permanent spawn error (#146)

2026-11-02. Part of #145. On 2026-11-02 01:32-05:40 every spawn failed
on a missing repo path, and the pool retried for hours: the crash-loop
rule (3 crashes in 10m, then a 15m pause) treats every failure alike.

**Classification** (`pkg/host/spawn.go`). `host.Permanent(err)` is true
for a `host.PermanentError`, `fs.ErrNotExist` or `exec.ErrNotFound`.
Before writing the spec the controller runs `host.CheckSpawn`: the
session cwd must exist (unless it is a worktree just made), and the
agent command must resolve, against the PATH in the spawn env. In
`ensureWorktree` a missing pool repo is permanent, and so is a
`git worktree add` failure whose output says "not a git repository",
"invalid reference", "not a valid object name", "Permission denied" or
"No such file or directory". Anything else (fetch failures, a lock) stays
transient and keeps the 3-in-10m rule.

**Quarantine.** A permanent spawn error quarantines the pool on the
first failure, before the session's tickets are released, so no tick
starts another session into the same error. The controller logs WARN
`pool quarantined: permanent spawn error` with the error and files one
escalation ticket per pool: title `Pool <name> quarantined: <error>`,
label `needs-human`, meta `{"pool": {"name", "quarantine"}}`, in the
first workflow `escalate_pool` of the pool's project that is not the
pool itself. An open one for the pool is not filed again. The gateway
opens an escalation thread for it like a run escalation, with a prompt
about the pool.

The quarantine is in memory and lasts until `jobshop reload` (which
clears every permanent quarantine, since the fix is a config change),
`jobshop pool unquarantine <name>` (also clears the crash-loop pause),
or a controller restart. `jobshop pool list` shows quarantined pools
with their error and paused pools with their end time. Pools can grant
`pool-list` and `pool-unquarantine`.
