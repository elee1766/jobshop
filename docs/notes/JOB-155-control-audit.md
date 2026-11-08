# pkg/control audit after the JOB-139 extractions (JOB-155, 2026-11-08)

Completion gate of JOB-139. All ten children (JOB-140 to JOB-149) are Done
and master was green. This audits the result against the parent's rule:
the controller holds wiring and coordination, not subsystem engines. The
audit found small leftovers (moved here) and four larger ones (filed).

## Import graph

Taken with `go list -f '{{.Imports}}' ./pkg/...` on master plus this PR.
Jobshop imports only, `pkg/` prefix dropped.

| Package | Imports |
|---|---|
| `control` | automation, automation/api, clock, config, the 11 subpackages below, github, hooks, hostspec, identity, ids, ingest, journal, learning, linear, logctx, memedit, memory, redact, rpc, store, store/sqlite, store/postgres, telemetry, writekey |
| `control/activity` | config, github, linear, rpc |
| `control/attention` | control/question, rpc, store |
| `control/chatlife` | config, control/attention, control/question, rpc, store |
| `control/ingestqueue` | clock, config, hooks, ingest, rpc, store |
| `control/journalcompress` | config, store |
| `control/journalread` | hostspec, journal, rpc, store |
| `control/journalship` | config, journal, redact, rpc, store, telemetry |
| `control/linearsvc` | apicall, config, ingest, linear, store |
| `control/question` | rpc, store |
| `control/scaleplan` | config, store |
| `control/sessionrt` | config |
| `control/triggers` | config, cron, hooks, ingest, learning, rpc, store, telemetry |
| `control/waitengine` | github, linear, store |

Findings:

- No import cycle (the build would fail) and no package imports `pkg/control`
  back. Checked by running `go list -deps` on every package under
  `pkg/control`, `pkg/hooks`, `pkg/automation`, `pkg/github`, `pkg/linear`
  and `pkg/ingest`. The only importers of `pkg/control` are `pkg/cli`,
  `pkg/desk` and `pkg/control/controltest` and `journalvl`, which start a
  controller on purpose and are tests or the binary's entry point.
- The subpackages form a small DAG: `question` <- `attention` <- `chatlife`.
  Everything else is a leaf under `control`. None imports a sibling engine
  except those three.
- `pkg/hooks` and `pkg/automation` do not import `pkg/control` or any
  `control/*` package. `ingestqueue` and `triggers` use `hooks`; `hooks`
  and `config` use `automation`; `control` does in its hook adapters.
- Direction is downward: `cli -> control -> {subpackages, hooks, github,
  linear, store} -> store/rpc/config`.

## Remaining `pkg/control/*.go` (non-test)

97 files, about 22,100 lines (22,500 before this PR). `wiring` means the file adapts a
subpackage or owns authorization, scope, the store write or the tick
lock for a subsystem. `engine` means a self-contained piece that does not
need the controller. Files marked "follow-up" are in the next section.

| Files | Role |
|---|---|
| `control.go`, `reload.go`, `watch.go`, `lifecycle.go`, `tracing.go`, `systemstatus.go`, `offpass.go`, `reconcile.go`, `pollers.go`, `drain.go` | wiring: controller struct, `Run`, tick and sweep, live config and reload, config file watcher, shutdown, spans and tick lock, off-pass jobs, pump goroutines |
| `methods.go`, `identity.go`, `writekey.go`, `upsert.go`, `ghcall.go`, `redactwrite.go`, `artifact.go`, `deskexec.go`, `memory.go`, `learn.go`, `gate.go`, `report.go`, `incidents.go`, `ticketgraph.go`, `tickettrace.go` | RPC handlers: grants, project scope, store reads and writes. `incidents.go` holds a 90 line slog ring buffer; it is controller telemetry and stays |
| `activityhost.go`, `ingestqueuehost.go`, `ingestdrain.go`, `ingestwork.go`, `admission.go`, `inboxpoll.go`, `dispatch.go`, `journal.go`, `journalsearch.go`, `journalshiphost.go`, `journalcompresshost.go`, `linearhost.go`, `triggershost.go`, `hostproc.go` | wiring: host adapters and RPC wrappers for `activity`, `ingestqueue`, `journalread`, `journalship`, `journalcompress`, `linearsvc`, `triggers`, `sessionrt` |
| `wait.go`, `waitwake.go` | wiring for `waitengine`: `wait.add/list/clear`, revision-checked apply, the pump and timers |
| `attention.go`, `attentionanswer.go`, `attentionchat.go`, `escalate.go`, `escalatechat.go`, `deskchat.go`, `question*.go`, `report.go` | wiring for `attention`, `question`, `chatlife`: grants, answer execution, chat sessions |
| `hook.go`, `hooks.go`, `hookhost.go`, `hookactions.go`, `hookresume.go`, `hookcommands.go`, `hookaudit.go`, `synchandler.go`, `intake.go` | wiring for `pkg/hooks` (JOB-148): admission, commit, graph apply under `runMu` |
| `session.go`, `chat.go`, `chatattachment.go`, `messages.go`, `messagesxml.go`, `sessionclose.go`, `sessionconfig.go`, `sessionrename.go`, `sessionstate.go`, `sessionuid.go`, `mappedhome.go`, `isolation.go`, `sandbox.go`, `isolatedcmd.go`, `inject.go`, `promptrules.go`, `preflight.go`, `quarantine.go`, `orphaned.go`, `placement.go`, `scale.go` | session ownership: the runner (ACP connection, turns, durable transitions) and what feeds it. Engine left: JOB-153 (mapped home, slots), JOB-152 (sandbox, isolated commands), JOB-154 (runner). `scale.go` now keeps only state and apply; the decisions are in `scaleplan` |
| `runs.go`, `runtitle.go`, `batch.go`, `retrypool.go`, `requestchanges.go`, `rebasecheck.go`, `resources.go`, `decision.go`, `revert.go`, `apiclients.go`, `githubref.go`, `events.go`, `subscribe.go` | run graph and ticket flow: run and batch writes, decisions, identity clients, event hub. Engine left: `subscribe.go` (JOB-156) and `revert.go` PR reads (JOB-159) |
| `autostep.go`, `scriptstep.go` | engine left, JOB-157: step execution (templating, run dir, command and http attempt) shares one file with claims and leases |
| `provider.go` | engine left, JOB-158: provider call transport and demand state |
| `outbound.go` | two HTTP clients for step and provider calls |

## Done in this PR (each under 300 lines)

- `memedit.Metrics`, `memedit.Decide` and the word-overlap helpers moved
  out of `gate.go` (about 100 lines). They are pure, so they sit with the
  memory edit collection the gate judges. `learn evaluate` still lists the
  steps and learnings, then records the verdict through `memedit.SetResult`.
  Their tests moved too (`TestMetrics`, `TestMetricsCitedAfterMerge`,
  `TestDecide`); `TestCuratorGate*` stay in control.
- `control/scaleplan`: `Desired`, `ForPool`, `CapStarts` and their types,
  out of `scale.go` (about 120 lines). Pure: pool config and session
  values in, a plan out. The controller still reads sessions, holds
  `scaleHist` under `c.mu` and applies cordons and starts.
- `sessionrt.BackgroundJobs` (was `bgjobs.go`, 118 lines): the `/proc`
  scan for jobs an agent left behind. It needs no controller; the runner
  calls it from `holdForJobs`. Added a unit test.

No behavior change. Names, thresholds and log lines are as before.

## Follow-ups filed (children of JOB-139, in Backlog)

- JOB-156: thread subscription polling and comment filters (`subscribe.go`,
  816 lines). Left by JOB-144 and JOB-145 on purpose; this is the largest
  engine still in control.
- JOB-157: command and http step execution (`autostep.go`, `scriptstep.go`,
  about 750 lines).
- JOB-158: provider slot calls (`provider.go`, 288 lines, over the 300 line
  bar once its state and tests move).
- JOB-159: revert and rebase PR reads (`ghPR`, `revertSource`) into
  `pkg/github`.

JOB-152, JOB-153 and JOB-154 (session runtime slices 2 to 4) already exist
under JOB-146 and stay open. They are not in this PR and the parent
should not close until they are done or descoped by the operator.

## Duplicates of existing engines

Checked `pkg/hooks`, `pkg/automation`, `pkg/github` and the Linear client.

- Hooks and automation: no second engine. `hookhost.go` is an adapter;
  `autostep.go` does not copy hook transport (it shares `hooks.RetryAfter`)
  and posts a different body (JOB-148).
- Linear: every request goes through `pkg/linear` clients. `linearsvc` and
  `control` call `Service.Do`/`DoKey`, so there is one backoff. No second
  GraphQL client.
- GitHub: two readers on purpose (gh CLI in `github.CLI`, REST in
  `github.Client`, JOB-143). Real overlap left: control's `ghComment`
  (subscribe.go) is a second comment type beside `github.Comment` and
  `ghPR` is a second PR type beside `github.PRInfo`. Both are covered by
  JOB-156 and JOB-159.
- The one `gh` subprocess in control is `apiclients.go`'s `gh auth token`
  for the host identity; it is not a client.

## Verification

`go build ./...`, `go vet ./...`, `golangci-lint run ./...` and
`go test ./pkg/...` pass, except `pkg/desk`, which needs the built UI
(`pkg/desk/dist`) that the audit checkout did not build. CI builds it.
