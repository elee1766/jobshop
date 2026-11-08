# pkg/control split, step three: ingest queue (JOB-141, 2026-11-08)

Third subsystem out of `pkg/control` (after `JOB-119-control-split.md`
and `JOB-89-journalcompress-split.md`). No behavior change: RPC names,
JSON, holders, leases, timers and log lines are as in
`679-ingest-queue.md` and `443-ingest-workers.md`.

## Boundary

`pkg/control/ingestqueue` owns the queue's coordination:

- `Service`: the open spool (`Spool`, `CloseSpool`), the drain
  (`DrainOnce`, `DrainFile`, `RunDrain`, `WakeDrain`), the inbox change
  notifier (`Appended`, `Advance`, `Changed`), the admission pump's
  schedule (`RunPass`), the `ingest.next` long poll (`Next`), the
  poller writes (`PutPolled`, `PutBatch`) and the in-memory external
  worker heartbeats (`Beat`, `Workers`, `LastBeat`, `Live`).
- `Runner`, `Run`, `Event`, `Result`, `Settle`, `Hooks`, `Lease`: the
  admission hook half. The runner needs no store, so `jobshop ingest
  work` uses it too. `Settle` decides an item's state from a run and a
  human's answer; it is pure.
- `Worker`: the `jobshop ingest work` loop (was `control.IngestWorker`).
- `applyJSONPatch` and `topLevelDiff` (RFC 6902 for mutating hooks).

It reaches the controller through `ingestqueue.Host`: clock, the inbox
store, the configured spool path, `WakeWaits` (waits whose conditions
name a new item) and `Drained` (the Linear consumer). It does not import
`pkg/control`. Inbox and spool mechanics stay in `pkg/ingest`.

## Stays in the controller (adapters)

- `ingestwork.go`: the `ingest.claim`, `ingest.admit` and
  `ingest.workers` RPCs (admin check, claim under `worker:<id>`, commit
  through `finishAdmission`), the missing-worker and backlog attention
  and scan signals, the inbox gauges.
- `admission.go`: `admitNext`, `finishAdmission` (state write, hold
  ticket, human answer), `releaseAnswered`, `holdForHuman`.
- `ingestdrain.go`: the `ingest.drain`, `ingest.notify` and `ingest.next`
  handlers (admin check, then the service).
- `dispatch.go` (`ingestPass`, dispatch of admitted rows), `inboxpoll.go`
  (payload shapes of poller rows), `ingestqueuehost.go` (the adapter).

Every store write, claim and persisted transition stays controller-owned.
No network or subprocess call moved: hooks still run on the pump
goroutine or in the external worker, outside the supervisor pass and
lock. `hookCommands` now only serves the controller's own hooks; an
admission hook without a controller runs `sh -c` in `ingestqueue`.

## Left in `pkg/control`

Dispatch of admitted rows and the poller payload types stay with the
subscription, activity and Linear code that consume them (JOB-142,
JOB-149). `control.go` still holds the pump wiring.
