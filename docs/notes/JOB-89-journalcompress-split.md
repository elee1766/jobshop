# pkg/control split, step two: journal compression (JOB-89, 2026-11-08)

Second subsystem out of `pkg/control`, after journal shipping (JOB-119,
`JOB-119-control-split.md`). No behavior change: the sweep, its hourly
spacing, the idle IO class and the log line are as in
`631-journal-compress.md`.

## Boundary

- `pkg/control/journalcompress` owns the sweep: `Sweeper.Maybe` starts
  one in the background when due (tick-only), `Sweep` is its body,
  `lowIOPriority` (Linux `ioprio_set`) moved with it.
- It reaches the controller only through `journalcompress.Host`: `Now`,
  `List` of closed sessions, `SessionDir`, `Go` (the controller's
  WaitGroup) and `Context` (the controller's, not the tick's).
- `pkg/control/journalcompresshost.go` is the adapter (`compressHost`);
  `Controller.compressor()` makes the sweeper on first use, like
  `shipper()`. `maybeCompressJournals` stays as the tick step.
- The package does not import `pkg/control`. Its unit test runs on a
  fake host; the controller-level test (compressed journals still served
  by `session.journal`, `session.subscribe` and recall) stays in
  `pkg/control/journalcompress_internal_test.go`.

## Left in journal

`journal.go` (`session.journal`) and `journalsearch.go` stay: they need
`scope`, `canReadJournal` and `requireGrant`. Moving them needs an
access interface first, so the next steps are ingest, then Linear (see
the candidates in the JOB-119 note).
