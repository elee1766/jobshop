# JOB-168: orchestration family (2026-11-09)

Child of JOB-164. Moves packages under `pkg/orchestration`, no behavior change.
The controller keeps the same files, tables, RPC and transactions.

## Layout

| Before | After |
|---|---|
| `pkg/control` | `pkg/orchestration/control` (and `controltest`) |
| `pkg/control/waitengine` | `pkg/orchestration/workflow` |
| `pkg/hooks` | `pkg/orchestration/hooks` |
| `pkg/ingest` | `pkg/orchestration/ingest` |
| `pkg/control/ingestqueue` | `pkg/orchestration/ingest/queue` |
| `pkg/control/{triggers,stepexec,scaleplan,slotprovider,threads,activity,linearsvc}` | `pkg/orchestration/<same name>` |
| `pkg/browse` | `pkg/orchestration/browse` |

Only the two renames change a package name: `waitengine` is `workflow` and
`ingestqueue` is `queue`. Every other package keeps its name. `workflow` is
the self-contained wait and wake planning that exists today. No graph or
transition code moved into it (see "Not moved").

## Controller role table

`pkg/orchestration/control` is 99 non-test files, about 20,900 lines. Roles
after the earlier extractions (JOB-141 to JOB-159, audit in
`JOB-155-control-audit.md`):

| Role | Files | What stays and why |
|---|---|---|
| Wiring and lifecycle | `control.go reload.go watch.go lifecycle.go tracing.go systemstatus.go offpass.go reconcile.go pollers.go drain.go` | struct, `Run`, the single `tickMu` supervisor pass, live config swapped as one value, shutdown |
| Authorization and RPC | `methods.go identity.go artifact.go deskexec.go report.go ticketgraph.go tickettrace.go tickettimeline.go incidents.go` | grants, project scope, read shaping |
| Durable transitions | `session*.go runs.go batch.go decision.go retrypool.go requestchanges.go escalate.go quarantine.go orphaned.go placement.go scale.go resources.go revert.go` | graph, claim and revision fences, atomic commits. Not split: one transaction per transition |
| Host adapters | `*host.go ingest*.go admission.go inboxpoll.go dispatch.go journal*.go hostproc.go hook*.go synchandler.go intake.go sandbox.go isolation.go isolatedcmd.go mappedhome.go` | bind an engine to the store, config and clock; no logic of their own |
| Conversation and learning wiring | `chat*.go deskchat.go attention*.go question*.go escalatechat.go messages.go inject.go learn.go gate.go memory.go recall.go promptrules.go` | grants and delivery for engines in `conversation` and `knowledge` |

## Import graph

Taken with `archcheck -inventory`.

- `control` imports the sibling engines under `orchestration`. Nothing imports
  it back: the only production importer is `cmd/jobshop` (JOB-174), and the
  rest are test harnesses (`controltest`, `packs`, `journalvl`).
- Engines import only config, domain, store, rpc, integrations, telemetry and
  each other as a DAG: `ingest/queue -> {hooks, ingest}`, `triggers -> {hooks,
  ingest}`, `threads -> linearsvc`, `linearsvc -> ingest`, `stepexec -> hooks`.
- `hooks` imports `execution/automation` and no `execution/session`, directly
  or through a command adapter. The adapter (`hookcommands.go`) is in control
  and calls `runIsolated`, which uses `execution/sandbox`.
- `stepexec` imports `execution/sandbox` for the run dir path helper and
  nothing session related (JOB-167).

New archcheck rule `orchestration-engines-no-session-or-sandbox-runtime`: the
engines that neither run a command nor supervise a session do not reach
`execution/session` or `execution/sandbox`, directly or through imports. It
does not name the ACP SDK: `rpc` carries it for four of them, and JOB-174
decides the wire types.

## Not moved, and why

- **Graph and transition logic into `workflow`.** `runs.go`, `batch.go` and
  `decision.go` interleave pure checks with store writes under `runMu` and
  the claim fence (25, 13 and 4 uses of `c.store`, `c.mu` or `c.runMu`). Extracting only the pure half would put
  one transition across two packages, which the issue forbids ("no split
  transactions for packaging convenience"). They stay until a transition can
  be moved whole. `graphwait.go` (wait edges for the ticket graph) is pure,
  but it shapes `rpc` graph results for a read handler, so it is read
  shaping and stays with the other graph handlers; moving it gains no
  boundary.
- **`hooks` host adapter.** It maps wire envelopes, applies admission policy
  and commits graph writes. Policy and atomic commit are the controller's.
- **Provider clients, journal shipping, isolation, conversation engines.**
  Already out of control (JOB-169, 173, 167, 171).

## Checked

- Same tests, same SQL, same RPC and CLI output: only import paths, the two
  package names and doc paths changed.
- `docs/layout.md`, CI (`go test -c ./pkg/orchestration/control`), the e2e
  runner and the runbooks name the new paths.
- The frozen `docs/plan.md` and dated notes keep their old paths on purpose.
