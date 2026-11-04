# jobshop

Multi-agent work orchestration on a SQLite ticket store. A controller runs
pools of ACP agents (for example `jcode acp`) that claim tickets, work
them and close them. It replaces our use of Gas City. Since 2026-11-01 it
runs the agents that develop it: issues on this repo are planned,
implemented, reviewed, merged and deployed by Jobshop pools.

```mermaid
flowchart LR
  CLI[jobshop CLI] -- admin.sock / cli.sock --> C[controller]
  C --> S[(SQLite)]
  C -- ACP over acp.sock --> H[session host, one per session]
  H -- stdio --> A[agent]
  A -- jobshop CLI --> C
  G[gateway: Matrix, Discord] --> C
  D[Desk: web UI] --> C
```

## What it does

- **Tickets.** One graph of tickets (`task`, `session`, `message`) with
  blocking edges, labels, comments, claims with leases and fences, and
  defers. The controller is the only writer. Agents use the `jobshop` CLI,
  which acts as their session.
- **Sessions.** One detached host process per session runs the agent,
  proxies ACP and journals every frame. Sessions survive controller
  restarts and deploys. Crash loops, stuck sessions and spawn errors are
  detected, and the pool is quarantined.
- **Pools.** Each pool has an agent, a repo (sessions get a git worktree),
  a prompt, `max_sessions` and timeouts. `type: chat` pools hold one
  session per conversation key.
- **Triggers.** These create tickets or start runs from store counts, cron,
  a shell check, or stall signals (unclaimed work, silent sessions,
  stalled steps, issues without a run, green PRs nobody merged).
- **Workflows and runs.** Steps run in pools, with retries and
  dependencies between runs. A step that runs out of retries escalates to
  a human. Batches group runs in an order.
- **Packs.** A pack is a directory of pools, prompts, triggers and
  workflows filled from vars. It can extend another pack, and several
  instances can sit side by side with a `prefix:`.
- **Trackers.** GitHub issues and PRs, or Linear issues. Tickets subscribe
  to issue threads, so new human comments come back to them. Identities
  map GitHub App installations and Linear tokens to repos and teams.
- **Guidance and guardrails.** `prompt_rules` inject text, files or
  command output into matching sessions' prompts. `hooks` allow, deny or
  ask a human before outside actions (`jobshop guard`).
- **Humans.** Workers report progress and ask questions (`jobshop
  report`). Questions and escalations go to the `attention` list, a chat
  thread through the gateway (Matrix or Discord), and Desk.
- **Desk.** A web UI (Svelte) over the admin socket. It shows tickets,
  runs, sessions and attention, and has chat with operator sessions.
- **Telemetry.** OpenTelemetry traces and logs, a Prometheus endpoint, and
  optional export of agent conversations.

## Packs in this repo

| Pack | What it is |
| --- | --- |
| `packs/github` | Works one GitHub repo (or Linear team) through issues and PRs. A planner reads a `target` issue and files work, `pr-work` runs implement, review-merge and post-merge, `ci-repair` fixes red master, and `stall-check` unsticks work. See its README. |
| `packs/selfdev` | `packs/github` with Jobshop's values. It builds this repo. |
| `packs/operator` | A chat pool for the operator: it manages tickets, issues and runs, and turns corrections into prompt PRs. |

## Run

```
nix build .#jobshop
jobshop init ./root          # writes root/jobshop.yaml
cd root && jobshop serve     # controller
```

A small `jobshop.yaml` with one GitHub repo:

```yaml
default_project: js
projects:
  js: {}
agents:
  jcode: {command: jcode, args: [acp]}
packs:
  - path: /path/to/jobshop/packs/github
    vars:
      repo: acme/widget
      checkout: /srv/widget        # main checkout; sessions get worktrees
      agent: jcode
      human: wile                  # GitHub user mentioned on escalations
      check_cmd: make test
```

The controller needs `gh` authenticated (`GH_TOKEN`) for triggers. Then:

```
jobshop config show                      # resolved pools, triggers, workflows
jobshop run start pr-work --var issue=12 --key issue-12
jobshop run list
jobshop session list
jobshop session attach <id> [--rw]
jobshop attention                        # what needs a human
jobshop reload                           # after editing yaml or packs
jobshop desk                             # web UI on 127.0.0.1:7480
```

`jobshop --help` lists the rest (tickets, edges, labels, send, batches,
reports, events, identities).

### Running as your own user

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
webhook at an `ingest.endpoints` entry with its own `secret_file`. See
`docs/notes/340-user-tokens.md`.

### Webhooks: `jobshop ingest serve`

`jobshop ingest serve` receives GitHub and Linear webhooks, checks the
signature, and writes each delivery to a spool database
(`ingest.spool`, default `<root>/ingest/spool.db`). It needs no
controller and does not touch jobshop.db. List one endpoint per hook,
each with its own secret; GitHub endpoints can limit which repos they
accept:

```yaml
ingest:
  endpoints:
    - {name: app, source: github, secret_file: secrets/gh-app, repos: [acme/app]}
    - {name: eng, source: linear, secret_file: secrets/linear-eng}
```

Each is served at `/<source>/<name>` unless `path` is set. A linear-app
identity with `webhook_secret_file` is also served at `/linear/webhook`.
It listens on `ingest.listen` (default `127.0.0.1:8740`), so put a
reverse proxy in front that terminates TLS and forwards only the
endpoint paths:

```
location /github/ { proxy_pass http://127.0.0.1:8740; }
location /linear/ { proxy_pass http://127.0.0.1:8740; }
```

It answers 202 when spooled, 204 for an event jobshop ignores, 401 for a
bad signature, 403 for a repo not on the endpoint's list, and 500 when
the spool write fails so the sender retries. Adding an endpoint needs a
restart; rotating a secret file does not. See
`docs/notes/334-ingest-spool.md`.

### NixOS

The flake exports `nixosModules.default`. It gives `services.jobshop` with
`settings` (the yaml), `workspaces.<name>` (repos cloned before start,
with live packs that follow the checkout), `gateway.<name>`,
`identities`, telemetry, and `selfDeploy`. Self-deploy promotes the
newest green master commit and rolls back on a failed health check. See
`nix/module.nix`, and `docs/layout.md` for deployment.

## Develop

```
go build ./... && go vet ./... && go test ./...
golangci-lint run ./...       # v2.15.0; or nix run .#golangci-lint
```

- Go 1.27, `CGO_ENABLED=0`. The store uses sqlc. Regenerate after
  editing `pkg/store/sqlite/queries/` or `migrations/`.
- Desk UI source is `ui/` (Svelte 5, Vite). It builds into
  `pkg/desk/dist`, which is committed and embedded. CI checks it is
  current.
- A change to `go.mod` or `go.sum` needs a new `vendorHash` in
  `nix/package.nix` (see AGENTS.md).
- Contributor rules for people and agents are in `AGENTS.md`.

## Docs

- `docs/plan.md`: design, phases and the dated decision log.
- `docs/layout.md`: packages, dependencies, build and deployment.
- `docs/data-model.md`: tables and queries.
- `docs/notes/<issue>-<slug>.md`: behavior as built, one note per change.
- `docs/verify/`: runbook checks run against a live host.
- Progress: the open issues labeled
  [`target`](https://github.com/elee1766/jobshop/issues?q=is%3Aopen+label%3Atarget).
