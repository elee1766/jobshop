# sessionrt: sandbox spec and isolated command mechanics (JOB-152, 2026-11-08)

Slice 2 of JOB-146 (slice 1: `docs/notes/JOB-146-sessionrt.md`). JOB-146
stays open. No behavior change.

## Moved to `pkg/control/sessionrt`

| File | Owns |
|---|---|
| `sandbox.go` | `BuildSandbox` (config to `sandbox.Sandbox`: ro, rw, guidance, hide, overlays, env keep, DNS), `SandboxKey` (the env-hash part), `KeepEnv`, `ExpandHome`, `Within`, `GitCommonDir`, `DNSChecks` (the once-per-setting pasta DNS check, with its own lock) |
| `command.go` | `IsolateSpec` (sandbox, mapped user and cgroup on a `proc.Spec`, scratch dir and cleanup), `Run` (start, wait, `ErrSetup`), `UserFor`, `HasEnv` |

All of them take the config as an argument and read no controller state.
`IsolateSpec` gets the two controller-owned pieces as callbacks:
`Exe` (the jobshop binary) and `Cgroup` (`Controller.poolCgroup`, which
owns the delegated cgroup setup).

## Left in the controller

`sandbox.go` and `isolatedcmd.go` keep thin callers: `poolSandbox`,
`sandboxFor`, `sandboxKey` and `checkSandboxDNS` read `c.config()`;
`runIsolatedCmd` passes `ic.cfg` (the pass's config in pass code, JOB-104)
and the scratch dir under the root; `sessionSandbox` is the debug RPC.
`Controller.dnsChecks` is a `sessionrt.DNSChecks`, replacing the
`dnsChecked` map guarded by `c.mu`.
