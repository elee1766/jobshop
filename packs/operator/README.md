# operator pack

A chat pool, `operator`, whose sessions talk to the operator and manage
work on one repo: create, comment and close tickets, file and comment on
GitHub issues, `jobshop send` to sessions and pools, start runs, read
ticket traces and session lists, and `jobshop reload`. When the operator
corrects behavior, the session turns it into a PR to the governing pack
prompt (merged by the refinery through a `pr-work` run) or an issue.
Replies are short; long output goes to an issue or ticket.
Prompt: `prompts/operator.md`. Design: `docs/plan.md` section 8.

`operator` is a `type: chat` pool with `attach: rw`: one session per
key, started by the first message for it. The Matrix gateway keys
sessions by thread; escalation threads (#125) use the same pool. The
pool prompt leads each session's first message turn. Without a gateway:

```sh
jobshop send --help   # by pool and key via session.send
jobshop session attach <session id> --rw
```

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
