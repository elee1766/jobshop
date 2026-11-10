# selfdev pack

Jobshop agents that develop Jobshop itself. This pack is
`packs/github` (flows, pools, triggers and prompts: see its README) with
Jobshop's values, through `pack.cue`:

| github var | selfdev value |
| --- | --- |
| `repo` | `gfx-labs/jobshop` |
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
| `repo` | Main checkout of github.com/gfx-labs/jobshop; sessions get worktrees of it |
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
    vars: {jobshop_repo: gfx-labs/jobshop, checkout: /var/lib/jobshop/src/jobshop, agent: jcode}
agents:
  # jcode's own memory is off: learnings come from jobshop (#320).
  jcode: {command: jcode, args: [acp], env: {JCODE_MEMORY_ENABLED: "0"}}
```

selfdev no longer brings the `operator` pool; see
`packs/operator/README.md`.

## Composition

The pack is CUE. A root CUE file in the module tree composes it:

```cue
package city

import selfdevpack "github.com/gfx-labs/jobshop/packs/selfdev"

packs: selfdev: selfdevpack.#Pack & {
	path: "packs/selfdev"
	#Config: {repo: "/var/lib/jobshop/src/jobshop", agent: "jcode", dev_sessions: 2}
}
```

The host YAML selects that root file with `pack: <file>.cue`, and still
owns agents, projects, identities and automation grants. The pack needs
`automation: {github: true, shell: true}`. CUE imports are local and
offline, so the root file lives in the module tree containing `cue.mod`,
`packs/selfdev` and `packs/github`. `path` is the asset directory
relative to the including root file.

`#Config.repo` is the checkout path; the GitHub pack's `vars.repo` is
`gfx-labs/jobshop` and `vars.checkout` the checkout. With `prefix: ""`
names are unchanged; a nonempty prefix renames local references.

`curate`, `evaluate` and `live-check` are Starlark workflows
(`workflow.star`). The `curate` and `evaluate` triggers call
`triggers.star:start` once the controller's `when.learnings` and
`when.edits` checks hold, with `max_open: 1`; the run keys are `curate`
and `evaluate`. `live-check` starts a script step on the `dev` pool with a
worktree, a 20-minute timeout and no retries: it runs `jobshop runbook
run` through `shell()` and posts the report on the issue through the
`github` built-in, and fails when a readonly block failed.

Check a config with `jobshop config check --file <host.yaml>`, which loads
the CUE, prompts and Starlark files without a controller.
