# Sandbox and isolation profiles attached by name or label (JOB-83, 2026-11-06)

## Attach by name

- A host pool entry that names a pack pool merges per field over it
  (#304). `sandbox.profiles` and `isolation.profiles` are lists, so they
  are unioned (pack's first, no duplicates) instead of replaced:
  `pools.sap-dev.sandbox.profiles: [host]` adds `host` to what the pack
  pool already names. Other `sandbox` fields merge as before (lists
  replace, maps merge per key).
- A prefixed pack's own profile names and the references to them are
  renamed together, for `isolation_profiles` as for `sandbox_profiles`.

## Isolation profiles

- `isolation_profiles.<name>` takes the fields of `isolation` (`user`,
  `limits`); `isolation.profiles: [..]` on defaults, a label rule, an
  agent or a pool applies them before that level's own fields. A profile
  cannot name profiles. Packs contribute entries like pools.
- Isolation does not append: `user` and `limits` are each taken whole
  from the last level (or profile) that sets them, as before.

## Attach by label

```yaml
defaults:
  sandbox_by_label:
    - {labels: ["instance:sap"], profiles: [host, egress-deny]}
  isolation_by_label:
    - {labels: ["instance:sap"], profiles: [mapped]}
```

- A rule is the fields of `sandbox` (or `isolation`) plus `labels`, which
  is required. A pool matches when it carries every label (`k=v` matches
  `k:v`, as in `pool list --label`). Pack instance labels count: the rules
  are checked after `instance:<name>` and `pack:<pack>` are added.
- Merge order, lowest first: defaults, matching rules in list order,
  agent, pool. Each level's profiles come before its own fields. Lists
  append in that order, scalars take the last level that sets one.
- Decision recorded on #569.

## packs/agent

The `sandbox` and `isolation` vars are gone. The pack declares them under
`removed_vars:` (a new pack.yaml key: name to replacement text), so a host
that still sets one fails to load with `var "sandbox" is removed:
attach sandbox_profiles to the pools ...`. Unlike `renamed_vars` there is
no warning period: the var carried an object that went into three pools,
and no rendering of it remains. No other pack had such vars.

## Inspection

- `jobshop pool show <pool> --sandbox` (admin or `pool-list` grant, RPC
  `pool.show`) prints the merged sandbox and isolation, one line per
  value with its source: `defaults`, `label rule N [labels]`, `agent
  <name>`, `pool`, or `profile <name> (<level>)`. A scalar shows the level
  that won. Without `--sandbox` it prints the pool's agent, project and
  labels. `--json` prints the RPC result.
- `jobshop config check` fails on an unknown profile name (sandbox,
  isolation, in a rule or not), a rule with no labels, and bad fields in
  a rule or profile. A rule whose labels match no pool is a warning, since
  a label typo would otherwise drop the profile silently.
- `jobshop session sandbox <id>` is unchanged and still shows the bwrap
  argv, so the before and after of moving a pool to profiles compare
  directly. `TestIsolatedPoolAsProfiles` checks the merged result of an
  inline pool and its profile form is identical.

## Not done

- The NixOS module's per_session range assertion reads the pool's own
  `isolation` and its `isolation.profiles` in `settings`. Pools that get
  their isolation from a label rule or a pack are quarantined at the
  first spawn if the range does not fit, as before.
- Comparing `jobshop session sandbox` on gasserver before and after is a
  runbook check: `docs/verify/sandbox-profiles.md`.
