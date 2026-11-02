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
| `deploy_check` | `jobshop version` at the merge commit or a descendant (self-deploy) |

## Vars

| Var | Meaning |
| --- | --- |
| `repo` | Main checkout of github.com/elee1766/jobshop; sessions get worktrees of it |
| `agent` | Agent name defined in the host's jobshop.yaml |
| `dev_sessions` | `max_sessions` for the dev pool. Default 2 |

```yaml
packs:
  - path: /var/lib/jobshop/src/jobshop/packs/selfdev
    vars: {repo: /var/lib/jobshop/src/jobshop, agent: jcode, dev_sessions: 2}
agents:
  jcode: {command: jcode, args: [acp]}
```

`jobshop config show` lists both `packs/github/pack.yaml` and
`packs/selfdev/pack.yaml` as sources of each entry.
