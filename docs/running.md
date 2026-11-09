# Running a root

Moved from the README (2026-11-05, #551). Pack authoring is in
[packs.md](packs.md), isolation in [isolation.md](isolation.md),
webhooks in [ingest.md](ingest.md), NixOS in [nixos.md](nixos.md).
For service YAML, CUE packs, Starlark callbacks and migration, start at
[configuration](configuration/README.md).

## Concepts

- **Tickets.** One graph of tickets (`task`, `session`, `message`) with
  blocking edges, labels, comments, claims with leases and fences, and
  defers. The controller is the only writer. Agents use the `jobshop` CLI,
  which acts as their session. See [data-model.md](data-model.md).
- **Sessions.** Each session runs in a detached host process from
  [elee1766/acphost](https://github.com/elee1766/acphost), which spawns
  the agent, proxies ACP over `acp.sock` and journals frames for replay.
  Sessions survive controller restarts and deploys. Crash loops, stuck
  sessions and spawn errors are detected, and the pool is quarantined.
- **Pools.** Each pool has an agent, a repo (sessions get a git worktree),
  a prompt, `max_sessions` and timeouts. `type: chat` pools hold one
  session per conversation key.
- **Triggers.** These create tickets or start runs from store counts, cron,
  webhook events, Starlark handlers, or scan signals (unclaimed work,
  silent sessions, idle steps, issues without a run, green PRs nobody
  merged). See [configuration/callbacks.md](configuration/callbacks.md).
- **Workflows and runs.** Steps run in pools, with retries and
  dependencies between runs. A step that runs out of retries escalates to
  a human. Batches group runs in an order. See
  [notes/88-workflow-config.md](notes/88-workflow-config.md) and
  [notes/90-step-retries-escalation.md](notes/90-step-retries-escalation.md).
- **Packs.** A pack is a directory of pools, prompts, triggers and
  workflows filled from vars. It can extend another pack, and several
  instances can sit side by side with a `prefix:`. Templates are Go
  `text/template` (`{{.repo}}`, `{{if}}`, `default`). The service file
  selects a root pack, which loads these instances. See [packs.md](packs.md).
- **Trackers.** GitHub issues and PRs, or Linear issues. Tickets subscribe
  to issue threads, so new human comments come back to them. Identities
  map GitHub App installations and Linear tokens to repos and teams.
- **Guidance and guardrails.** `prompt_rules` inject text, files or
  command output into matching sessions' prompts. `hooks` gate claims and
  outside actions (see [isolation.md](isolation.md)).
- **Humans.** Workers report progress and ask questions (`jobshop
  report`). Questions and escalations go to the `attention` list and Desk,
  and optionally to an A2A peer (below).
- **Desk and A2A.** Desk is a web UI (React) and HTTP API over the admin
  socket. It shows tickets, runs, sessions and attention, and has chat
  with operator sessions. An A2A gateway hands escalation chats to
  another agent. The Matrix and Discord gateways were removed
  (2026-11-06). See [notes/202-desk.md](notes/202-desk.md).
- **Telemetry.** OpenTelemetry traces and logs, a Prometheus endpoint, and
  optional export of agent conversations.

## Run

```
nix build .#jobshop
jobshop init ./root          # writes root/jobshop.yaml and root/pack.yaml
cd root && jobshop serve     # controller
```

A small `jobshop.yaml` selects its work definitions:

```yaml
pack: pack.yaml
```

The root `pack.yaml` loads a reusable pack:

```yaml
default_project: js
projects:
  js: {}
agents:
  jcode: {command: jcode, args: [acp]}
packs:
  - path: packs/main               # relative to the root
    vars:
      checkout: /srv/widget        # main checkout; sessions get worktrees
      agent: jcode
```

The controller needs `gh` authenticated (`GH_TOKEN`) for triggers. Then:

```
jobshop config show                      # resolved pools, triggers, workflows
jobshop run start change --var task="add a --verbose flag" --key verbose
jobshop run list
jobshop session list
jobshop session attach <id> [--rw]
jobshop attention                        # what needs a human
jobshop reload                           # after editing yaml or packs
jobshop desk                             # web UI on 127.0.0.1:7480
```

`jobshop --help` lists the rest (tickets, edges, labels, send, batches,
reports, events, identities). The full tree is in
[../llm/cli.md](../llm/cli.md).

## Escalate to another agent

An escalation can go to another agent, not only to a chat room. A
`network: a2a` gateway sends each escalation chat to an A2A v1.0 peer,
such as Hermes with its A2A plugin. The conversation's `contextId` is the
ticket id. The peer's replies go to the escalation session, which still
resolves the escalation. Desk and humans can answer at any time, and the
first answer wins. The gateway is outbound only, so Jobshop needs no
public endpoint.

```yaml
gateway:
  hermes:
    network: a2a
    url: https://hermes.example/a2a     # card: <url>/.well-known/agent-card.json
    token_file: /run/secrets/hermes-a2a # bearer token, never in yaml
    pool: operator
    escalations: [all]                  # or causes: run, pool, wait, guard, label
    poll: 30s
```

`escalations` is the routing policy. A cause that is not listed is not
sent, and the peer decides whether to ask its own human. See
[notes/547-a2a-gateway.md](notes/547-a2a-gateway.md).

## Running as your own user

No GitHub App or Linear App is needed. Give jobshop a personal access
token and a Linear API key, each in a file only the jobshop user reads:

```yaml
identities:
  me-gh: {kind: github-token, token_file: /secrets/gh-pat, repos: ["acme/*"]}
  me-lin:
    kind: linear
    token_file: /secrets/linear-key
    workspace: acme
    teams: [ENG]
    webhook_secret_file: /secrets/linear-webhook   # optional workspace webhook
projects:
  js: {identities: [me-gh, me-lin]}
```

The PAT (fine-grained or classic) is used wherever an App installation
token would be: `jobshop identity token`, the git credential helper, the
gh wrapper and controller reads. Agents post as you, so comments by the
token's own login (`GET /user`) are ignored like the App bot's. Check
with `jobshop identity check acme/widget`. App-only features, such as
reconciling missed App webhook deliveries, are off; point a repo
webhook at an `ingest.endpoints` entry with its own `secret_file`
([ingest.md](ingest.md)). See
[notes/340-user-tokens.md](notes/340-user-tokens.md).

## Secret files

Any `*_file` value (identity `token_file`, `private_key_file`,
`client_id_file`, `client_secret_file`, `webhook_secret_file`, ingest
and admission `secret_file`, gateway `token_file`) is a path, or a
mapping `{type: file, path: ...}` or `{type: command, command: ...}`. A
command runs with `sh -c` as the jobshop user, each time the secret is
read (30s timeout), and its stdout is the secret:

```yaml
  me-gh: {kind: github-token, token_file: {type: command, command: "pass show gh/jobshop"}, repos: ["acme/*"]}
  me-lin: {kind: linear, token_file: {type: command, command: "op read op://ci/linear/key"}, ...}
```

## Secret scanning (redact)

Text leaving the controller (session events, journal reads and shipping,
command step output, stored comments, reports, learnings, artifacts) and
`jobshop gh`/`linear` writes are scanned with betterleaks' default rules
plus jobshop's built-ins (`pkg/support/redact/rules.toml`). `redact` adds rules
and turns rules off. Host-only: packs cannot set it.

```yaml
redact:
  rules:                 # betterleaks rule fields, snake_case
    - id: acme-token
      description: ACME internal token
      regex: '\bacme_[A-Za-z0-9]{32}\b'
      keywords: [acme_]  # required
      value_group: 0     # optional: capture group holding the secret
  disabled_rules: [generic-api-key]
```

A bad regex, an id already used, or an unknown id in `disabled_rules`
fails `jobshop config check` with the rule id. Reload rebuilds the
scanner; a reload with bad rules keeps the last good one.

## Journal compression

A closed session's `journal.ndjson` is compressed with zstd to
`journal.ndjson.zst` (plus a seq index) once it has been closed for
`journal.compress.after`. The sweep runs on start and hourly, one
session at a time at idle IO priority. Every journal reader (Desk,
`session attach`, recall, shipping) reads both forms. Host-only; off by
default (#631).

```yaml
journal:
  compress:
    after: 24h   # empty: off
    level: 3     # zstd 1-22
```
