# CUE packs and Starlark hooks

2026-11-07. JOB-112.

The service/root split from JOB-122 remains: `jobshop.yaml` selects the
root work pack and retains host settings. `pack: pack.cue` opts into native
CUE package loading. Imported packages and same-package files compose
concrete work definitions. Named instances have explicit asset roots,
typed vars and prefixes. The loader rejects remote dependency resolution,
unknown exported fields and non-concrete values before activation.

Starlark modules define bounded callbacks rather than suspended workflows.
Go-backed action constructors return data for the controller. Immediate
HTTP, shell and Jev calls require service-owned grants. HTTP credential
handles are origin-bound. Shell runs as the controller user, not in an
agent-step sandbox. Jev selects only from supplied options and never
provides authorization. External effects are not exactly-once.

Reloads stage a complete valid configuration and retain the previous
configuration on failure. Existing scheduling, deduplication, retries,
review, admission, approval and exact-head evidence remain controller
policy. A format migration must preserve deployed names, subject keys and
in-flight runs rather than starting replacement work.

The [configuration guide](../configuration/README.md) contains the schema,
API reference, complete executable example, diagnostics and migration with
rollout/rollback steps. The [runbook](../verify/cue-starlark.md) separates
local automated checks from operator-only acceptance. YAML pack migration
is explicit. This change does not deploy or rewrite an existing host.
