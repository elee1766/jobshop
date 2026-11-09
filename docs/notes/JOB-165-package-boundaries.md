# JOB-165: package ownership map and enforced import boundaries

2026-11-09. First child of JOB-164. Baseline is master at e8f723bd, after
JOB-140..149 and JOB-152..159. Nothing here moves a package. It records where
each one goes, which child moves it, and adds the check that keeps the graph
honest while they do.

## What was built

`pkg/tooling/archcheck` reads `go list -e -deps -json` for the whole module
(with `-tags victorialogs`, so tagged tests are in the graph) and checks it
against `jobshop.go` (ownership table, family layers) and `rules.go` (named
rules). `cmd/archcheck` runs in CI right after `go vet`, and `TestRepository`
runs the same check in `go test`.

```sh
go run ./pkg/tooling/archcheck/cmd/archcheck             # check
go run ./pkg/tooling/archcheck/cmd/archcheck -inventory  # every package: owner, prod, test-only, external edges
go run ./pkg/tooling/archcheck/cmd/archcheck -why ./pkg/hooks github.com/elee1766/acphost/pkg/sandbox
```

Production, test-only and closure edges are kept apart. A rule picks one:
`Prod`, `ProdAndTest`, `TestOnly`, or `Closure` (everything the production
code reaches, with the shortest chain printed). Test harnesses (`storetest`,
`controltest`, `vltest`, `journalvl`) are marked in the table. Tests may import
them, production code may not, and they are not checked as importers.

The check fails when: a package has no row in the table, a row matches no
package, a rule pattern matches nothing (a rename cannot leave a dead rule), the
build is broken, or a listed exception is stale. Each exception names one rule,
one source, one target, a child issue and a reason. There is no wildcard
exception.

## Rules enforced now

| Rule | Looks at | Says |
|---|---|---|
| `control-is-top` | prod and test | nothing outside `pkg/control` and the harnesses imports `pkg/control` |
| `no-harness-in-production` | prod | production code never imports a test harness |
| `command-paths-no-session` | closure | hooks, automation and stepexec do not reach `control`, `sessionrt`, `chatlife`, `hostspec` or `acphost.acphost` |
| `command-paths-no-acp` | closure | the same packages do not reach the ACP SDK |
| `lower-layers-no-session-runtime` | closure | store, knowledge, config, integrations, support and automation do not reach acphost or the ACP SDK |
| `starlark-only-in-automation`, `cue-only-in-packcue` | prod and test | one owner per language runtime |
| `sqlite-driver-owners`, `postgres-driver-owners` | prod and test | only the backends (and the ingest spool) open a database |
| `store-backends-are-siblings` (both ways) | prod and test | sqlite and postgres do not import each other |
| `independent-leaves` | prod and test | clock, cron, logctx, redact, tmpl, memory, packcue, ids, apicall, hostspec import no other package of the module |
| `family-<name>` | prod | a family imports only the families listed for it in `families()` |

The family layers are the JOB-164 layout, expressed over today's paths. Rules
for packages that do not exist yet (`domain`, `pack`, `storage`) arrive with the
child that creates them. Domain independence is an `Only` rule (`From: domain`,
std only) that JOB-166 adds with the package.

## Baseline exceptions (9 entries, 7 left after JOB-171, each owned by a child)

| Edge | Removed by |
|---|---|
| `cli -> control` (cli hosts `control.Run`) | JOB-174 |
| `desk` tests `-> control` (build `control.Options`) | JOB-174 |
| `stepexec -> sessionrt` (`GitCommonDir`), and the ACP SDK reached through it | JOB-167 |
| `hooks -> rpc -> acp-go-sdk` (guard decisions, automation result, ID params) | JOB-166 |
| `hooks -> cuelang.org/go` (`schema.go`) | JOB-172 |
| `journalread -> hostspec` (`Env` reads spec.json) | JOB-173 |
| `chatlife -> attention`, `chatlife -> question` | JOB-171 (removed: they moved under `conversation` as its own leaves) |

Findings that are not edges to remove, kept here so no child rediscovers them:
`config -> automation` (`hookcheck.go` compiles hook programs at load) is
allowed for the `config` family and JOB-172 decides whether validation stays in
config. `config -> store` is types only (`ScopeRef`, `RunScope`,
`ValidKindPattern`, `Learning*`, `DefaultReviewRounds`) and JOB-166 moves what
domain needs. `rpc` imports acphost `cgroup`, `sandbox`, `userns` only for
wire fields on the session spec, and JOB-166/167 decide whether those stay in
`rpc` or become domain-neutral types.

## Helper destinations (decided from callers, not names)

| Package | Callers (production) | Destination |
|---|---|---|
| `ids` | config, control, both store backends | stays a leaf generator (its only dependency is sonyflake, which domain must not take). JOB-166 moves identifier kinds and checks (`IsSession`, session tag) that domain needs |
| `cron` | config (validation), `control/triggers` | stays a leaf. Lives with config (`config/cron`), since config parses schedules at load and triggers is the only evaluator |
| `writekey` | control only (ingest dispatch, Linear write keys, `writekey.claim/finish` RPC) | `storage/store/writekey`. It is a durable store collection used by more than one subsystem, not provider-only bookkeeping |
| `hostspec` | cli, control, controltest, journalread | `execution/session`. It is the session's spec file, on acphost. Must stay import-free of the module |
| `apicall` | github, linear, automation/github, automation/linear, cli, linearsvc | `integrations/apicall`, not `internal`: automation and cli sit outside the integrations family, and Go internal rules would force a wider export |
| `identity` | control only | `integrations/identity` |
| `logctx` | cli, control, telemetry | `support/logctx`. Three callers in two families and telemetry needs it, so it earns a separate leaf |
| `clock` | control, ingestqueue, gateway, controltest | `support/clock` |
| `redact` | config, control, journal, journalship | `support/redact` |
| `packcue` | config | `pack/cue` (JOB-172) |
| `learning`, `memedit` | control, triggers, storetest | `knowledge`. `memory` stays independent of both (no store import) |
| `runbook`, `flaky` | cli, `ci-flaky` | `tooling` |

There is no evidence yet of provider HTTP code worth sharing between `github`
and `linear` beyond `apicall`. JOB-169 compares them before creating
`integrations/internal/transport`, and creates it only if both use it.

## Current-to-target ownership

The authoritative table is `owners()` in `jobshop.go`, one row per package with
Go files (65), checked against the live graph. Summary by family and child:

| Family | Packages today | Target | Child |
|---|---|---|---|
| shared types | none yet | `domain` | JOB-166 |
| execution | `sessionrt` (host liveness, processes, shell probe, cgroup, ACLs, mounts, mapped homes, slots, sandbox spec, command, ACP client), `hostspec`, `automation/...` | `execution/session` (liveness, connect, client, slots), `execution/sandbox` (cgroup, ACL, mounts, sandbox, command, `GitCommonDir`), `execution/automation/...` | JOB-167 |
| orchestration | `control`, its subpackages `activity attention question ingestqueue linearsvc scaleplan slotprovider stepexec threads triggers waitengine`, `hooks`, `ingest` | `orchestration/{control,workflow,hooks,ingest}`. Policy and coordination stay in control | JOB-168 |
| integrations | `apicall github linear identity` | `integrations/...` | JOB-169 |
| knowledge | `learning memedit memory` | `knowledge`, `knowledge/memory` | JOB-170 |
| conversation | `control/chatlife` and the chat, message and attachment files of control | `conversation` | JOB-171 |
| pack, config | `packcue`; `config config/tmpl cron` | `pack/cue`; `config`, `config/tmpl`, `config/cron` | JOB-172 |
| storage | `store/...`, `writekey`, `journal`, `control/journal{ship,compress,read,vl}` | `storage/store/...`, `storage/journal/{ship,compress,read}` | JOB-173 |
| transport, support, tooling | `rpc rpc/rpcschema cli desk gateway`; `telemetry clock redact logctx`; `runbook flaky` | `transport/...`, `support/...`, `tooling/...` | JOB-174 |

Leftover implementation still in `pkg/control` (98 files, 82 write call
sites, 23 mutex fields, 4 goroutine launches), with the destination each child
confirms from its own caller trace:

| Files in control | Concern | Likely owner |
|---|---|---|
| `chat.go deskchat.go escalatechat.go attentionchat.go questionchat.go chatattachment.go messages.go messagesxml.go inject.go` | conversation mechanics | JOB-171 (routing policy and claims stay in control) |
| `learn.go gate.go memory.go recall.go` | learnings, curator gate, local memory writes | JOB-170 |
| `apiclients.go identity.go githubref.go ghcall.go linearhost.go` | provider clients and identities | JOB-169 (host adapters stay) |
| `isolation.go isolatedcmd.go sandbox.go mappedhome.go sessionuid.go hostproc.go hookcommands.go deskexec.go` | process and isolation use | JOB-167 |
| `journal.go journalsearch.go journalshiphost.go journalcompresshost.go writekey.go redactwrite.go` | journal and write-key hosts | JOB-173 (hosts stay thin) |
| `hook*.go hooks.go admission.go intake.go` | hook host adapter, admission policy | stays (JOB-148: the adapter and admission are policy). JOB-168 narrows its command API |
| everything else (`control.go methods.go session.go runs.go scale.go placement.go escalate.go quarantine.go decision.go report.go ...`) | policy, coordination, durable transitions, the supervisor pass | stays |

State to preserve while moving (unchanged by this PR): one `tickMu` supervisor
pass with off-pass calls (JOB-103), runner goroutine per session joined by
`r.done` and `c.wg`, `liveConfig` swapped as one value, store transactions for
claims and graph writes, the ingest spool as its own SQLite file, and the
journal `.zst` files read through acphost readers.

## Sequencing

JOB-165 (this) blocks JOB-166 and JOB-175. JOB-166 blocks 167 to 174. JOB-168
also waits for JOB-167's narrow command API. After JOB-166 the rest can run in
parallel, with three real prerequisites between them:

1. JOB-167 must finish `stepexec -> sessionrt` (`GitCommonDir` moves to the
   sandbox side) before JOB-168 moves `stepexec` under orchestration.
2. JOB-173 moves `journalread`'s env source out before moving it, which needs
   JOB-167 to settle where `hostspec` lives.
3. JOB-172 decides `config -> automation` (hook program validation) before JOB-167
   moves `automation`, or JOB-167 moves it and updates the family row.

Each child updates its rows in `jobshop.go`, deletes its exceptions from
`known.go` (a stale entry fails CI), and adds the rules its move makes possible.

## Validation

- Real prohibited edges added to source files, each caught by the real CLI and
  removed afterwards: `store -> control` (go list reports the import cycle, so
  the check refuses to pass on a broken build), `hooks -> hostspec` (and
  `stepexec` through it), a `journal` test file `-> control`, and `gateway`
  production `-> controltest`.
- `TestRepositoryCatchesDeliberateEdges` repeats eleven such edges against the
  loaded real graph. `TestRepositoryLegitimateTestHarnessEdgesPass` pins that
  harness imports from tests are allowed.
- `archcheck_test.go` builds temporary modules and runs real `go list` on them.
