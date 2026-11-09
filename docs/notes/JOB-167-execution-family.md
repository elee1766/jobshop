# JOB-167: execution family (2026-11-09)

Child of JOB-164. Moves packages, no behavior change. Persisted `spec.json`
layout, cgroup, ACL and sandbox behavior are the same code.

## Where things went

| Was | Now | Owns |
|---|---|---|
| `pkg/control/sessionrt` (session half) | `pkg/execution/session` | host liveness, stop, detached start, background job scan, shell probe, mapped homes, slot allocation, host connect, ACP client and traffic, error classes |
| `pkg/hostspec` | `pkg/execution/session` (`spec.go`) | `spec.json` read and write on acphost. `session.Spec`, `session.Read`, `session.Write`, `session.HostOptions` |
| `pkg/control/sessionrt` (isolation half) | `pkg/execution/sandbox` | sandbox spec from config, mapped user, POSIX ACLs, mount checks, delegated cgroups and pool limits, `IsolateSpec` and `Run` (one isolated command), `GitCommonDir`, `ExpandHome` |
| `pkg/automation/...` | `pkg/execution/automation/...` | Starlark runtime, `api` leaf and the bounded capabilities, unchanged |

The split follows what a caller needs, not file names. Hook, prompt rule,
Desk exec, trigger check and command step all run through
`sandbox.IsolateSpec` and `sandbox.Run` (the controller's `runIsolated`).
None of that reaches `session`, `acphost.acphost` or the ACP SDK.

## Changes beyond the move

- `sandbox.ApplySlot` and `RestrictSession` take the `*userns.User` and the
  worktree dir instead of an `acphost.Spec`, so `sandbox` does not need the
  session spec type. Callers pass `spec.User` and `spec.Dir`.
- `sandbox` returns `xerr.PermanentError` directly (the type `acphost`
  aliases), so `acphost.Permanent` still sees it and `sandbox` does not
  import `acphost.acphost`.
- `SandboxKey` is now `sandbox.Key` (revive stutter).
- `ExpandHome` moved with the sandbox spec code, the only mechanics that
  need it besides homes.

## Boundary check

- New family `sandbox` (may import support, config, domain). `session` may
  import `sandbox`. `orchestration` and `execution/session` rows updated.
- `command-paths-no-session` and `command-paths-no-acp` now also start from
  `pkg/execution/sandbox`, and no longer list `hostspec` (it is `session`).
- Removed the two baseline exceptions for `stepexec -> sessionrt` and the ACP
  SDK reached through it. The other baseline exceptions stay with their children.
- `hostspec` is no longer an independent leaf: it moved into `session`, which
  imports acphost and `sandbox` for the other mechanics.
- `journalread -> hostspec` was removed by JOB-173 before this landed, so no
  journal package imports `session`.

## Not done

Scheduling, placement and admission stay in control. No executor registry and
no new backend. The hook host adapter (`hookcommands.go`) is unchanged: it
still calls the controller's `runIsolated`, which JOB-168 narrows.
