# Jobshop: draft implementation plan

Status: **draft for discussion.** This is a starting point, not a decision
record. Every section lists open questions, and estimates are rough. Challenge
anything here; where this plan and the evidence disagree, the evidence wins.

## 1. Why this exists

Jobshop is a multi-agent orchestration layer modeled on
[Gas City](https://github.com/gastownhall/gascity) (fork:
`~/repo/github.com/elee1766/gascity`, branch `feat/sqlite-work-provider`).
Gas City works for us, but:

- Its default store is `bd` (the beads CLI) on Dolt. Dolt was too slow and
  resource-heavy for our deployment, so we run a fork with a native SQLite
  store (`[beads] provider = "sqlite-work"`).
- Gas City's Go domain logic is written against a generic store interface
  (`internal/beads/beads.go`, `beads.Store`), and SQLite plugs in there.
  The friction is elsewhere: pack prompts and formulas call `gc bd ...`, which
  on Dolt cities shells out to the `bd` binary. On `sqlite-work`, `gc bd` is a
  hand-written native adapter (`cmd/gc/cmd_bd_sqlite_work.go`) that covers only
  part of the `bd` surface (no `mol wisp`/`burn`, `formula`, `gate`, `blocked`,
  `--include-infra`, ...).
- Roughly 18% of non-test Go files reference Dolt (server lifecycle, version
  gates, proxied transport, split-store guards, doctor checks). Each upstream
  merge adds more Dolt/`bd` assumptions to work around.
- We use a small subset of Gas City, and extending it feels at odds with
  upstream's direction.

The goal is a smaller system we fully control, where the ticket store is a
storage interface owned by this project, everything is built on it, and SQLite
is the first backend (Postgres later).

**Approach: build independently and dogfood.** Jobshop is not a migration
target for the live Gas City deployment. It is built from scratch, and as soon
as it can run agents, it runs the agents that work on Jobshop itself. Gas City
stays as a reference for design and edge cases, not as code we must stay
compatible with. Section 2 describes how we use Gas City today so readers know
which features have proven useful, not as a parity checklist.

### Open questions
- How much upstream Gas City progress do we expect to want later?
- At what point, if any, would the live city move onto Jobshop?

## 2. How we use Gas City today

Snapshot of the live city (`gasserver:/home/agent/gc`, read-only, 2026-11-01).
Treat as context, not spec; verify before relying on details.

**Storage**
- `[beads] provider = "sqlite-work"` for work tickets, per city and per rig.
- `[storage.classes]`: `graph`, `sessions`, `messaging`, `orders`, `nudges`
  all routed to binding `infra` (`provider = "sqlite-beads"`, `.gc/store`).
- `event_hooks = false`.

**Runtime**
- Single provider family: `jcode` via ACP (`session = "acp"`), with variants
  per model (opus for mayor/workers, sonnet for patrol/emulator, gpt-6.1-sol).
- `max_active_sessions = 20` city-wide, 14 per rig, pools of 1-6.

**Agents (from the imported `gastown` pack, all re-prompted locally)**
- `mayor`: coordinator (`prompts/coordinator.md`), on demand.
- `polecat`: worker pool per rig (`prompts/worker.md`), 1-6 sessions.
- `refinery`: PR delivery per rig (`prompts/refinery.md`), max 1.
- `deacon`, `boot`, `witness`: observers (`prompts/observer.md`, 9 lines).
  The observer prompt explicitly says the controller owns reconciliation; they
  only report blockers by mail.
- `dog`: suspended. Several implicit provider pools: suspended.
- Extra packs: `model-routing`, `android-emulator` (checker pool).

**Rigs**: account, retail, trade, gascity, sap, uniswapgo.

**Formulas in use**: `mol-gfx-pr-work` (default sling formula, one durable
source plus one step), `mol-gascity-delivery`, `mol-blocker-assess`,
`mol-main-ci-repair`, `mol-gfx-emulator-check`, `mol-gfx-pilot`, plus the
gastown patrol formulas.

**Orders**: most built-ins skipped (`jsonl-export`, `reaper`, `wisp-compact`,
`gate-sweep`, `cross-rig-deps`, ...). Local orders: `blocked-revalidator`
(hourly), `gascity-delivery`.

**Commands agents actually call** (from our prompts and formulas, approximate):
`gc hook --claim --json`, `gc bd update --claim`, `gc bd heartbeat`,
`gc bd show/update`, `gc mail send/reply/peek/archive`, `gc convoy status`,
`gc sling`, `gc session list/peek/wait`, `gc runtime drain-ack`,
`gc storage status`, plus our own `gx` CLI (`task context/finish`,
`pr state`, `mayor brief/record`, `linear sling`, `emulator request`).

**Implication**: our real surface is claim/heartbeat/finish on tickets,
routing work to pools, a small set of formulas, mail, and session supervision.
We do not use most of the gastown patrol machinery as designed.

### Decided (2026-11-01)
- Day-one scope is the phase plan (section 7).
- Observer agents (deacon/boot/witness): Jobshop ships none. Gas City
  already moved their patrol duties into the controller (its parity audit
  maps witness and deacon patrol to the reconciler tick and controller
  loop), and in the live city they send almost no mail. In Jobshop those
  duties are deterministic: dead-session release, lease sweep, idle-kill,
  quarantine, metrics and alerts. The layer for observer-style agents must
  still exist: roles are config, so any install can define an observer
  pool, and scheduled judgment work (e.g. stale PR review) runs as a
  schedule assigning tickets to a coordinator pool. We just do not build
  or run one ourselves.
- `gx` stays a separate client of the Jobshop CLI/admin socket.

## 3. Glossary

Decided 2026-11-01: plain, descriptive names, no themed vocabulary. Prefer
the term that says what the thing is; keep Gas City's term only where it is
already plain. "Jobshop" stays as the project name.

| Gas City | Jobshop | Notes |
|---|---|---|
| Controller loop (incl. health patrol) | Controller | Reconcile, restart, quarantine, idle-kill |
| Control dispatcher | Dispatcher | Runs control tickets (retry, fan-out, ...) |
| Bead | Ticket | |
| Formula | Workflow | Template definition |
| Molecule | Run | Instantiated workflow (tickets + edges) |
| Wisp | Ephemeral run | A flag on run, not a separate concept |
| Agent | Agent | Configured kind of worker |
| Pool | Pool | |
| Session | Session | |
| Rig | Project | One repo / work area |
| Pack | Pack | Bundle of agents, prompts, workflows |
| Sling | Assign | Puts a ticket in a pool's queue; not a claim |
| Claim / release | Claim / release | Claimant field is `holder`, not `assignee` |
| Gate | Gate | Passive hold until closed |
| Check | Check | Condition evaluated by the dispatcher |
| Convoy | Batch | Group of tickets tracked together |
| Order | Schedule | Scheduled workflow trigger |
| Mail | Message | |
| Event log | (none) | No event bus. Controller sees every write and notifies over the admin socket; ticket `history` is an audit log. Observability is OTel (section 9) |

Roles (mayor, polecat, ...) are configuration, not engine concepts. Jobshop
keeps Gas City's "zero hardcoded roles" principle.

## 4. Design principles (proposed)

Carried over from Gas City's `AGENTS.md` where they still fit:

- No role names in Go. Behavior comes from config and prompts.
- Go handles transport, not judgment.
- Work persists in tickets; sessions are disposable. Recovery means re-reading
  state, not remembering it.
- No status files; query live state.

New for Jobshop:

- **The ticket store is the only persistence interface**, owned by this
  repo. No CLI passthrough to an external tracker.
- **One CLI, native on the store.** Agent-facing commands are designed for
  Jobshop, not for compatibility with `bd`'s flags or JSON.
- **Backend-neutral SQL** where practical, so Postgres is an implementation,
  not a rewrite.
- **YAML config** (decided 2026-11-01).

### Decided (2026-11-01)
- Generic graph of tickets with an engine-defined `kind`, not separate
  tables per concept (sessions and messages are kinds too).
- Kind-specific engine data is typed per kind in Go. Packs get a free-form
  `meta` document the engine never reads. Packs do not declare schemas.
  See `docs/data-model.md`.

## 5. Ticket store

Full tables and queries: `docs/data-model.md`.

Starting point: Gas City's `beads.Store` (24 methods: create, get, update,
close, reopen, delete, list and filter, `Ready`, children, metadata, `Tx`,
`DepAdd`/`DepRemove`/`DepList`). Things Gas City grew later that seem worth
building in from the start:

- **Atomic claim** with a lease and a fence, plus heartbeat and release
  conditioned on the current holder. See `sqlite_store_claim.go`,
  `sqlite_store_fence.go`, `sqlite_store_conditional.go`.
- **Revision-guarded updates** (compare-and-swap).
- **Ready frontier query** (open, unblocked, not deferred).
- **Graph apply**: create a run's tickets and edges in one transaction
  (`sqlite_store_graph_apply.go`).
- **Ephemeral flag** for short-lived tickets, reaped after close.
- **Engine-defined edges**: `blocks` and `waits_for` block readiness;
  `tracks` is informational. Parent-child is the `parent_id` column.

A conformance test suite that every backend must pass, written before the
SQLite implementation.

### Decided (2026-11-01)
- One database per install. `project` is a column. No storage classes, no
  split stores.
- SQLite driver: `modernc.org/sqlite` (pure Go, no cgo).
- Ticket IDs: per-project prefix plus a Sonyflake ID in 13-character
  Crockford base32, e.g. `js-0031w02j0r012`. Details: `docs/layout.md`.
- Claims are released by the controller as soon as it knows a session is
  dead. The lease is a backstop, swept by the controller every tick, and
  expiry returns the ticket straight to `open`.
- Comments have their own table. `history` is an audit log, not an event
  stream.
- CLI and admin sockets speak JSON-RPC 2.0 with NDJSON framing, matching
  ACP, using the connection from `github.com/coder/acp-go-sdk`
  (decided 2026-11-01; replaced `go.lsp.dev/jsonrpc2`, see
  `docs/layout.md`).

## 6. What could be reused from Gas City

Gas City is MIT-licensed. Dependency scan of `internal/` (non-test lines):

**No beads/Dolt dependency; likely portable:**
- `internal/formula` (~10k lines): workflow parsing and compile. Depends on
  `beadmeta`, `git`, `gitcred`, `remotesource`, `pathutil`, `gchome`.
- `internal/runtime` (~12k) and `internal/runtime/tmux` (~16k). We use ACP
  (`internal/runtime/acp`); check its dependencies before relying on it.
- `internal/events` (~8k), `internal/worktree`, `internal/shellquote`,
  `internal/fsys`, `internal/clock`.

**Logically store-neutral but import `internal/beads` (which pulls in the
upstream beads/Dolt library):**
- `internal/molecule` (~3.7k): instantiate a workflow into a run.
- `internal/dispatch` (~12k): control ticket execution.
- `internal/readywork`, `internal/convergence`, `internal/orders`.
- SQLite store code in `internal/beads/sqlite_*.go` (~4.8k).

**Probably better rewritten, using Gas City as reference:**
- `internal/session`, `internal/worker`, `internal/sling`, `cmd/gc`
  controller. They carry a lot of upstream policy and 45-60 internal
  dependencies each.

Estimates come from `go list -deps` and line counts on 2026-11-01. They do not
account for test code or hidden coupling.

### Decided (2026-11-01)
- Copy Gas City code only where its internals are clean (self-contained,
  few dependencies, readable). Everything else is reference only. No
  vendoring.
- Dispatcher scope is decided in phase 4 from the workflows we actually
  run.

## 7. Phases

Decided 2026-11-01. Each phase ends with something runnable. The order is
driven by self-hosting: build the smallest loop that can run an agent on a
Jobshop ticket in this repo, then let that loop build the rest. Until then,
development happens by hand (or with an outside agent).

**Phase 0: groundwork**
- Repo layout, CI, lint. Plain `go` for development and CI; Nix flake for
  the deployable package and NixOS module (`docs/layout.md`). No Bazel.
  CI on GitHub Actions: build, `go vet`, `golangci-lint`,
  `go test -race`, `nix build`.
- Data model: `docs/data-model.md`.

**Phase 1: store, controller sockets, CLI** (done 2026-11-01)
- `store` package: interface, types, errors, engine kinds (`task`,
  `session`, `message`) with typed fields.
- Conformance suite, then the SQLite backend.
- Controller process skeleton: opens the store, serves the admin socket
  and per-session CLI sockets over JSON-RPC (section 8.5). All writes go
  through it.
- CLI: `jobshop ticket create/show/list/update/close`, `comment`, `assign`,
  `claim --next` (three tiers), `heartbeat`, `release`, `ready`, with
  stable JSON output. Never opens SQLite directly.

**Phase 2: sessions and self-hosting**
- YAML config loader (project, agent, pool).
- Session host (section 8): spawn via `/proc/self/exe`, ACP proxy, journal,
  replay, backpressure, leak guards, controller restart and adoption.
  Required for self-hosting, since agents will redeploy the controller.
- ACP runtime via jcode, resume with `session/load`.
- Controller tick: start a session when a pool has ready work, restart it
  if it dies, release a dead session's claims, sweep expired leases.
- Pinned release binary and promote/rollback on gasserver.
- **Milestone: Jobshop runs its own agents.** One project (this repo), one
  pool, agents claim and close tickets directly and deliver via PR as
  `cog`. From here, new work is filed as Jobshop tickets.

**Phase 3: hardening and visibility** (built by Jobshop agents)
- Quarantine, idle-kill, drain.
- Observability (section 9): OTel traces and logs, Prometheus metrics,
  conversation export.
- Messages with queued and interrupt delivery.
- Human attach.

**Phase 4: workflows and runs**
- Port the formula concepts we use; decide which workflow features to keep.
- Run creation via graph apply.
- Only the dispatcher features our workflows need; retry limits live here.

**Phase 5: the rest of our surface**
- Batches (convoys), schedules (orders).
- Admin-socket change subscriptions and an API if the dashboard is ported.
- Gateway prerequisites (section 10) are already covered by phases 1-3.

**Later**: Postgres backend, against the same conformance suite.

**Isolation (decided 2026-11-01)**: no bwrap or separate agent user at the
start. Everything runs as the `jobshop` user on gasserver (an isolated OVH
machine), in its own cgroup slice. Per-session sandboxing (section 8.6)
stays an option for later.

**Self-hosting setup (decided 2026-11-01)**
- Host: `gasserver` (isolated OVH machine), next to the live Gas City.
  Work happens over SSH. No bwrap at the start. (agent-server is a
  separate machine and not used.)
- GitHub identity: the `cog` user, with write access to this repo to push
  branches, open PRs and merge.
- Isolation from Gas City: own Linux user `jobshop`, own data dir and
  SQLite file, own controller process, and its own cgroup slice capped at
  6 cores and 8 GB (decided 2026-11-01), so it cannot starve the live city.
  Deployment details: `docs/layout.md`.
- Known-good binary: the NixOS module pins a specific `jobshop` store path.
  Promotion is bumping the flake input and switching; rollback is the
  previous generation.
- Agents work in git worktrees, one per ticket, and deliver via PR.
- Merge policy: branch protection requires CI. `cog` merges when CI passes,
  for all changes (decided 2026-11-01). The pinned binary is the safety net.

After the milestone, add what self-hosting shows is missing, in the order
it hurts. Bring over Gas City features only when a real need appears.

## 8. Sessions: host process, hot restart, CLI socket (proposed)

Goal: restart or upgrade the controller without killing agent sessions or
losing in-flight turns, with no tmux or systemd dependency, and give agents
a CLI that knows which session it belongs to.

### 8.1 Shape

```mermaid
flowchart LR
  C[controller<br/>user: agent] -- ACP over acp.sock --> H[session-host<br/>user: agent]
  H -- stdin/stdout pipes --> A[agent process, e.g. jcode ACP<br/>user: jobshop-agent later]
  A -- jobshop CLI over cli.sock --> H
  H -- scoped ticket ops --> C
  C --> S[(SQLite store)]
  H --> J[(session journal)]
```

- One `session-host` process per session. A host bug or crash loses one
  session, and host upgrades roll out as sessions turn over. Cost is about
  10 MB per idle host. A single shared daemon would reintroduce a process
  whose restart kills every session.
- Prior art (none in Go): hydra-acp (TypeScript, one shared daemon),
  karet-agentd proposal (Rust, same proxy shape as this), acpx/acpxd
  (TypeScript/Swift; their PRs document reconnect bugs worth avoiding),
  aptove-bridge (Rust). ACP drafts: RFD #533 (multi-client attach,
  `session/detach`) and the session-resume RFD (proxy that journals
  messages and serves `session/load`). Read hydra-acp and the acpxd PRs
  before writing the host.

### 8.2 Spawning

- The controller runs `exec.Command("/proc/self/exe", "session-host",
  <id>)` with `Args[0] = "jobshop"`, `SysProcAttr{Setsid: true}`, stdin
  from `/dev/null`, stdout/stderr to `<root>/sessions/<id>/host.log`. It
  does not wait on the child, so the host outlives the controller. No
  tmux, no systemd, no hand-rolled double fork (decided 2026-11-01).
- New process, not a fork: Go cannot fork safely, and `os/exec` is
  fork+exec. The host inherits nothing except what is passed explicitly.
- `/proc/self/exe` is for correctness, not security: the host runs the
  exact build of the controller that spawned it, even if the release
  symlink moves. Security comes from users and file ownership (8.6).
- Before spawning, the controller writes a session spec to
  `<root>/sessions/<id>/spec.json`: agent command, args, env, working
  dir, socket paths, host protocol version, binary path.
- A host keeps running the build it started with. Keep each release
  binary until no live host references it; `session reap` reports which
  are still in use.

### 8.3 What the host does

1. Reads the spec, writes a pidfile and takes a lock in its session dir.
2. Starts the agent with pipes on stdin/stdout, `Pdeathsig: SIGKILL`,
   in its own process group. Marks itself a child subreaper.
3. Listens on `acp.sock` (controller) and `cli.sock` (agent CLI).
4. Proxies ACP between `acp.sock` and the agent's pipes. It is not a full
   ACP client and mostly forwards frames. It parses only:
   - `initialize`: caches the agent's response and replays it to a
     reconnecting controller instead of re-initializing the agent.
   - Agent-to-client requests (permission prompts, fs/terminal calls):
     held while no controller is connected.
   - Every agent frame gets a sequence number and is appended to a
     journal. A reconnecting controller sends its last seen sequence and
     the host replays the gap.
   - Optionally, `session/load` served from the journal, so resume works
     even if the agent does not advertise `loadSession` (karet's idea).
5. Serves `cli.sock` for the agent's `jobshop` commands (8.5).
6. Exits when the agent exits, after flushing the journal.

The host holds no ticket or policy logic, so it rarely needs to change.
Its protocol to the controller is versioned so a new controller can adopt
hosts started by an old one.

### 8.4 Hot restart and leaks

Controller restart sequence:
1. New binary starts, opens the store, takes the controller lock.
2. Scans `<root>/sessions/*/` and the sessions table. Reconnects to live
   hosts, renews their lease, replays missed output. Marks unreachable
   ones dead.
3. Resumes the normal tick. No sessions are started or stopped until
   reconciliation finishes.

None of the leak guards below fire on a controller restart: they tie the
agent to its host, not to the controller.

Leak prevention, layered:
1. Agent cannot outlive host: `Pdeathsig: SIGKILL`, and the host kills its
   process group on exit.
2. Agent's own children cannot escape: host is a child subreaper, so
   orphans reparent to it and die with the group.
3. Host cannot outlive its usefulness: exits when the agent exits; holds a
   lease and shuts down if no controller has connected for a configurable
   window (default around 30 minutes).
4. Controller reaps unknowns: each host is recorded with PID, process
   start time and socket. Reconciliation kills hosts with no record (start
   time checked against `/proc` to avoid PID reuse) and removes stale
   records and sockets. The pidfile written before the record covers the
   spawn-then-crash window.
5. Visible: `jobshop session list --all` shows store vs live processes;
   `jobshop session reap` cleans up.
   (2026-11-01) `session list` is implemented (admin socket, method
   `session.list`). Each session ticket (open and in-progress by default,
   closed too with `--all`) is joined with its session dir: `alive`
   (`host.Alive`), `host_pid` (host.pid), `agent_pid` (from the host's
   welcome, known once this controller connected), `current` (the ticket
   it holds), `seq` / `seq_time` (last journal frame) and `exit` (exit.json
   reason, dead hosts only). Directories under `<root>/sessions/` with a
   live host and no open session ticket are listed with `orphan: true`;
   so is a closed session whose host still runs. `session reap` is not
   implemented yet.

Not planned: upgrading a running host in place (re-exec preserving fds).
Possible later; see the Pdeathsig thread caveat noted in discussion.

As built in phase 2 (decided 2026-11-01):
- The CLI socket stays in the controller, not the host. The host only
  proxies ACP. Agent CLI calls fail while the controller restarts, and
  agents retry; this kept the host small (one stdio pipe, one socket).
- Host files in `sessions/<id>/`: `spec.json`, `host.lock` (flock; held
  means alive), `host.pid`, `host.log`, `acp.sock`, `journal.ndjson` (every
  frame with a seq), `exit.json`. The host rewrites request ids so a new
  controller never collides with an old one, caches the `initialize`
  result, re-sends unanswered agent requests to a new controller, and
  reports a prompt that finished while detached as `_jobshop/turn_end`.
- Stopping a host is SIGTERM to its pid (it forwards to the agent group,
  SIGKILL after 5s); the controller SIGKILLs the host after 10s.
- Reconciliation on start: open session tickets are adopted (live host) or
  restarted (new host, `session/load` if advertised, else a new
  conversation). Hosts in `sessions/` with no open ticket are stopped.
  Item 4 above (records with PID start time) and `session reap` are not
  built; the flock replaces the PID-reuse check.
- Driver loop per session: `claim --next` as the session, prompt the
  ticket, then check the session no longer holds it. If it does, one
  reminder turn, then release with a 10 minute deferral. No work for
  `idle_timeout` (default 5m) ends a pool session (`idle`); more than 3
  restarts in 10 minutes ends it (`crashed`). Turns over `turn_timeout`
  (default 2h) get `session/cancel`.
- Permission requests are auto-approved (allow_always, else allow_once).
  fs and terminal client methods are not offered.
- Supervisor tick: heartbeat claims held by driven sessions, then per pool
  start sessions while ready work exceeds open sessions, up to
  `max_sessions`. Sessions whose agent is not in the config are external
  (a human runs an agent with `JOBSHOP_SOCKET`) and are not driven.
- `jobshop shutdown` leaves sessions running for the next controller;
  `shutdown --all` ends every session first.
- With `repo` set, each session gets `git worktree add --detach` of `ref`
  at `sessions/<id>/work`, removed when the session ends.
- `jcode acp` by default runs tools in the shared user daemon, which does
  not inherit the agent's env (`JOBSHOP_SOCKET`, `PATH`). Configure it as
  `jcode acp --socket ${SESSION_DIR}/jcode.sock`: a private daemon per
  session that inherits the env and exits with the session (verified
  2026-11-01). Agent args expand `${SESSION_DIR}`.
- `defer_until` gates only automatic pickup (`claim --next`, ready). A
  claim by id ignores it, as in Gas City (decided 2026-11-01): ids are
  deliberate, usually the session finishing its own work. Clear a hold with
  `jobshop ticket update <id> --defer 0`.

Recovery below the host:
- Agent or host died: start a new agent and resume with `session/load`
  (agent-native if advertised, else from the host journal). Recovers
  conversation, not in-flight tool calls.
- Everything lost: work state is in tickets; the replacement session
  re-claims via `claim --next`.

### 8.5 CLI over sockets

- Agent-facing `jobshop` commands never open the database. They talk to a
  socket.
- The host starts the agent with `JOBSHOP_SOCKET=<root>/sessions/<id>/cli.sock`.
  Every `jobshop` call from the agent and its subprocesses inherits it.
- The socket identifies the session, so calls inherit its holder,
  current ticket, project, pool and worktree. `jobshop close` with no
  arguments closes the session's current ticket. No `--holder` or
  `--session` flags in prompts.
- The host forwards requests to the controller tagged with the session;
  the controller enforces scope: only this session's tickets plus what the
  pool's config allows.
- Humans and the controller's own tooling use an admin socket
  (`<root>/admin.sock`, owned by `agent`, `0700`) with full access. The CLI
  uses it when `JOBSHOP_SOCKET` is unset.
- Build this in phase 1, since retrofitting direct-DB CLI calls is the
  expensive part.
- As built in phase 1 (decided 2026-11-01):
  - Until the session host exists, the controller itself listens on each
    open session's `cli.sock` and tags calls with that session. In phase 2
    the host owns the socket and forwards, with the same scope rules.
  - "Current ticket" is derived, not stored: the session's single
    in-progress claim. Holding zero or several means the id is required.
  - Session scope: project and holder are fixed by the socket; sessions
    cannot assign, reopen, create sessions or shut down. Closing a session
    ticket removes its socket and releases everything it held.
  - Exit codes for agents: 3 no work, 4 claim lost / not holder,
    5 not found, 6 revision conflict, 1 anything else.

### 8.6 Users and security

- Now: controller, hosts and agents all run as the `jobshop` user inside
  its own slice on gasserver (decided 2026-11-01).
  Accepted risk: an
  agent can do anything the controller can (edit the store, replace
  binaries, signal processes, connect to other sessions' sockets). Socket
  scoping guards against mistakes, not hostile agents.
- Later, if needed: agents run as a separate user,
  `jobshop-agent`. The host starts the agent via a narrow sudoers rule
  (`agent` may run only the agent command as `jobshop-agent`, NOPASSWD;
  on NixOS via `security.sudo.extraRules`). The data dir, store, binaries
  and admin socket stay `agent`-owned and `0700`, so agents cannot touch
  them and cannot ptrace or signal the controller or hosts.
  Needs: worktrees owned by or group-shared with `jobshop-agent`, and
  `cog` git credentials and model-provider auth readable by it.
- Later, if needed: per-session isolation (agents can still read each
  other's worktrees and sockets under one shared user). Options: the host
  starts each agent under bubblewrap (no root; read-only system and
  toolchain, read-write worktree, its own CLI socket, network on),
  Landlock applied by the host, or a pool of agent users. bwrap also
  removes the need for the separate `jobshop-agent` user and sudoers rule.

### 8.7 Libraries

- Decided 2026-11-01: `github.com/coder/acp-go-sdk` on both sides
  (`AgentLoader` for `session/load`). The same package's `Connection`
  already carries the CLI and admin sockets, so there is one JSON-RPC
  implementation in the binary. The host may use only its framing; the
  controller uses the client side.
- Gas City's `internal/runtime/acp` is a hand-rolled JSON-RPC client;
  useful as reference for edge cases, not as a dependency.

### 8.8 Build order

- The CLI socket and the session host are both required for the
  self-hosting milestone (phase 2), since agents will redeploy the
  controller.

### Decided (2026-11-01)
- jcode's ACP server supports `session/load`, so resume after an agent or
  host death uses it directly. Journal-backed `session/load` in the host is
  not needed for jcode; keep it only as a fallback for agents without it.
- When the host's buffer is full and no controller is connected, it stops
  reading the agent's stdout (backpressure pauses the agent). It never
  drops data.
- Orphaned host lease: 30 minutes without a controller connection.
- CLI socket protocol: JSON-RPC over the acp-go-sdk connection (section 5).

- Human attach (decided 2026-11-01): read-only to start (tail the
  journal). The attach protocol includes an input method from day one,
  disabled by config, so enabling write access later needs no protocol
  change.

### Open questions
- None currently.

## 9. Observability (proposed)

Decided 2026-11-01: full OpenTelemetry tracing and logging, Prometheus
metrics, and opt-in export of ACP conversation content. This replaces Gas
City's `internal/events` JSONL bus for observability purposes.

**Operational telemetry (always available, configured in YAML)**
- Traces: OTel SDK, OTLP exporter. Spans for controller ticks, socket
  requests (CLI and admin), store transactions, session lifecycle (spawn,
  adopt, restart, reap), and claim/heartbeat/close. Trace context is
  propagated over the CLI socket, so an agent's `jobshop close` shows up
  under its session's trace.
- Logs: `log/slog` with an OTel bridge, so logs carry trace and span ids.
  JSON to stderr/file by default, OTLP when configured.
- Metrics: Prometheus `/metrics` on a configurable listen address (OTel
  metrics SDK with the Prometheus exporter, so OTLP push is also possible).
  Starting set: sessions by pool/state, session restarts, queue depth and
  ready count per pool, claim latency, lease expiries, ticket closes by
  resolution, store transaction latency, host reconnects and replay size,
  conversation export drops.
- Common attributes: `jobshop.project`, `jobshop.pool`,
  `jobshop.session.id`, `jobshop.ticket.id`, `jobshop.kind`.

**Conversation export (opt-in, off by default)**
Same scope as our Gas City branch's `internal/convtel`
(`docs/guides/conversation-telemetry.md` in the fork): ACP content as OTLP
logs and traces for later analysis in a log store. Rebuilt on the OTel Go
SDK, sharing providers and exporters with operational telemetry; convtel's
hand-rolled OTLP/HTTP code is not reused (decided 2026-11-01). Its doc is a
reference for record shapes and redaction rules only.
- Enabled per pool, with a global default (decided 2026-11-01):

  ```yaml
  telemetry:
    conversation_export:
      default: false
  pools:
    worker:
      conversation_export: true   # overrides the default
  ```

  The controller resolves the setting at spawn and writes it into the
  session spec, so the host does not read config. Changing it applies to
  new sessions.
- Emitted by the session host, which already sees every ACP frame. One
  trace per session, spans per turn and per tool call, log records for
  prompts, assistant messages, provider-exposed reasoning, tool calls and
  results, permission requests and replies, lifecycle.
- Records carry ticket, project, pool, agent, model and ACP session id, so
  a query can pull the full context of how a ticket was worked.
- Keep convtel's safety properties: separate explicit opt-in from
  operational telemetry, redaction before enqueue (secret-named keys,
  launch-env secret values, and credential patterns via gitleaks' detector
  and default rules), per-field size cap with a
  truncation marker, bounded non-blocking queues with drop counters, export
  failures never affect a session, stable event ids for dedupe.
- Destination: user-configured OTLP endpoints, not Jobshop's concern which
  backend stores them (decided 2026-11-01). No loopback restriction.
- The host's journal remains the record of truth; export is best effort.

**Exporter configuration**
- Per-signal full endpoint URLs (traces, logs, metrics), headers, and
  protocol (`http/protobuf` default, `grpc` optional). Per-signal URLs are
  required because some backends use non-standard paths (VictoriaMetrics
  stack below); the Go SDK's `WithEndpointURL` supports this.
- Standard `OTEL_EXPORTER_OTLP_*` env vars are honored as fallbacks.
- Prometheus `/metrics` is always available for scraping regardless.

```yaml
telemetry:
  otlp:
    traces:  { endpoint: http://127.0.0.1:10428/insert/opentelemetry/v1/traces }
    logs:    { endpoint: http://127.0.0.1:9428/insert/opentelemetry/v1/logs,
               headers: { VL-Stream-Fields: "service.name,jobshop.pool" } }
    metrics: { endpoint: http://127.0.0.1:8428/opentelemetry/v1/metrics }
  prometheus:
    listen: 127.0.0.1:9464
  conversation_export:
    default: false
```

**Our deployment (gasserver)**: VictoriaLogs for logs and conversation
export, VictoriaTraces for traces, VictoriaMetrics for metrics (scrape
`/metrics` or OTLP push). All three accept OTLP/HTTP directly, so no
collector is required. VictoriaTraces is documented upstream as not yet
production-ready; acceptable for our own use. Exact paths to verify
against the installed versions.

### Open questions
- None currently. Backend choice is deployment config, not design.

## 10. Chat gateway (future, design must not preclude)

**Not a priority and not scheduled in any phase** (decided 2026-11-01).
This section exists so the earlier design keeps it implementable. Only the
"What this requires now" items below affect current work; everything else
here is a sketch to revisit when the gateway is actually built.

A human-facing way into the system, modeled on Hermes Agent's messaging
gateway: one long-lived session per chat conversation (room, DM or thread).

- `jobshop gateway` is a separate process and an ordinary client of the
  controller's admin socket. No gateway code in the controller.
- Platform adapters (Matrix first, given existing encrypted relays) map a
  conversation to a session key, e.g. `matrix:<room>`.
- Chat sessions belong to a pool that is not ticket-driven: the controller
  keeps one session per active key, starts it on the first message,
  idle-kills it, and resumes it later via `session/load` / the host
  journal.
- The gateway sends "deliver message to session key X" and streams
  updates back. The controller stays the only ACP client of each host.
- Permission prompts are relayed to the chat; the host already holds them
  until answered.
- A chat session's agent uses the same `jobshop` CLI via its `cli.sock`,
  with scope from its pool config. A coordinator pool can assign tickets,
  read status, and send messages (mail) to worker sessions. This is how
  the mayor role comes back: as config, from chat.
- Per-adapter user allowlists; YAML maps rooms to pools/agents.

What this requires now, so it can be added later without redesign:
1. The admin socket API covers: list sessions, start a session by key,
   send input to a session, subscribe to its updates, answer permission
   prompts. The gateway, the CLI and any future dashboard all use it.
2. Pools have a type (ticket-driven vs keyed/on-demand) from the start.
3. Messages are store-backed and addressed to sessions, pools or keys, so
   a chat session can mail a worker and a worker's reply can surface in
   chat.

Message delivery to a busy session (decided 2026-11-01):
- Default: queue. The message is delivered at the start of the
  recipient's next turn (or by the host as a new prompt once the current
  turn ends).
- Course correction: a sender may mark a message `--interrupt`. The
  controller sends ACP `session/cancel` for the in-flight turn, waits for
  the turn to end (falling back to SIGINT after a timeout, as Gas City
  does), then sends a new prompt carrying the correction plus a short note
  that the previous turn was interrupted. Work state is unaffected: the
  ticket stays claimed, and the agent continues from the correction.
- Interrupt is a permission in pool config: humans and coordinator pools
  may interrupt; workers may only queue by default.
- Chat: a plain chat message to a busy session queues; a marked message
  (e.g. a `/steer` or `!` prefix) interrupts.

- A message to an idle worker wakes it immediately (decided 2026-11-01).

Decided 2026-11-01: one session per room or thread, shared by everyone in
it.

## 11. Risks

- Rebuilding session supervision and claim recovery means rediscovering edge
  cases Gas City already hit. Read its git history and incident notes
  (`engdocs/contributors/reconciler-debugging.md`, `health-patrol.md`) before
  writing the controller.
- Scope creep toward Gas City parity.
- Self-hosting failure: a bad change to the controller can stop the agents
  that would fix it. Needs a known-good binary to fall back to.

## 12. Alternatives considered

- **Keep extending `sqlite-work` in the Gas City fork.** Lowest short-term
  cost. Merge drag grows, and the packs still assume `bd`'s CLI.
- **Hard fork and prune Gas City.** Delete `bd`/Dolt, make sqlite the only
  store, make `gc bd` native. Keeps the controller's edge-case handling.
  Medium one-time cost.
- **Fork `bd` with a SQLite backend.** `bd` has a backend plug-in point
  (`internal/storage/backends`), but the `DoltStorage` interface is large
  (about 100 methods across 15 interfaces) and the SQL is Dolt-flavored.
- **`beads_rust` (`br`).** Maintained, SQLite + JSONL, but follows classic
  beads: no molecules, different JSON shapes.
- **Gas City `file` provider.** Same CLI gap as `sqlite-work`, and documented
  as for local trials only.

## 13. References

Gas City (fork checkout `~/repo/github.com/elee1766/gascity`):
- `AGENTS.md`: design principles, primitives, layering invariants.
- `engdocs/architecture/nine-concepts.md`, `controller.md`, `dispatch.md`,
  `health-patrol.md`, `glossary.md`.
- `docs/reference/specs/formula-spec-v2.md`: formulas, control tickets, gates.
- `docs/reference/sqlite-work-provider.md`: our SQLite provider.
- `internal/beads/beads.go`: `Store` interface and `Bead` type.
- `internal/beadmeta`: metadata key conventions.

Live city (gasserver, read via `/home/a/gfx/scripts/as-agent.sh`):
- `/home/agent/gc/city.toml`, `prompts/`, `formulas/`, `ops/`.
