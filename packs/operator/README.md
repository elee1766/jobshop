# operator pack

A chat pool, `operator`, whose sessions talk to the operator and manage
work across every pack instance on the host: create, comment and close
tickets, file and comment on issues in each instance's repo, `jobshop
send` to sessions and pools, start runs with the instance's prefixed
workflow, read ticket traces and session, pool and trigger lists, and
`jobshop reload`. When the operator corrects behavior, the session turns
it into a PR to the governing pack prompt on Jobshop's repo (merged by
the refinery through a `pr-work` run) or an issue. Replies are short;
long output goes to an issue or ticket. Prompt: `prompts/operator.md`.
Design: `docs/plan.md` section 8, host scope in
`docs/notes/248-host-operator.md`.

Include it once per host, without a prefix, next to the repo packs
(#248). The session reads the instances from `jobshop config show`
(`instances`: path, prefix, vars such as `repo`, `checkout`, `tracker`,
`linear_team`, and the pools, workflows and triggers each declares), so
the prompt names no repo.

`operator` is a `type: chat` pool with `attach: rw`: one session per
key, started by the first message for it. The Matrix gateway keys
sessions by thread; escalation threads (#125) use the same pool. The
pool prompt leads each session's first message turn. Without a gateway:

```sh
jobshop send --help   # by pool and key via session.send
jobshop session attach <session id> --rw
```

Admin grant: `config-reload`, `config-show`, `session-list`,
`pool-list`, `trigger-list`, `attention`. Tickets and run list/show need
no grant.

## Vars

| Var | Meaning |
| --- | --- |
| `jobshop_repo` | Where Jobshop changes (prompt fixes) go, `owner/name` |
| `checkout` | Jobshop's checkout; sessions get worktrees of it |
| `agent` | Agent name defined in the host's jobshop.yaml |
| `operator_sessions` | `max_sessions` for the pool. Default 0 (manual) |

```yaml
packs:
  - path: /var/lib/jobshop/src/jobshop/packs/selfdev
    vars: {repo: /var/lib/jobshop/src/jobshop, agent: jcode}
  - path: /var/lib/jobshop/src/jobshop/packs/github
    prefix: acphost-
    vars: {repo: elee1766/acphost, checkout: /var/lib/jobshop/src/acphost, agent: jcode, human: elee1766, check_cmd: make test}
  - path: /var/lib/jobshop/src/jobshop/packs/operator
    vars: {jobshop_repo: elee1766/jobshop, checkout: /var/lib/jobshop/src/jobshop, agent: jcode}
```

`packs/selfdev` no longer includes this pack. With the NixOS module, add
the entry to `settings.packs` with the jobshop workspace's directory as
the path (not to `livePacks`: those share the workspace's vars and
prefix).
