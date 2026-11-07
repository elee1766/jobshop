# Configuration and automation

2026-11-07. JOB-112.

Jobshop uses three layers. They have different trust and lifecycle rules:

| Layer | Owns | Does not own |
| --- | --- | --- |
| `jobshop.yaml` | Storage, identities and secrets, service endpoints, reload policy, host limits, `automation` grants and `jev` | Reusable workflow behavior |
| `pack.cue` and imported CUE packages | Projects, agents, pools, resources, prompts, workflows, trigger registration, instance variables and defaults | Host permissions or secret values |
| `.star` modules | Bounded trigger and workflow callbacks, filtering, routing, data transformations | A scheduler, suspended durable execution, agent sandbox policy or approval authority |

The [JOB-122 split](../notes/JOB-122-root-pack.md) introduced the service's
`pack:` selector. JOB-112 adds native CUE composition and Starlark callbacks
at that boundary. A hook returns actions and exits. The controller owns
persistence, scheduling, deduplication, retries, escalation and applying
those actions. There is no CEL layer and no embedded expression pipeline.

## Read in order

1. [Pack reference](packs.md): complete example, CUE imports, schema,
   defaults, constraints, names, files and reloads.
2. [Callback reference](callbacks.md): wiring, event/run/step envelopes,
   lifecycle and tested workflow/trigger examples.
3. [Built-ins and actions](automation.md): capability contracts, credentials,
   budgets and diagnostic tracebacks.
4. [Jobcity parity inventory](jobcity.md): concrete review, delegated
   intake, routing, timer and learnings examples from a dated source snapshot.
5. [Migration guide](migration.md): YAML-to-CUE changes, parity checks,
   rollout, in-flight runs and rollback.
6. [Verification runbook](../verify/cue-starlark.md): local checks and
   separately identified operator checks.

The executable example is
[`pkg/config/testdata/cue-starlark-example`](../../pkg/config/testdata/cue-starlark-example).
It is a local CUE module, needs no CUE registry or external credentials,
and disables automatic agent sessions. `config check` is safe to run
without a controller. Starting its workflow is an explicit write and
requires a disposable root and an installed ACP agent.

Existing service operations, identities, and sandbox policy are covered
by [running a root](../running.md), [ingest](../ingest.md), and
[isolation](../isolation.md). The [legacy pack guide](../packs.md) describes
YAML packs during migration. CUE is the destination authoring format,
not a way to interpolate YAML templates before running them.
