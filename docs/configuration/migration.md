# Migrate YAML packs to CUE and Starlark

2026-11-07. JOB-112. This guide changes authoring, not scheduling authority.
Do not combine it with a repository move, pool rename, new merge policy,
Jev routing policy or a production deploy. See the
[pack reference](packs.md), [callback reference](callbacks.md),
[built-ins](automation.md), and
[verification runbook](../verify/cue-starlark.md).

## Compatibility and removal boundary

JOB-122 already separated service YAML from the root work pack. Use that
split rather than creating a competing service format. Service YAML,
secret references, host security defaults and existing controller storage
remain in place. `pack: pack.cue` explicitly selects the native CUE loader.
Existing YAML root/child packs and declarative workflows/triggers remain
migration inputs. They are not converted automatically when the binary is
upgraded. Loading a legacy YAML pack emits a deprecation warning with this
migration guide's path. Mixed service/work definitions retain their
existing migration warnings. A warning does not convert the source or
disable compatibility loading. Do not infer a removal release or deadline
that is not announced.

CUE is the destination pack format. After operators migrate, removal of
the legacy YAML pack loader/template semantics must be a separately
announced compatibility change. Service YAML is not scheduled for removal.
Keep the old binary, service files and pack checkout revision available
through the rollback window. A CUE file is not readable by an older binary
just because it resolves to the same work definitions.

## 1. Capture a baseline before editing

Use the current revision, not a remembered example. Save:

- Binary version, service file and includes, root/child pack revisions,
  prompts and scripts, CUE module/replacement metadata if already used.
- The controller's resolved config, including source/provenance and
  instance summaries, enabled/disabled entries and watch state.
- Workflow names and step IDs, pool/project/resource names, instance
  prefixes, escalation pools, identity handles and repository overrides.
- Trigger source filters, schedules, source intervals, cooldowns, open
  caps, subject keys and existing run dedup keys.
- In-flight run IDs and states, step attempts, pending trigger/inbox work,
  approvals, current-head evidence and known ambiguous external calls.

Read-only controller inspection:

```sh
jobshop version
jobshop config show
jobshop run list
jobshop trigger list
jobshop attention
```

Store snapshots may include sensitive configuration references and
operator data. Protect them and take a store backup using the backend's
supported procedure before the authorized cutover. Never copy a live
SQLite file without accounting for its WAL, and do not edit SQL rows to
make a renamed workflow look like the old one.

## 2. Finish the service/root split

Before, a combined service might contain:

```yaml
# jobshop.yaml, before
reload: {watch: true}
default_project: js
projects: {js: {}}
agents:
  jcode: {command: jcode, args: [acp]}
packs:
  - path: packs/widget
    prefix: widget-
    vars: {checkout: /srv/widget, repo: acme/widget, agent: jcode}
```

After JOB-122, preserve its meaning in two files:

```yaml
# jobshop.yaml, service only
pack: pack.yaml
reload: {watch: true}
```

```yaml
# pack.yaml, root work definitions
default_project: js
projects: {js: {}}
agents:
  jcode: {command: jcode, args: [acp]}
packs:
  - path: packs/widget
    prefix: widget-
    vars: {checkout: /srv/widget, repo: acme/widget, agent: jcode}
```

Keep identities, credentials, listeners, store, telemetry, hard limits and
runtime security defaults in the service file. Move them only if they were in
the wrong file. Recheck relative paths if the root work pack moves. Check
Nix-generated includes too: a service override can silently mask an
otherwise correct new root pack.

## 3. Convert static composition first

Create a local CUE module and checked-in packages. For example:

```cue
// cue.mod/module.cue
module: "jobcity.example/local@v0"
language: version: "v0.17.0"
```

```cue
// pack.cue
package city

import widgetpack "jobcity.example/local/packs/widget"

default_project: "js"
projects: js: {}
agents: jcode: {command: "jcode", args: ["acp"]}
packs: widget: widgetpack.#Pack & {
    prefix: "widget-"
    vars: {
        checkout: "/srv/widget"
        repo: "acme/widget"
        agent: "jcode"
    }
}
```

Define `#Pack` in `packs/widget/pack.cue`, with an explicit
`path: "packs/widget"` and the existing work sections. Start from the
[complete tested example](../../pkg/config/testdata/cue-starlark-example),
not an early design sketch. Preserve the old steps and triggers while
changing their static container. Then select `pack: pack.cue` in a
**candidate** service file. Do not point the live service at a partially
converted package.

Translation rules:

| Legacy form | CUE destination |
| --- | --- |
| `packs:` list with `path`, `vars`, `prefix` | Imported `#Pack` definitions unified into a named `packs` map, explicit asset `path` and preserved `prefix` |
| Pack `defaults:` variables | Typed `vars` fields with marked defaults such as `branch: *"main" \| string` |
| `{{.repo}}` in static pack configuration | CUE references/interpolation, not a quoted Go template |
| Stringified boolean/count variables | Concrete boolean/integer values with CUE constraints |
| `extends:` inheritance | CUE imports and unification of reusable definitions, with conflicts resolved deliberately |
| Renamed/removed variable metadata | Explicit migration of callers and schema constraints, not ignored keys |
| Runtime prompt `{{.task}}` | Keep the existing prompt Go template and supply the same runtime inputs |
| Runtime branch/filter/discovery logic | A `.star` function, after static composition parity is established |

CUE unification does not implement last-writer-wins. A constraint such as
`max_sessions: 2` cannot be overridden with 4 unless the definition permits
it. Multiple source files contribute to one package. Use definitions or
hidden fields for helpers so they do not become unknown public config
fields. Remote CUE registry access is disabled. Pin and check in local
sources/replacements before testing on a host without a developer cache.

## 4. Compare resolved static behavior

Run the candidate binary without a controller. The `resolve` command
loads and validates the same configuration as `check`, prints resolved
JSON by default, and removes format-specific `pack`/`packs` composition
metadata:

```sh
jobshop config check --file /path/to/before/jobshop.yaml
jobshop config check --file /path/to/candidate/jobshop.yaml
jobshop config resolve --file /path/to/before/jobshop.yaml > before.json
jobshop config resolve --file /path/to/candidate/jobshop.yaml > after.json
jq -S . before.json > before.sorted.json
jq -S . after.json > after.sorted.json
diff -u before.sorted.json after.sorted.json
```

These commands write comparison files in the current directory. Use a
private scratch directory, not the live root. `--format yaml` is available
for inspection. `--replace OLD=NEW` reads pack paths from a candidate
checkout for deployment checks. Rewrite only known checkout roots and
compare the intended paths, not accidental machine-specific differences.
Do not discard all prompt, profile or path changes as noise.

First require equality for a static-format-only change. Any deliberate
remaining difference needs an explanation against the baseline: projects,
agents, pools/capacity, enabled flags, prompt rules, escalation, resources,
tracker identity, retry/review policy, source filters and timers. Source
paths are expected to change from YAML to CUE and are checked separately.
The comparison is not proof of callback behavior or external authority.

## 5. Convert behavior without changing policy

Move one behavior at a time to a callback. A callback receives an event or
recorded run/step state, performs bounded evaluation, returns actions and
exits. It does not wait for an agent or replace the controller loop. Keep
agent triage initially. Do not replace it with Jev merely because the new
runtime makes that possible.

Representative jobcity parity checklist:

| Flow | Preserve and demonstrate |
| --- | --- |
| Review-request reconciliation | All six existing GitHub webhook wakeups, their filters, targeted listing optimization, full authoritative reconciliation every 15 minutes, current delegated/review authority and the exact old subject/dedup identity |
| Delegated Linear intake | Delegated-only authoritative selection, workspace/team/project filters, source intervals, issue rereads and the same issue key. A webhook is a wakeup, not delegation evidence |
| Stuck-work handling | Each existing signal, age threshold, per-subject cap, cooldown, existing-run discovery and escalation route |
| Scheduled improvement sweep | Original cron/interval, repository selection, disabled flags, open-run cap and schedule dedup scope |
| Implementation routing | Recorded structured triage result, allowed pool scope, retries/escalation, shipping/post-merge checks, rejection/failure handling and merge authority |
| Learnings | Explicit existing wakeup/source mapping and intervals. Stuck-work signals are not a substitute |

Do not invent enriched GitHub payloads. An `issue_comment` is not a full
pull-request object. Fetch the authoritative PR and head when needed.
Preserve all repository-specific overrides, separate city/operator packs,
disabled Canoe and every deliberately enabled feature. A single webhook
example cannot establish this matrix's parity.

The [dated jobcity inventory](jobcity.md) records the concrete six webhook
registrations, timers, keys and repository differences from source. Use the
[callback examples](callbacks.md) as API examples, not as a
replacement for the full jobcity inventory. Record the six actual source
registrations from the baseline and test each. Keep targeted review reads
for targeted events and the periodic full listing for recovery. Do not
replace authenticated admission/screening with a permissive callback.
Retain authentication, delegation and injection checks, ordering, timeout
and fail-closed handling until an explicitly validated replacement exists.

### Inputs, outputs and conditions

Replace string expression pipelines with normal Starlark dictionaries,
branches, loops and helpers. Validate custom payloads at the adapter or
callback boundary and report missing fields contextually. Structured step
output must come from the recorded output contract, not scraped completion
prose. Preserve output validation and failure/escalation behavior. The
[callback reference](callbacks.md) states which contracts the shipped
integration supports. Do not deploy a schema key merely because it appears
in a proposal.

## 6. Preserve existing runs and deduplication

A format conversion must not create new logical work:

1. Preserve workflow names, prefix spelling, step IDs, run input keys,
   trigger names and subject/run dedup keys. `review:PR` and `review-PR`
   are different identities. Do not normalize or rename them casually.
2. Keep the definition and prompt revision required by in-flight runs
   available. Do not switch an old run to new callback semantics without
   proving its persisted transition model is compatible.
3. Drain or isolate old-format runs before changing a workflow's
   execution shape. New starts can use a separately reviewed versioned
   workflow only with an explicit subject-to-existing-run lookup, so the
   old run wins rather than being duplicated.
4. Do not run the old and new trigger as concurrent writers for one
   subject. Comparing candidate action data is safe only when the
   candidate does not call mutating capabilities or apply actions.
5. Re-deliver representative wakeups and reconcile after reload/restart.
   Compare original run IDs, not just counts. Verify cooldown/open caps
   and approval/exact-head guards still apply.

Do not delete run rows, clear dedup keys, reset poll cursors or rename the
store to make tests pass. If compatibility cannot be established, stop new
starts for the affected workflow, let existing work drain under the old
configuration, then cut over. That is a controlled migration, not a claim
that arbitrary old runs can be replayed through a new script.

## 7. Roll out in bounded stages

These are operator actions, not actions taken by this documentation PR.

1. Pass offline checks and parity tests with all immediate capabilities
   disabled. Review CUE source/provenance, `.star` code and host grants.
2. Use a disposable store and read-only or low-risk flow. Compare proposed
   actions with the old behavior. There is no implied production shadow
   mode: a function that calls POST or shell is not read-only because its
   returned actions are discarded.
3. Verify duplicate delivery, stale completion, invalid output, timeout,
   cancellation, broken reload and restart boundaries. Confirm denied
   capabilities remain denied and tokens cannot cross origins.
4. Freeze one approved candidate revision, snapshot the baseline/store,
   and record in-flight runs and ambiguous calls. Change only the approved
   service selector/pack revision under normal deployment authorization.
5. Confirm the active config is the candidate, not saved last-good
   fallback. Inspect errors/attention and one bounded end-to-end flow.
   Compare subject keys, original run IDs, attempts, routes and evidence.
6. Expand to the remaining flows only after each inventory row passes.
   Keep the rollback artifacts until the observation window is complete.

## 8. Handle ambiguous side effects and roll back

Action deduplication does not promise exactly-once HTTP, shell or paid
model calls. A remote mutation may succeed just before the controller
crashes or loses its response. Before retrying, reconcile the external
system using a stable operation/idempotency key. Record the observed
result and choose a safe continuation through the controller. Never retry
a payment, merge, publish or other irreversible action based only on a
missing local result.

For rollback, stop new starts for the affected scope through the existing
operator controls. Reconcile unknown external effects first. Restore the
last known service and pack revisions, then validate with the intended
binary before activation. An older binary requires its older YAML pack
selector, not the new CUE path. Preserve the store, run IDs, dedup records,
inbox/poll state, pending approvals and sessions. Resume the same work,
not new runs with replacement keys.

Do not restore a pre-cutover database over writes made since cutover as a
routine rollback. That loses accepted transitions and can repeat external
effects. A full store restore requires a separate recovery plan that
accounts for every subsequent external mutation and store write. When
new callback-created runs cannot be interpreted safely by the old binary,
keep the compatible binary running long enough to drain or resolve them
instead of forcing a downgrade.

## Acceptance record

For each migrated flow retain: baseline/candidate revisions, resolved diff,
source-event inventory, input and expected action examples, dedup keys,
existing-run IDs, output validation result, timeout/cancellation result,
reload/restart result, capability tests and operator rollback decision.
The checked-in fixture tests loader/runtime examples. They do not certify
a live jobcity migration, remote services, deployment or production state.
No deployment is part of writing or validating this guide.
