# JOB-175 acceptance audit of the package restructuring

2026-11-09. Audit of JOB-164 (JOB-165 to JOB-174) at master 15eb26d5.

## Result

Accepted. One stray directory removed here, one pre-existing bug filed (#1128).

## Checked

- Graph: 67 packages under 13 top-level families, 0 known archcheck
  violations. The sibling notes JOB-165 to JOB-174 match the tree.
- Enforcement: real prohibited imports fail `go run ./pkg/tooling/archcheck/cmd/archcheck`
  (`pack` importing `store`, `config` importing `cuelang.org/go/cue`). The
  deliberate-edge test covers the other rules.
- Gates: `go build`, `go vet`, `go test ./...`, `golangci-lint` clean.
  `go test -race` passes on pack, config, orchestration and transport.
- Storage: the SQLite suites pass. A disposable PostgreSQL 17 passes the
  postgres, storemigrate and journal suites (81 tests, 0 skipped). Without
  `JOBSHOP_TEST_POSTGRES` those tests skip, so CI-less local runs do not
  prove the Postgres side.
- Restart compatibility: the controller restart, slot, ingest and hook
  recovery tests pass. A controller built at v0.9.22 (before the
  restructuring) started a run with a live turn, and the new binary adopted
  the same host process with no respawn. The run finished on the original
  host. A last-good snapshot written by the old binary loads in the new tree
  and is served by `control.Run` when the pack is broken.

## Found

- `pkg/desk/dist` was committed again by #1126, next to the real, ignored
  `pkg/transport/desk/dist`. Nothing reads it. Removed.
- `jobshop serve` refuses to start on a failing pack even with a last-good
  snapshot, because `startTelemetry` loads config strictly before the
  controller runs. Same on v0.9.22, so not a regression. Filed as #1128.

## What `pkg/orchestration/control` still owns

99 non-test files, 43,933 lines (counted from the tree; the file groups
below cover each file once). Nothing imports it except `cmd/jobshop` and
test harnesses (`control-is-top`).

| Group | Files | Lines | Why it stays |
|---|---|---|---|
| Wiring and lifecycle | `control reload watch lifecycle tracing systemstatus offpass reconcile pollers drain` | 1,980 | The struct, `Run`, the single `tickMu` supervisor pass, live config swapped as one value, shutdown. This is the controller. |
| RPC authorization and read shaping | `methods identity artifact deskexec report ticketgraph tickettrace tickettimeline incidents graphwait events subscribe journal journalsearch recall` | 4,133 | Grants and project scope are decided here. Read shaping builds `rpc` results from store rows. |
| Durable transitions | `session sessionclose sessionconfig sessionrename sessionstate sessionuid runs batch decision retrypool requestchanges escalate quarantine orphaned placement scale resources revert provider rebasecheck wait waitwake upsert runtitle githubref autostep scriptstep` | 7,321 | Graph, claim and revision fences, one transaction per transition. **Mixed.** |
| Host adapters | `activityhost browsehost hookhost hookcommands hookaudit hooks hook hookactions hookresume ingestqueuehost journalcompresshost journalshiphost linearhost triggershost ingestdrain ingestwork dispatch inboxpoll intake admission synchandler hostproc sandbox isolation isolatedcmd mappedhome preflight apiclients ghcall outbound writekey redactwrite` | 4,528 | Bind an engine in a sibling package to the store, config, clock and grants. `hookhost` and `hookactions` also apply admission policy and commit graph writes. |
| Conversation and learning wiring | `chat chatattachment deskchat attention attentionanswer attentionchat questionchat questionctx escalatechat messages inject learn gate memory promptrules` | 2,925 | Grants and delivery for the engines in `conversation` and `knowledge`. |

Still mixed, with no pending child:

- Durable transitions. `runs.go`, `batch.go` and `decision.go` interleave pure
  graph checks with store writes under `runMu` and the claim fence.
  JOB-168 left them because a split would put one transaction across two
  packages. `session.go` (1,872 lines), `methods.go` and `control.go` (about
  1,000 each) are the largest files. This was a recorded decision in JOB-168,
  not a user descope, and no child issue exists for it.
- The hook host adapter (`hookhost.go`, `hookactions.go`): policy and atomic
  commit, same reason.

## Old-to-new package map

Generated from `git diff -M30%` between 0.9.22 (a5f450bd, after JOB-165) and
this tree, non-test Go files, counts in brackets. Every one of the 68 old
packages is listed. Packages new since 0.9.22 with no old counterpart:
`pkg/domain` (7 files from `pkg/store`), `pkg/orchestration/browse` (added by
JOB-163) and `pkg/integrations/internal/transport` (JOB-169), and the new
files `pkg/pack/{assemble,merge,pack,prefix,watch}.go` and
`pkg/config/{scriptentries,template}.go` (JOB-172).

| Old | New |
|---|---|
| `pkg/apicall` | `pkg/integrations/apicall` |
| `pkg/github`, `pkg/linear`, `pkg/identity` | `pkg/integrations/{github,linear,identity}` |
| `pkg/automation` and `/api /command /github /http /jev /linear /tickets` | `pkg/execution/automation` and the same subpackages |
| `pkg/control/sessionrt` | `pkg/execution/session` (9), `pkg/execution/sandbox` (5) |
| `pkg/hostspec` | `pkg/execution/session` (merged into `spec.go`) |
| `pkg/cli` | `pkg/transport/cli` |
| `pkg/desk`, `pkg/gateway` | `pkg/transport/{desk,gateway}` |
| `pkg/rpc`, `pkg/rpc/rpcschema`, `.../cmd/deskgen` | `pkg/transport/rpc` and the same subpackages |
| `pkg/clock`, `pkg/logctx`, `pkg/redact`, `pkg/telemetry` | `pkg/support/<same>` |
| `pkg/flaky`, `pkg/flaky/cmd/ci-flaky`, `pkg/runbook` | `pkg/tooling/{flaky,flaky/cmd/ci-flaky,runbook}` |
| `pkg/tooling/archcheck` and `/cmd/archcheck`, `pkg/ids` | unchanged |
| `pkg/config` | `pkg/config` (stays) and `pkg/pack` (5 files: pack loading) |
| `pkg/config/tmpl`, `pkg/packcue` | `pkg/pack/tmpl`, `pkg/pack/cue` |
| `pkg/hooks` | `pkg/orchestration/hooks` (15), `pkg/pack/cue` (1, `schema.go`) |
| `pkg/cron` | `pkg/config/cron` |
| `pkg/control` | `pkg/orchestration/control` (96, plus 1 deleted and re-added as `ingestqueuehost.go`), `pkg/conversation` (2) |
| `pkg/control/controltest` | `pkg/orchestration/control/controltest` |
| `pkg/control/waitengine` | `pkg/orchestration/workflow` |
| `pkg/control/{activity,linearsvc,scaleplan,slotprovider,stepexec,threads,triggers}` | `pkg/orchestration/<same>` |
| `pkg/ingest`, `pkg/control/ingestqueue` | `pkg/orchestration/ingest`, `pkg/orchestration/ingest/queue` |
| `pkg/control/chatlife` | `pkg/conversation` (2: `conversation.go`, `escalation.go`) |
| `pkg/control/attention`, `pkg/control/question` | `pkg/conversation/{attention,question}` |
| `pkg/learning`, `pkg/memedit` | `pkg/knowledge` (1 and 2 files) |
| `pkg/memory` | `pkg/knowledge/memory` |
| `pkg/store` | `pkg/storage/store` (11), `pkg/domain` (7) |
| `pkg/store/{sqlite,sqlite/sqldb,postgres,postgres/pgdb,storemigrate,storerules,storetest}` | `pkg/storage/store/<same>` |
| `pkg/writekey` | `pkg/storage/store/writekey` |
| `pkg/journal`, `pkg/journal/vltest` | `pkg/storage/journal`, `pkg/storage/journal/vltest` |
| `pkg/control/{journalread,journalship,journalcompress,journalvl}` | `pkg/storage/journal` (1), `.../ship`, `.../compress`, `.../journalvl` |

Merged: `hostspec` into `session`, `chatlife` and two control files into
`conversation`, `learning` and `memedit` into `knowledge`, `packcue` and
`hooks/schema.go` into `pack/cue`. Removed: none outright. `pkg/control` has no
files left (the directory is gone).

## Scope

All ten sibling issues (JOB-165 to JOB-174) are Done. No scope is descoped by
the user. The only pending items are the two mixed areas above, which have no
child issue, and #1128 for the serve path.

## Testing restarts

The host environment hash includes volatile shell variables such as `_`. Pin
the environment of both binaries, or a restart test shows a respawn that is
not a regression.
