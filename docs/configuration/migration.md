# Migrate YAML packs to CUE and Starlark

2026-11-08. From the YAML pack format (`pack.yaml`, `extends:`, Go
template vars, shell `check:` triggers, `when.stall`, static `steps:`) to
CUE packs whose workflows, triggers, hooks and desk commands run in
Starlark. See the [pack reference](packs.md), the
[callback reference](callbacks.md), the [built-ins](automation.md) and the
[verification runbook](../verify/cue-starlark.md). The self-dev instance on
gasserver was migrated this way ([cutover](../verify/selfdev-cue-migration.md)).

The YAML loader still reads `pack.yaml` and warns. The in-repo packs
(`packs/github`, `packs/selfdev`, `packs/operator`) are CUE only, so a host
that loads them by path must migrate when it upgrades.

## Renames

These names changed with no compatibility layer. Update every reference
in packs, prompts, host config, scripts and runbooks.

| Old | New |
| --- | --- |
| `when.stall` | `when.scan` |
| a stall | a finding |
| signal `step_stalled` | `step_idle` |
| run key `stall:<signal>:<subject>` | `scan:<signal>:<subject>` |
| `run.key_prefix` default `stall` | `scan` |
| `meta.trigger.stall` | `meta.trigger.finding` |
| `trigger list` field `stalls` | `findings` |
| hook `event.source == "stall"` | `"scan"` |
| github pack workflow `stall-check`, pool `stall`, `prompts/stall.md` | `triage`, `triage`, `prompts/triage.md` |
| vars `stall_model`, `stall_effort` | `triage_model`, `triage_effort` |
| vars `stall_after`, `stall_step_after`, `stall_human_after` | `scan_after`, `scan_step_after`, `scan_human_after` |
| vars `stall_pr_skip_authors`, `stall_pr_skip_labels` | `pr_review_skip_authors`, `pr_review_skip_labels` |
| operator var `repo` | `jobshop_repo` |

A run keyed `stall:...` no longer holds its subject: after the upgrade the
scan starts a second run for the same finding. Close or drain those runs
first:

```sh
jobshop run list | jq -r '.[] | select(.title | test("stall")) | .id'
jobshop run cancel <id>
```

## 1. Baseline

Save the current revision's behavior before editing: `jobshop version`,
`jobshop config resolve --file jobshop.yaml > before.json`, `jobshop run
list`, `jobshop trigger list` and the service file. Take a store backup with
the backend's procedure. Note in-flight runs: a run started from a static
`steps:` workflow does not resume against a Starlark workflow of the same
name, so drain each workflow before converting it.

## 2. Service and root pack

`jobshop.yaml` keeps service settings: store, listeners, identities,
secrets, telemetry, sandbox defaults and capability grants. Work
definitions (projects, agents, pools, packs) move to a root CUE file that
`jobshop.yaml` selects with `pack: pack.cue`:

```yaml
# jobshop.yaml
pack: pack.cue
automation: {github: true, shell: true}
identities:
  agent-app:
    kind: github-app
    app_id: 123
    private_key_file: /run/secrets/github-app-key
    installations: {acme: 456}
    repos: [acme/widget]
```

The root file lives in a CUE module. Imports are local and offline, so the
module names the jobshop checkout as a local replacement:

```cue
// cue.mod/module.cue
module: "example.local/host@v0"
language: version: "v0.17.0"
deps: "github.com/elee1766/jobshop@v0": {v: "v0.0.0", default: true}
```

```cue
// cue.mod/local-module.cue
deps: "github.com/elee1766/jobshop@v0": replaceWith: "/srv/jobshop"
```

## 3. Packs list to CUE imports

```yaml
# pack.yaml, before
default_project: js
projects: {js: {}}
agents: {jcode: {command: jcode, args: [acp]}}
packs:
  - path: /srv/jobshop/packs/github
    prefix: widget-
    vars: {repo: acme/widget, checkout: /srv/widget, agent: jcode, human: wile, stall_model: sonnet}
  - path: /srv/jobshop/packs/operator
    vars: {repo: acme/jobshop, checkout: /srv/jobshop, agent: jcode}
```

```cue
// pack.cue, after
package host

import (
	githubpack "github.com/elee1766/jobshop/packs/github"
	operatorpack "github.com/elee1766/jobshop/packs/operator"
)

default_project: "js"
projects: js: {}
agents: jcode: {command: "jcode", args: ["acp"]}

packs: widget: githubpack.#Pack & {
	path:   "/srv/jobshop/packs/github"
	prefix: "widget-"
	vars: {repo: "acme/widget", checkout: "/srv/widget", agent: "jcode", human: "wile", triage_model: "sonnet"}
}
packs: operator: operatorpack.#Pack & {
	path: "/srv/jobshop/packs/operator"
	vars: {jobshop_repo: "acme/jobshop", checkout: "/srv/jobshop", agent: "jcode"}
}
```

- `path` is the asset directory (prompts, `.star` files), relative to the
  root file or absolute.
- An entry without `prefix` gets `<name>-`. The in-repo `#Pack`s default
  `prefix` to `""`, so set it explicitly where the YAML entry had one.
- `vars` are typed: counts are integers and toggles are booleans, not
  strings. Unknown vars are an error.
- `selfdev` takes its settings as `#Config: {repo: <checkout>, agent: ...}`.

For your own packs:

| YAML | CUE |
| --- | --- |
| `defaults:` | typed `vars` with defaults, `branch: *"main" \| string` |
| `{{.repo}}` in pack fields | CUE references, `repo: V.repo` or `"\(V.repo)"` |
| `extends:` | import the base package and unify its `#Pack` |
| `removed_vars` / `renamed_vars` | change the callers; CUE rejects unknown fields |
| `{{.task}}` in prompts | unchanged: prompts keep Go templates over run inputs |

CUE unification is not last-writer-wins: a value the definition fixes
cannot be overridden. Mark overridable values as defaults (`*2 | int`).
Host-level pool overrides (`pools.dev.model`) still merge per field.

## 4. Static steps to Starlark workflows

A workflow names two callbacks instead of a `steps:` list. The controller
calls them, applies the returned actions in one transaction, and calls
`on_step_completed` when a step closes.

```yaml
# before
ci-repair:
  vars: [sha]
  steps:
    - {id: fix, pool: dev, retries: 2, title: "Fix red master CI at {{.sha}}", prompt_file: prompts/ci-fix.md}
    - {id: merge, pool: refinery, retries: 2, title: "Merge CI fix for {{.sha}}", prompt_file: prompts/ci-merge.md}
```

```cue
// after
"ci-repair": {
	vars: ["sha"]
	on_start:          "workflows.star:ci_start"
	on_step_completed: "workflows.star:ci_completed"
}
```

```python
def ci_start(run):
    return [start_step(id = "fix", pool = "dev", retries = 2, inputs = run.inputs,
                       title = "Fix red master CI at " + run.inputs["sha"],
                       prompt_file = "prompts/ci-fix.md")]

def ci_completed(run, step):
    if step.id == "fix":
        return [start_step(id = "merge", pool = "refinery", retries = 2, needs = ["fix"],
                           inputs = run.inputs, title = "Merge CI fix for " + run.inputs["sha"],
                           prompt_file = "prompts/ci-merge.md")]
    return [finish_run(outcome = step.outcome)]
```

- `when: {step: implement, outcome: pr}` becomes an `if` on `step.outcome`.
- A `command:` step becomes `start_step(script = fn, worktree = True)`;
  `fn(run, step)` calls `shell()` in the pool's sandbox and worktree.
- `retry_pool`, `needs`, `retries` and `output_schema` are `start_step`
  arguments. Step IDs and prompt files stay the same, so prompts need no
  change beyond the renames.

## 5. Triggers

Count and cron conditions (`when: {pool, ready_below}`, `when.cron`), `on:`
webhook filters, `fallback`, `cooldown`, `max_runs` and `max_open` keep
their meaning. What changes is the action and the shell checks.

**`check:` plus `run:` becomes a handler.** The handler reads the source
itself and returns the runs to start. GitHub reads go through the
`github` built-in with the identity that covers the repo, never `gh`:

```yaml
# before
master-red:
  on: [{github: {event: workflow_run, action: completed, repo: "{{.repo}}", workflow: ci, conclusion: failure}}]
  fallback: 5m
  when:
    check: gh run list --repo {{.repo}} --workflow ci --status completed --limit 1 --json conclusion,headSha --jq '...'
  run: {workflow: ci-repair, vars: {sha: "{{.check_output}}"}, key: "{{.check_output}}"}
```

```cue
// after
"master-red": {
	on: [{github: {event: "workflow_run", action: "completed", repo: V.repo, workflow: "ci", conclusion: "failure"}}]
	fallback: "5m"
	handler:  "triggers.star:master_red"
	args: {repo: V.repo, ci_workflow: "ci"}
}
```

```python
def master_red(event, args):
    runs = github.get("/repos/" + args["repo"] + "/actions/runs",
                      params = {"status": "completed", "per_page": "30"}).json()["workflow_runs"]
    runs = [r for r in runs if r["name"] == args["ci_workflow"]]
    if not runs or runs[0]["conclusion"] != "failure":
        return []
    sha = runs[0]["head_sha"]
    return [start_run(workflow = "ci-repair", key = sha, inputs = {"sha": sha})]
```

A handler with `on:` runs on each admitted delivery and every `fallback`.
The key is the dedup identity, so keep the old key spelling. `ticket:`
becomes `create_ticket(title, key, pool, ...)`.

**`when.stall` becomes `when.scan`.** Store signals are unchanged apart
from `step_idle`. A signal that ran a shell `check` names a Starlark
`script` that returns `{subject: detail}`:

```yaml
# before
when:
  stall:
    - {signal: step_stalled, after: 3h}
    - {signal: issue_no_run, repo: "{{.repo}}", check: "gh issue list ... --jq '...'"}
run: {workflow: stall-check, vars: {signal: "{{.signal}}", subject: "{{.subject}}", detail: "{{.detail}}"}}
```

```cue
// after
when: scan: [
	{signal: "step_idle", after: "3h"},
	{signal: "issue_no_run", repo: V.repo, script: "triggers.star:issue_no_run", args: {repo: V.repo}},
]
run: {workflow: "triage", vars: {signal: "{{.signal}}", subject: "{{.subject}}", detail: "{{.detail}}"}}
```

The controller still keeps per-finding history, `max_runs`, cooldowns,
`max_open` and the open-run filter for `issue_no_run`. A script runs
outside the tick at most once per `poll` (default 5m). A scan trigger may
also name a `handler` instead of `run:` to choose the workflow per finding.

## 6. Hooks and desk commands

A `command:` hook or desk command becomes `handler: file.star:fn` with
`args`. Guard hooks fail closed: `{}` allows, `{"decision": "deny",
"reason": ...}` denies, an error denies.

```yaml
# before
hooks:
  - on: action
    actions: [push]
    pools: [operator]
    command: 'case "$JOBSHOP_TARGET" in ""|master|main) exit 1;; esac'
```

```cue
// after
hooks: [{
	on: "action"
	actions: ["push"]
	pools: ["operator"]
	handler: "guard.star:push"
	args: protected: ["master", "main"]
	timeout: "5s"
}]
```

```python
def push(event, args):
    target = event.get("target", "")
    if not target:
        return {"decision": "deny", "reason": "name the branch: jobshop guard push --target <branch>"}
    if target.replace(":", "/").split("/")[-1] in args["protected"]:
        return {"decision": "deny", "reason": "push a branch and open a PR"}
    return {}
```

A desk handler `fn(input, args)` returns the dict Desk's `api.exec` gets.
`intake` hooks and `url` hooks keep their old form.

## 7. Grants

Capabilities are off by default and granted in `jobshop.yaml`:
`automation.github` for the `github` built-in (needs a `github-app` or
`github-token` identity covering the repo), `automation.linear` for
`linear.graphql` (a `linear-app` or `linear` identity covering the team),
`automation.shell` for
`shell()` in script steps and handlers, `automation.http` for `http`. A
handler shell runs as the controller with its environment; a script step
runs in the pool's sandbox. Remove `GH_TOKEN` and `gh` from the
controller's PATH once nothing shells out to it.

## 8. Check and cut over

```sh
jobshop config check --file candidate/jobshop.yaml
jobshop config resolve --file candidate/jobshop.yaml > after.json
```

`config check` compiles every `.star` file a workflow, trigger, scan
signal, hook or desk command names and checks its exports. Compare
`after.json` with `before.json`: pools, agents, prompt rules, trigger
filters, cooldowns and caps should match apart from the renames; workflows
now show callbacks instead of steps. `--replace OLD=NEW` points pack paths
at a candidate checkout.

Then:

1. Drain every open run, not only those of converted workflows, and close
   runs keyed `stall:`. A run pins the hook manifest it started under;
   after the packs change, its next callback fails with `hook manifest
   changed` and `run resume` refuses it (#1012).
2. Switch the service to the candidate config and binary together. An
   older binary cannot read the CUE root, and an automatic deploy that
   checks the old host config against the new packs fails on the renamed
   vars: change the host config first, or switch by hand.
3. `jobshop trigger list`: each trigger appears once; scan triggers list
   findings or the reason they do not hold.
4. `jobshop automation list`: handler, scan script and hook invocations
   end `done` or `sync_done`. `jobshop automation show <id>` shows a
   failure's error and its GitHub calls.
5. Walk one run of each workflow end to end.

Rollback is the previous binary with the previous YAML packs from git.
Keep the store: do not restore a database over writes made since cutover.
Runs started by Starlark workflows do not continue under the old static
definitions; let them finish before downgrading.
