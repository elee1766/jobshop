# selfdev pack

Jobshop agents that develop Jobshop itself. This pack is
`packs/github` (flows, pools, triggers and prompts: see its README) with
Jobshop's values, through `extends:` in `pack.yaml`:

| github var | selfdev value |
| --- | --- |
| `repo` | `elee1766/jobshop` |
| `checkout` | this pack's `repo` var (the host checkout) |
| `human` | `elee1766` |
| `check_cmd` | `go build ./... && go vet ./... && go test ./...` |
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
