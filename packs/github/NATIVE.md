# GitHub native pack

2026-11-07. Additive migration candidate. `pack.yaml` and all prompt files
remain unchanged. Do not select this candidate on a live controller before
comparing its resolved configuration and draining legacy workflow runs.

Import the package from a checked-in CUE module:

```cue
package city

import githubpack "github.com/elee1766/jobshop/packs/github"

packs: github: githubpack.#Pack & {
    path: "packs/github"
    prefix: ""
    vars: {
        repo: "acme/widget"
        checkout: "/srv/widget"
        agent: "jcode"
        human: "operator"
    }
}
```

The root supplies projects and agents. The service remains YAML and selects
the root CUE file with `pack: pack.cue`. The reusable package supplies the
existing six pools, seven workflows, five triggers and five prompt rules.
Counts and booleans remain typed. `prefix` defaults to the legacy empty
prefix, and `dev_pool` defaults to that instance's own `dev` pool.

`vars.asset_prefix` defaults to `""`. A composing selfdev instance whose
asset root is `packs/selfdev` supplies `"../github/"`, so pool prompts,
escalation prompts, rules and callback entries still load the original
GitHub files. Callback prompt files resolve from the GitHub module's own
directory, not the composing instance directory.

## Exact static parity and callback boundary

`vars.starlark_workflows` defaults to `false`. This selects declarative
workflow graphs expressed directly in CUE, not YAML inclusion. Their step
IDs, dependency edges, outcome conditions, retries, retry pools, titles and
prompt files match the YAML pack. It also preserves batch-start's initial
step graph and recorded `skipped` tickets.

Setting `vars.starlark_workflows: true` selects `workflows.star` for all
seven agent workflows. Starlark branches on recorded outcomes and creates
only selected downstream steps. Retry accounting, review findings,
strong-model fallback and escalation remain controller-owned. Prompts are
loaded from the existing files and rendered with resolved run inputs.
An existing reviewer is reused after implementation is reopened.

Callback workflows do not create downstream tickets until their preceding
step completes. A non-PR implementation and a held PR finish without
creating skipped downstream tickets. This is not initial-graph parity:
legacy batch-start and callers that link to a future `post-merge` ticket
must keep declarative workflows until the controller supports that graph.
Do not switch an open static run to callbacks by changing its workflow
name in place. Drain it or preserve its exact legacy definition separately.

All five triggers remain declarative CUE. Native handlers cannot construct
the replan task or consume controller-computed stall subjects with the same
per-subject history caps. Handler cooldown/dedup semantics also differ from
legacy closed-ticket cooldown and check-output history. Keeping these
triggers declarative preserves their exact source filters, fallback polls,
checks, keys, cooldowns and `max_runs` policy without adding host shell or
HTTP grants. This candidate therefore does not claim a full Starlark
trigger migration.

## Offline acceptance

```sh
go test ./packs/github
```

The retained tests load the actual pack and prompt files through the
production CUE/YAML loaders, compare all static policy, then drive callback
workflows through a disposable controller. Automatic sessions and triggers
are disabled in the controller fixture. No GitHub, Linear, model or other
external service is called.
