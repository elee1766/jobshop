# selfdev pack

Jobshop agents that develop Jobshop itself. This pack is
`packs/github` (flows, pools, triggers and prompts: see its README) with
Jobshop's values, composed through CUE in `pack.cue`:

| github var | selfdev value |
| --- | --- |
| `repo` | `elee1766/jobshop` |
| `checkout` | this pack's `repo` var (the host checkout) |
| `human` | `elee1766` |
| `check_cmd` | `go build ./... && go vet ./... && go test ./... && golangci-lint run ./... && nix fmt -- --check $(git ls-files '*.nix')` |
| `docs_rule` | notes in `docs/notes/<issue>-<slug>.md`, checks in `docs/verify/<slug>.md` (#131) |
| `ci_workflow` | `ci` |
| `guide` | Go module, read AGENTS.md and docs/ (docs/plan.md) |
| `deploy_check` | `jobshop version --server` (the controller's commit, #180) at the merge commit or a descendant (self-deploy) |

## Vars

| Var | Meaning |
| --- | --- |
| `repo` | Main checkout of github.com/elee1766/jobshop; sessions get worktrees of it |
| `agent` | Agent name defined in the host's jobshop.yaml |
| `dev_sessions` | `max_sessions` for the dev pool. Default 2 |
| `learn_sessions` | `max_sessions` for the curator's `learn` pool. Default 1 |
| `curate_min_learnings` | Open learnings since the last pass that start a `curate` run. Default 10 |
| `curate_every` | Start a pass after this long when at least one learning is open. Default 24h |
| `curate_scope` | Learning scope globs the curate trigger counts (#524). Default this repo and its pools, so other instances' learnings do not start it |

## Native CUE and Starlark

`pack.cue`, `workflow.star` and `triggers.star` define this pack. The obsolete
selfdev `pack.yaml` was removed after verified native cutover. Do not select
this directory through YAML `packs:` or Nix `livePacks`; those load `pack.yaml`.
Use a CUE root selected by host YAML's `pack:` field. Include the host-level
[operator pack](../operator/README.md) separately, once per host.

Within this checkout's CUE module, a root CUE file can compose the pack:

```cue
package city

import selfdevpack "github.com/elee1766/jobshop/packs/selfdev"

packs: selfdev: selfdevpack.#Pack & {
	path: "packs/selfdev"
	prefix: ""
	#Config: {
		repo: "/var/lib/jobshop/src/jobshop"
		agent: "jcode"
		dev_sessions: 2
		starlark_workflows: true
		starlark_triggers: false
	}
}
```

The host YAML selects that root file with `pack: <file>.cue`. Host
configuration still owns agents, projects, credentials and automation
capability grants. CUE imports are local and offline, so keep the root file
in the module tree containing
`cue.mod`, `packs/selfdev` and `packs/github`. `path` is the asset directory
relative to the including root CUE file, not the imported package. Use an
absolute asset path when the root file lives elsewhere in the module.

`#Config` accepts the legacy selfdev variable names and defaults, including
the checkout path named `repo`. The exported runtime `vars.repo` is instead
`elee1766/jobshop`, as GitHub workflow templates require, and `vars.checkout`
is the checkout path. This separates the two meanings that YAML `extends`
kept in different layers. Pool, workflow, trigger and prompt-rule names stay
unchanged with `prefix: ""`. A nonempty prefix renames local references,
including the default development pool. External agent and chat-pool names
stay external.

The native pack imports GitHub policy and its original assets. Its curator
and evaluator use Starlark agent-step callbacks with the same titles,
prompts and retry limits. Their original maintenance run triggers remain
the default. The controller still checks learning counts, scopes, elapsed
time and edit eligibility through `when.learnings` and `when.edits`. The
run keys remain `curate` and `evaluate`. Prompt files and their runtime
templates are not copied or rewritten.

`#Config.starlark_workflows` defaults to `true`. Set it to `false` to keep
the CUE composition but use the original static agent steps in both
selfdev and GitHub. This provides a static parity check before selecting
scripted workflows. Names, prompts and retry limits stay the same in
either mode.

`#Config.starlark_triggers` defaults to `false`, independently of workflow
mode. This preserves the original maintenance trigger behavior even with
scripted workflows. Set it to `true` to opt into timer handlers that only
request the admitted run, with `max_open: 1` retaining one open run per
trigger. Learning and edit eligibility remain controller-owned in either
mode. Leave both settings false for the complete CUE-static comparison.

`live-check` follows `starlark_workflows` too. Scripted, its callback
starts a script step (JOB-124, `start_step(script=...)`) on the `dev`
pool with a worktree, a 20-minute timeout and zero retries. The step
function runs `jobshop runbook run` and `gh issue comment` through
`shell()`, which runs in the `dev` sandbox and worktree, and fails the
step when a readonly block failed. The host must grant
`automation: {shell: true}`. Static, it keeps the original command step.

Check a candidate with `jobshop config check --file <host.yaml>`. That loads
the actual CUE and original prompt files without starting a controller.
Callback acceptance must also load and invoke the actual `.star` files with
the runtime's run, step and timer envelopes and no HTTP or shell grants.
Drain legacy runs before selecting scripted workflows, so existing static
runs are not resumed against a different workflow shape.
