# pkg/control/sessionrt: session process and isolation mechanics (JOB-146, 2026-11-08)

Slice 1 of JOB-146. JOB-146 stays open: its remaining slices are
JOB-152, JOB-153 and JOB-154. No behavior change.

## Boundary

`pkg/control/sessionrt` holds the mechanics under a session host. It has
no ticket, pool or session state: callers pass directories, ids, specs
and config values. It does not import `pkg/control`, and it builds on
acphost (`proc`, `cgroup`, `userns`, `sandbox`) rather than wrapping it.

| File | Owns |
|---|---|
| `host.go` | `HostAlive`, `HostPid`, `StopDetached` (SIGTERM, wait on `supervisor.lock`, SIGKILL), `StopProcs` (agent daemons that outlive the host, found by `JOBSHOP_SOCKET` in their environ) |
| `start.go` | `StartDetached` with a `Placement` (systemd scope, cgroupfs hosts cgroup, or the controller's own) and the fallback when the placement cannot be used |
| `cgroup.go` | `Setup`/`SetupAt` (`Delegation`), `TuneAt`, `MoveHosts`, `For` (pool limits to an acphost cgroup), `ReadStats` |
| `acl.go` | POSIX ACL encoding, `RestrictDir`, `RestrictSession`, `GrantSearch`, `ShareDir`, `ApplySlot` |
| `mounts.go` | `CheckMounts` (can the mapped uid use every mount source and the cwd), `Allowed`, `CopyCredential` |
| `shellprobe.go` | `DefaultShellProbe`, `ProbeShell` |

## What stays in the controller

- `hostproc.go`: `stopHost` (chooses the in-process `Options.Hosts` stop
  first, then `sessionrt`) and `startDetachedHost` (reads the config for
  the placement and passes the host env and logger).
- `isolation.go`: reads pool config to build `spec.User` and `spec.Cgroup`
  (`applyIsolation`, `poolCgroup`), the env-hash key, the cgroup metrics,
  and the `sync.Once` that sets up the delegated cgroup on first need
  (`Controller.cgroup` is a `sessionrt.Delegation`).
- `mappedhome.go`, `sessionuid.go`: slot choice under `slotMu`, the
  per-session home lifecycle and `checkMappedAccess`, which need the
  config, the store and the root dir.
- `sandbox.go`, `isolatedcmd.go`: build the sandbox from live config.
- `session.go`: the `runner` (ACP connection, turns, restart, durable
  transitions). It reaches `sessionrt` through `stopHost`,
  `startDetachedHost`, and direct calls to `HostAlive`, `HostPid`,
  `ShareDir`, `RestrictSession` and `ProbeShell` in `spawnHost`.

## Synchronization

`sessionrt` starts no goroutine that outlives its call, except the flock
waiter in `waitHostExit`, which owns its file and ends when the
supervisor releases the lock (or is killed), so a stop never leaves it
reading a closed fd (acphost#227). It holds no package state except the
read-only `DefaultShellProbe` list and its cgroup controller names. The
controller keeps `cgroupOnce`,
`hostsMoved`, `slotMu` and the runner map, and joins runners through its
`WaitGroup`, as before.

## Left for the other slices

- JOB-152: sandbox spec and isolated-command assembly.
- JOB-153: mapped home and slot allocation.
- JOB-154: the runner's live connection and ACP traffic. This one needs a
  coupling analysis first: the runner mixes connection mechanics with
  ticket claims, prompts and durable transitions.
