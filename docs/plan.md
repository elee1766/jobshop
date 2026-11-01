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

### Open questions
- Which of these do we actually need on day one versus later?
- Do the observer agents (deacon/boot/witness) still earn their cost, or can
  the controller plus events cover them?
- Is `gx` part of Jobshop, or does it stay a separate client?

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
| Event log | Events | |

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

### Open questions
- Should the store expose a generic graph (nodes, typed edges, metadata), or
  first-class types (ticket, message, session, schedule)? Gas City stores
  everything as beads plus about 200 `gc.*` metadata keys
  (`internal/beadmeta`). That is flexible but implicit.
- How strict do we want the schema to be versus a metadata map?

## 5. Ticket store interface (sketch)

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
- **Ephemeral tier** for ephemeral runs.
- **Typed edges**: `blocks`, `parent-child`, `tracks`, `waits-for` cover nearly
  all of Gas City's usage.

A conformance test suite that every backend must pass, written before the
SQLite implementation.

### Open questions
- Should classes (work, graph, sessions, messaging, ...) be separate tables,
  separate stores, or a column?
- Do we need Gas City's "split city" (class bindings across stores) at all?
  It accounts for a lot of complexity (`cmd/gc/bd_split_*`).
- Multi-project: one database or one per project? We currently run one
  work store per rig plus a shared infra store.
- Is `modernc.org/sqlite` (pure Go, what Gas City uses) acceptable, or do we
  want cgo?

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

### Open questions
- Copy and adapt, or vendor at a pinned commit? Copying loses history but
  allows trimming.
- How much of `internal/dispatch` do we use? Our formulas may not need
  fan-out, tally or drain at all.

## 7. Phases (rough)

Ordering is a proposal. Each phase ends with something runnable. The ordering
is driven by dogfooding: get to the smallest loop that can run an agent on a
Jobshop ticket in this repo, then let that loop do more of the remaining work.
Until then, development happens by hand (or with an outside agent).

**Phase 0: groundwork**
- Repo layout, CI, lint. Build is plain `go` (`go build`, `go test`), no
  Bazel (decided 2026-11-01). CI on GitHub Actions: build, `go vet`,
  `golangci-lint`, `go test -race`.
- Write down the ticket data model: fields, statuses, edge types, metadata
  conventions.

**Phase 1: ticket store**
- `store` package: interface, types, errors.
- Conformance suite.
- SQLite backend (adapt from Gas City `sqlite_*` where it fits).
- Minimal CLI: `jobshop ticket create/show/list/update/close`, `claim`,
  `heartbeat`, `release`, `ready`, with stable JSON output.

**Phase 2: workflows and runs**
- Port `internal/formula`. Decide which workflow features to keep.
- Run creation via graph apply.
- Port the subset of `internal/dispatch` our workflows need.

**Phase 3: controller and sessions**
- Config loader (project/agent/pool) in YAML.
- Runtime provider: ACP first (that's all we use), tmux maybe later.
- Controller tick: pool demand from queues, start and drain sessions,
  restart, quarantine, idle-kill.
- `jobshop assign`, and the hook: `jobshop claim --next` (three-tier query:
  my in-progress, my pre-assigned, pool queue).

**Phase 4: the rest of our surface**
- Messages (mail), batches (convoys), schedules (orders).
- Events, and an API if the dashboard is to be ported.

**Dogfood milestone (target, may land between phases 1 and 3)**
- Smallest loop that works on itself: one project (this repo), one pool of
  one agent, `assign`, `claim --next`, `heartbeat`, `close`, and a controller
  that starts a session when the queue is non-empty and restarts it if it
  dies.
- No workflows, dispatcher, messages or batches required. Agents close
  tickets directly.
- After this, new features are filed as Jobshop tickets and built by Jobshop
  agents, with a human reviewing.

**Self-hosting setup (proposed 2026-11-01)**
- Host: `gasserver`, next to the live Gas City. Work happens over SSH.
- GitHub identity: the `cog` user, with write access to this repo to push
  branches, open PRs and merge.
- Isolation from Gas City: own data dir and SQLite file, own user systemd
  unit, own session cap so Jobshop cannot starve the live city's 20 slots.
  Separate Linux user (`jobshop`) vs reusing `agent` is open.
- Known-good binary: the controller runs a pinned build from a release path
  (e.g. `~/.local/share/jobshop/bin/jobshop-<sha>`), never the working tree.
  Promotion is an explicit step, and rollback is repointing a symlink.
- Agents work in git worktrees, one per ticket, and deliver via PR.
- Merge policy: branch protection requires CI. `cog` may merge on green
  for normal changes. Changes under the controller, store schema or
  migrations need a human approval first.

**Phase 5: hardening from use**
- Add what dogfooding shows is missing, in the order it hurts.
- Bring over Gas City features only when a real need appears.

**Later**: Postgres backend, against the same conformance suite.

### Open questions
- Is Phase 3 too large for one phase?
- Should the dogfood milestone come before workflows (phase 2)? Probably
  yes, since a flat ticket queue is enough to start.
- What safety rails does self-hosting need (agents editing the controller
  that runs them; review before merging; a way to roll back a broken build)?
- Which runtime does the dogfood loop use: ACP via jcode as in our city, or
  tmux as the simpler first provider?

## 8. Risks

- Rebuilding session supervision and claim recovery means rediscovering edge
  cases Gas City already hit. Read its git history and incident notes
  (`engdocs/contributors/reconciler-debugging.md`, `health-patrol.md`) before
  writing the controller.
- Scope creep toward Gas City parity.
- Self-hosting failure: a bad change to the controller can stop the agents
  that would fix it. Needs a known-good binary to fall back to.

## 9. Alternatives considered

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

## 10. References

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
