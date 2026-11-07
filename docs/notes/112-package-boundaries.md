# Automation package boundaries

2026-11-07. JOB-112.

| Package | Owns |
| --- | --- |
| `config` | Service settings, pack composition, instance naming and configuration validation |
| `packcue` | CUE evaluation, offline imports, diagnostics and source dependencies |
| `automation` | Starlark loading, invocation and builtin bindings |
| `automation/api` | Shared capability contracts with no import of the engine or implementations |
| `automation/http` | HTTP capability implementation |
| `automation/command` | Command capability implementation |
| `automation/jev` | Jev capability implementation |
| `hooks` | Durable invocations, attempts, external-call evidence, recovery and public audit |
| `control` | Controller admission, scheduling, sandbox integration and atomic ticket changes |

Capability implementations depend on the shared contract, not the Starlark engine. The engine composes the implementations. Hook orchestration calls a narrow controller host interface rather than importing `control`. Configuration does not import hook orchestration.

These boundaries change code ownership, not configuration syntax, stored records or callback behavior. Existing event identity, manifest fencing, ordered observations, redaction and ambiguous-outcome recovery remain required across the package boundary.
