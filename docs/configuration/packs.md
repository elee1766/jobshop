# CUE pack reference

2026-11-07. JOB-112. See [configuration](README.md) and
[migration](migration.md).

## Service entry point

The service file stays YAML. Select CUE explicitly:

```yaml
# jobshop.yaml
pack: pack.cue
reload:
  watch: true
  debounce: 2s
```

`pack:` is relative to the service file. Do not copy work definitions back
into the service file as a way to compose CUE instances. Keep storage,
identities, listener configuration, telemetry, secret providers, runtime
sandbox defaults, host limits, `automation`, and `jev` in service YAML.
Root packs reject host-only fields.

`jobshop init` and the existing shipped YAML packs remain useful migration
inputs. Selecting CUE does not convert those files. Use a binary containing
the CUE loader for offline checks as well as for the controller.

## Runnable example

The checked-in
[example module](../../pkg/config/testdata/cue-starlark-example) contains:

```text
jobshop.yaml
pack.cue
cue.mod/module.cue
packs/demo/pack.cue
packs/demo/prompt.md
packs/demo/implement.md
packs/demo/workflow.star
packs/demo/events.star
packs/demo/capabilities.star
event.json
```

From the repository checkout, validate without starting the controller:

```sh
go run ./cmd/jobshop config check --file pkg/config/testdata/cue-starlark-example/jobshop.yaml
```

The root package imports a local package and supplies concrete instance
variables:

```cue
package city

import demopack "jobshop.example/automation/packs/demo"

default_project: "js"
projects: js: {}
agents: demo: {command: "jcode", args: ["acp"]}

packs: demo: demopack.#Pack & {
    vars: {
        repo: "acme/widget"
        enabled: true
    }
}
```

The reusable package constrains input values and supplies defaults:

```cue
package demo

#Pack: {
    let V = vars
    path: "packs/demo"
    vars: {
        repo: string & =~"^[^/]+/[^/]+$"
        branch: *"main" | string
        enabled: *false | bool
    }
    pools: dev: {
        agent: "demo"
        dir: "."
        max_sessions: 0
        prompt_file: "prompt.md"
    }
    workflows: change: {
        enabled: V.enabled
        vars: ["task"]
        steps: [{
            id: "implement"
            pool: "dev"
            title: "Implement: {{.task}}"
            prompt_file: "implement.md"
            retries: 1
        }]
    }
}
```

The resolved names are `demo-dev` and `demo-change`. The pool's
`max_sessions: 0` prevents automatic agent execution in the example. To
exercise a run, copy the module to a disposable root, configure an installed
ACP agent and a real checkout, and deliberately raise the pool capacity.
Never test a migration by pointing the example at a production store.

## Packages, imports and reproducibility

The named file selects a CUE package. Loading evaluates that package,
including same-package sibling files and its imports, not only the text in
`pack.cue`. Same-package files contribute constraints through unification.
Imports are CUE packages, not YAML includes or runtime downloads.

The example's `cue.mod/module.cue` is:

```cue
module: "jobshop.example/automation@v0"
language: version: "v0.17.0"
```

Commit module metadata, imported local source and local replacement
configuration with the pack revision. The loader is offline: external
module fetches and version discovery fail even if a developer's registry
cache or environment could otherwise satisfy them. Use checked-in
same-module packages or CUE local module replacements. A version pin alone
does not make a remote package available offline. Test with the same
Jobshop build that will run the controller. Hook invocation never resolves
or downloads CUE dependencies.

For a separately named checked-in module, the main module can pin a
dependency in `cue.mod/module.cue`:

```cue
module: "jobshop.example/replaced@v0"
language: version: "v0.17.0"
deps: "packs.example/demo@v0": {
    v: "v0.1.0"
    default: true
}
```

Put its directory replacement in `cue.mod/local-module.cue`, not in
`module.cue`:

```cue
deps: "packs.example/demo@v0": replaceWith: "./vendor/demo"
```

The directory is relative to the main module root. Check it in, including
its own `vendor/demo/cue.mod/module.cue` declaring the matching identity:

```cue
module: "packs.example/demo@v0"
language: version: "v0.17.0"
```

Then `import demopack "packs.example/demo"` loads that local source.
Its exported instance still needs an asset `path`, for example
`path: "vendor/demo"`, relative to the root pack directory. A replacement
to another remote module would still require an external fetch and is not
an offline-vendoring strategy. Check the entire candidate with the
Jobshop loader rather than relying on a standalone CUE cache.

CUE labels introduce scope. Give imports an alias distinct from an
instance label, such as `demopack` versus `packs.demo`. Bind instance vars
with `let V = vars` outside a workflow if that workflow also has its own
`vars` field. Otherwise a reference can resolve to the nearer label,
producing an unused-import or wrong-type error.

All public values must resolve concretely before activation. An unset
`repo: string`, unresolved disjunction, contradictory constraint or wrong
type fails loading. CUE definitions such as `#Pack` and `_hidden` helpers
are not exported as Jobshop configuration. Put reusable schemas and helper
values there, rather than adding unknown public fields. After CUE export,
Jobshop strictly decodes its existing Go configuration model and checks
cross-references, durations and policy constraints.

CUE interpolation is static, for example `ref: "origin/\(vars.branch)"`.
CUE strings are not passed through legacy pack Go-template substitution.
Runtime run prompts and titles still use their existing Go templates,
such as `{{.task}}`. Do not put an event field or untrusted executable
fragment into generated `.star` source.

## Exported shape

The root package exports the existing root work-definition fields:

| Field | Meaning |
| --- | --- |
| `default_project`, `projects` | Shared project definitions and default selection |
| `agents`, `resources` | Shared agents and resource-slot definitions |
| `pools` | Pool definitions with existing agent, checkout, capacity and profile fields |
| `workflows`, `triggers` | Existing declarative forms or Starlark callback wiring |
| `prompt_rules`, `hooks` | Existing prompt injection rules and guardrail hooks, not Starlark module imports |
| `escalation`, `learnings` | Existing escalation defaults and learning-loop definitions |
| `desk_commands`, `sandbox_profiles`, `isolation_profiles`, `meta_packs` | Existing shared work-definition extensions |
| `packs` | Mapping of names to resolved reusable instances |

Instance values export `path`, optional `prefix`, optional `vars`, and
reusable work-definition fields (`agents`, `pools`, `workflows`, `triggers`,
`prompt_rules`, `hooks`, `escalation`, `learnings`, `desk_commands`,
`sandbox_profiles`, `isolation_profiles`). Projects and resources belong
in the shared root. Unknown exported keys are errors, not ignored data.
This is the existing model, not a second independent scheduler schema.
See [legacy field examples](../packs.md#the-parts) and
[callback wiring](callbacks.md) and [built-ins](automation.md) for their
respective forms.

### Instances and references

- `packs` is a map in CUE, not the YAML `packs:` list of filesystem
  references. Each instance needs an explicit nonempty `path` asset base.
- An omitted `prefix` defaults to the instance key plus `-`. Preserve a
  deployed prefix explicitly during migration. `prefix: ""` deliberately
  exposes unprefixed names and can collide with shared definitions.
- Instance keys are processed in sorted order, then shared root
  definitions override child contributions. Do not depend on CUE source
  file order as a merge order.
- Local named references follow existing prefixing rules. Shared root
  agents, pools and resources are explicit shared references. The
  controller rejects routing a workflow to another instance's private
  pool. A script's dynamic references are validated when invoked.
- `vars` are resolved structured values, not text-template substitutions.
  Keep booleans and counts typed. Do not quote a boolean merely because
  the YAML template previously returned a string.
- Disabled entries remain validated before removal from the active model.
  `enabled: false` is not an escape from invalid references or missing
  required values. Preserve deliberate flags, including disabled Canoe.

Define ownership for shared pools. If two features use one pool, put it in
the shared root or enable it whenever either feature requires it. Turning
off one feature must not remove another feature's dependencies. Preserve
pack escalation defaults and the distinction between local and shared
operator pools. Prompt rules retain their existing instance-scoped pool
coverage and text, file, command and memory sources.

## Files, provenance and reload

A root-level relative prompt path resolves beside the root package. Each
instance's `path` is relative to the root package directory, not implicitly
to the imported package's source file. Prompt and hook assets then resolve
from that instance asset base. Explicit paths avoid changing meaning when
an imported definition moves between source files.

Keep path bases stable when moving a pack. Absolute checkout, secret and
host paths are not CUE module imports. `config check --replace old=new`
can rewrite deployment checkout paths for an offline check. It does not
prove that a remote host has the same credentials or executable files.

The resolved config records contributing CUE source files and text-file
provenance. With `reload.watch: true`, the watcher follows loaded root and
imported CUE files, CUE module metadata, direct `.cue` additions in those
package directories, referenced prompt assets and loaded Starlark modules.
It does not recursively watch the whole CUE checkout. Adding an import on
a valid reload expands the watch set.

Statically declared prompt paths are discovered during loading. A
`start_step(prompt_file=...)` path computed by Starlark is validated and
read when the action is planned, not statically inferred from every script
branch. Such an asset is not automatically a watched file solely because
its name appears in a callback. The accepted step ticket retains its prompt
body. Keep dynamic assets versioned with the callback source.

A reload is staged and validated before it replaces the active config.
An invalid CUE value, missing prompt or broken hook keeps the last valid
configuration and raises the existing config-invalid diagnostics. A
pack-only startup error may use the saved last-good resolved config under
the existing startup policy. Do not treat fallback as successful migration.
Fix the source and confirm both the active revision and error clearance.

A saved resolved config is not an archive of executable `.star` files or
prompt assets read dynamically by actions. Statically resolved prompt text
may be embedded, but that is not a complete pack checkout. Keep the
pack/module/assets revision available at its expected paths for rollback
and pinned runs. Last-good configuration cannot by itself restore deleted
or edited source files.

## Troubleshooting a pack

| Symptom | Check |
| --- | --- |
| Non-concrete value | Supply required vars, choose one disjunction branch, or mark a real default with `*` |
| Conflicting values | Read every CUE source position in the error. Unification constrains rather than overwrites |
| Unknown exported field | Move helper data to `#Definition` or `_hidden`, or correct the Jobshop field name |
| External module unavailable offline | Check in the source and configure a supported local import/replacement |
| Missing prompt or hook | Inspect the instance `path`, root package directory, filename case and module boundary |
| Wrong pool/workflow name | Compare the old prefix, new default prefix and local versus shared reference |
| An edit does not reload | Confirm `reload.watch`, loaded import membership and watched file paths, then use an explicit `jobshop reload` |
| Config check passes but a hook fails | Static loading cannot evaluate every branch. See callback diagnostics in the [automation reference](automation.md) |
