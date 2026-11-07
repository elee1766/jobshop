# Root work-definition pack

2026-11-07. JOB-122.

`jobshop.yaml` selects one root work-definition file with `pack: pack.yaml`.
The path is relative to the service file, following a checkout symlink as
existing pack paths do. Runtime `include:` files still resolve beside the
service link. An explicit empty `pack:` is invalid.

The root pack owns `default_project`, `projects`, `agents`, `resources`,
`pools`, `workflows`, `triggers`, `prompt_rules`, `hooks`, `escalation`,
`learnings`, `desk_commands`, `sandbox_profiles`, `isolation_profiles` and
`meta_packs`. Its `packs:` list loads existing reusable packs with their vars,
prefixes and `extends:` chains. These children keep their existing schema.
Projects and resources in the root pack are shared across child instances.

Service settings stay in `jobshop.yaml`: storage, listeners, identities and
credentials, telemetry, reload policy, limits and runtime security defaults.
A root pack cannot introduce host-only settings. Root `defaults:` means pack
template variables, not service `defaults:` sandbox and isolation policy.

Child instances merge in list order, then the root pack's own definitions.
Legacy service definitions and `packs:` remain accepted as final overrides
for migration, including Nix module includes. Mixing work definitions into a
service that selects a root pack emits a migration warning. No existing host
is rewritten automatically. `jobshop init` creates both files without
replacing an existing file.

Relative child and prompt paths resolve beside the root pack. Provenance and
instance labels survive composition. Config check path rewriting applies to
the root pack as well as its children. Reload watches the root file and its
referenced text files, not the entire root checkout. Child pack directories
remain recursively watched. A reload failure keeps the running config, and a
pack-only startup failure may use the saved resolved last-good config.

This changes the loading boundary, not the execution model or pack language.
A future pack authoring format can replace the work-definition loader without
moving service configuration out of YAML.
