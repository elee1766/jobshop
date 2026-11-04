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

`repo` is the old name of `jobshop_repo`; it still works for one
release and logs a warning (`renamed_vars` in pack.yaml, #261).

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

## Improvement sessions

A session keyed `improve:<slug>` works on improving Jobshop or a pack
(#249). It gets `prompts/improve.md` after the pool prompt (prompt rule
`operator-improve`, `match.keys`) and its own worktree of `checkout`:

```sh
jobshop send pool:operator --key improve:dev-prompt "the dev prompt should say X"
```

It either files the issue on the right repo and starts that instance's
`<prefix>pr-work`, or branches, pushes and opens the PR itself, then
starts `pr-work` and comments `PR <url> head <sha>` on the run's
`implement` step so the dev session adopts it and the refinery merges
it. Example: `jobshop run start pr-work --var issue=301 --key issue-301`,
`jobshop run show <run>`, `jobshop comment -t <implement step> "PR
https://github.com/elee1766/jobshop/pull/302 head 1a2b3c4"`.

The pack's `push` action hook denies `jobshop guard push --target`
master or main (any ref form) and a push with no target for the
`operator` pool. Notes: `docs/notes/249-improvement-sessions.md`.
