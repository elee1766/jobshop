# pkg/control split, step one: journal shipping (JOB-119, 2026-11-08)

First subsystem out of `pkg/control` under JOB-89's plan (one subsystem
per PR, most self-contained first). No behavior change.

## Boundary

- `pkg/control/journalship` owns journal shipping to VictoriaLogs
  (#530): per-session cursors and pending rows, batching, backoff,
  the lag gauge and attention `journal-lag`, and `journal.backfill`'s
  work. `Shipper.Run` is the pump, `Pass` one pass, `Backfill` the RPC's
  body, `LagAttention` the attention items, `Client(cfg)` the configured
  VictoriaLogs client.
- It reaches the controller only through `journalship.Host`: config,
  clock (`Now`, `After`), `List` of session tickets, `SessionDir`,
  `Redactor` (session env plus rules), `Work` (run, workflow, step of a
  ticket) and `SaveCursor` (raises `SessionFields.ShippedSeq`).
- `pkg/control/journalshiphost.go` is the adapter (`shipHost`) and
  makes the shipper on first use, so controllers built in tests without
  `Run` get one too. Access checks (`journal.search` grant, admin-only
  backfill) stay in the controller's RPC handlers.
- `journalship` does not import `pkg/control`; unit tests run it on a
  fake host and an httptest VictoriaLogs.

## Next candidates

In JOB-89 order, smallest coupling first:

1. The rest of journal: `journal.go` (session.journal reads),
   `journalcompress.go` plus `ioprio_*`, `journalsearch.go`. Needs
   `scope` and `canReadJournal` to stay in control or move behind a
   small access interface.
2. Ingest: `ingestwork.go`, `ingestworker.go`, `ingestdrain.go`,
   `inboxpoll.go` (already leans on `pkg/ingest`).
3. Linear: `linear*.go`, `intake.go`.
4. Waits and triggers: `wait*.go`, `trigger*.go`, `watch.go`,
   `subscribe.go`. Tied to the tick lock and run state, so later.
