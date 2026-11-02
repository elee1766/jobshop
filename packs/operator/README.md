# operator pack

A chat pool, `operator`, whose sessions talk to the operator and manage
work on one repo: create, comment and close tickets, file and comment on
GitHub issues, `jobshop send` to sessions and pools, start runs, read
ticket traces and session lists, and `jobshop reload`. When the operator
corrects behavior, the session turns it into a PR to the governing pack
prompt (merged by the refinery through a `pr-work` run) or an issue.
Replies are short; long output goes to an issue or ticket.
Prompt: `prompts/operator.md`. Design: `docs/plan.md` section 8.

Until chat pools (`type: chat`, #122) and the gateway (#123) land, the
pool starts no sessions on its own (`max_sessions: 0`). Start one by
hand and talk to it with `jobshop send` and `session attach`:

```sh
JOBSHOP_SOCKET= jobshop session new --agent jcode --pool operator --key operator:cli
JOBSHOP_SOCKET= jobshop ticket create "Operator chat" --pool operator
JOBSHOP_SOCKET= jobshop send operator:cli "file an issue for X, and tell the dev pool Y"
JOBSHOP_SOCKET= jobshop session attach <session id>
```

The ticket carries the pool prompt into the conversation (message
prompts do not include it); the session keeps it in progress for the
whole chat.

## Vars

| Var | Meaning |
| --- | --- |
| `repo` | GitHub repo, `owner/name` |
| `checkout` | Main checkout of the repo; sessions get worktrees of it |
| `agent` | Agent name defined in the host's jobshop.yaml |
| `guide` | Repo-specific reading guide. Default: README and contributor docs |
| `operator_sessions` | `max_sessions` for the pool. Default 0 (manual) |

```yaml
packs:
  - path: /var/lib/jobshop/src/jobshop/packs/operator
    vars: {repo: elee1766/jobshop, checkout: /var/lib/jobshop/src/jobshop, agent: jcode}
```

`packs/selfdev` extends this pack with Jobshop's values, so gasserver
gets the `operator` pool through its existing live pack (`livePacks =
[ "packs/selfdev" ]`) with no host change.
