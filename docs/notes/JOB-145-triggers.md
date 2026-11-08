# Trigger and scan evaluation moved out of pkg/control (JOB-145)

2026-11-08. No behavior change. Part of the JOB-139 extractions, after
journal shipping (JOB-119).

## What moved

`pkg/control/triggers` now owns:

- `Engine.Eval`: dedup, cooldown-from-close, cron, counts, learnings,
  edits and check conditions of a poll trigger, and `Info` for
  `trigger.list`.
- Scan evaluation: `Detect`, `evalScan` (max_open, max_runs, cooldown, the
  #704 subject hold), the signals read from the ticket store (ready,
  step_idle, session_silent, needs_human, blocked, check, script) and the
  script signal cache.
- Event routing: `EventOf`, `On`, `Events`, `OnEvent` (webhook rows to
  triggers, `fallback:` poll spacing in `pollDue`).
- Off-tick checks (JOB-109, JOB-103): `Pass`, `OffTick`, `RunChecked`,
  the per-trigger check jobs, the 8-slot bound and the replay rounds.

Runtime state moved with it: last poll per trigger, check jobs, check
slots and the script signal cache. Everything durable is unchanged: it is
`meta.trigger` (name, key, finding) on the tickets and runs a trigger
filed, and the hook invocation keys, so a restart or a replay finds the
same state and files nothing twice.

## What stays in pkg/control

The controller is still the only writer. `Host` (implemented by
`triggerHost` in `triggershost.go`) gives the engine config, clock, store
reads, the tick lock, spans, the callback queue (`pkg/hooks` `Put` and
`LastFire`, through the small `Callbacks` interface) and the check
runner. The writes stay behind it: `Fire` (ticket, run or command),
`FireFinding` (run start with title read and key), `RecordFinding` (the
learning). Trigger commands, `startRun` and the batch commits are
unchanged.

The signals over state other subsystems own are controller methods the
engine calls through `Signals`: pool quarantine, resource waits, ingest
backlog and worker, Linear issues (with their poll cache, still in
`linearintake.go`), and the signal script call. They return
`triggers.Finding`. The intake, Linear issue and run title reads use
`triggers.OffTick`, so they still run outside the pass.

## Checks that keep the old guarantees

- `TestPassReadsConfigOnce` lists the controller methods the engine calls
  as pass roots. `TestPassConfigInEngine` checks the engine's own code:
  only `Events` and `checkLoop` read the host's live config.
- Existing scan, prefix, event, off-tick, cooldown, cap and restart tests
  in `pkg/control` run through the engine unchanged (names aside).
- `TestOffTickTimeout` still sees the `poller=check` log and metric.

## What remains

`hookactions.go` and `hookhost.go` still call `Eval` and `OpenCount` for
handler triggers. The Linear and GitHub signal readers move with their own
extractions. `fireCommand` and `finishCommand` stay: they write tickets
and run isolated commands the controller owns.
