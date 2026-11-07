# selfdev pack

Jobshop agents that develop Jobshop itself. This pack is
`packs/github` (flows, pools, triggers and prompts: see its README) with
Jobshop's values, through `extends:` in `pack.yaml`:

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

```yaml
packs:
  - path: /var/lib/jobshop/src/jobshop/packs/selfdev
    vars: {repo: /var/lib/jobshop/src/jobshop, agent: jcode, dev_sessions: 2}
  # The operator chat pool is host-level, included once (#248).
  - path: /var/lib/jobshop/src/jobshop/packs/operator
    vars: {jobshop_repo: elee1766/jobshop, checkout: /var/lib/jobshop/src/jobshop, agent: jcode}
agents:
  # jcode's own memory is off: learnings come from jobshop (#320).
  jcode: {command: jcode, args: [acp], env: {JCODE_MEMORY_ENABLED: "0"}}
```

`jobshop config show` lists both `packs/github/pack.yaml` and
`packs/selfdev/pack.yaml` as sources of each entry. selfdev no longer
brings the `operator` pool; see `packs/operator/README.md`.

## Native CUE and Starlark

`pack.cue`, `workflow.star` and `triggers.star` are the native version beside
the unchanged `pack.yaml`. Adding these files does not select them for an
existing host. Validate the native configuration and callbacks first, switch
the host separately, then remove the legacy YAML only after cutover checks.
No host configuration is changed by this pack.

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
prompts and retry limits. Their trigger handlers only request the admitted
run. The controller still checks learning counts, scopes, elapsed time and
edit eligibility through `when.learnings` and `when.edits`. `max_open: 1`
retains one open run per trigger, and the run keys remain `curate` and
`evaluate`. Prompt files and their runtime templates are not copied or
rewritten.

`live-check` remains a static CUE command workflow with its original command,
`dev` worktree, 20-minute timeout and zero retries. The current Starlark
`start_step` action supports agent steps only, not command type, command
text, worktree or timeout fields. Running the command through a callback's
shell capability would lose the durable command-step lifecycle and is not
an equivalent migration.

Check a candidate with `jobshop config check --file <host.yaml>`. That loads
the actual CUE and original prompt files without starting a controller.
Callback acceptance must also load and invoke the actual `.star` files with
the runtime's run, step and timer envelopes and no HTTP or shell grants.
Drain legacy runs before selecting scripted workflows, so existing static
runs are not resumed against a different workflow shape.

Scripted runs also pin their accepted callback and configuration manifest.
Changes to a workflow's vars, instance metadata or pool/workflow project
maps can change that manifest while a run is open. The controller then
requires the pinned configuration to be restored before resuming, rather
than continuing under changed policy. Drain native runs before reloads
that alter their accepted manifest. This differs from legacy static runs
and must be included in cutover acceptance.
