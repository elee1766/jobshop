# Ingest dispatch (#442, 2026-11-04)

Part of #318. Before this, pollers delivered comments and activity to
sessions directly and wrote an inbox row on the side, so admission hooks
could not stop or rewrite what a session saw.

Now:

- `pollThreads` and `pollActivity` only write inbox rows (delivery ids
  `poll:<repo|linear>:...` and `poll:activity:...`), advance their read
  position, then run `ingestPass`.
- `ingestPass` releases answered holds, admits pending rows, then
  dispatches admitted ones. The dispatcher claims rows as holder
  `dispatch` with a 5 minute lease, so a crashed pass is retried.
- Comment rows are grouped by thread and delivered to followers with
  the admitted payload (after mutating hooks). Each delivery uses write
  key `dispatch/<row id>`, so a redelivered row sends nothing twice.
- Activity rows feed the activity log, revert checks and events.
- Webhook rows are already handled by their handler; dispatch only
  marks them dispatched.
- The old per-comment Linear intake filter in the poller is gone; the
  intake hooks run as validating admission hooks (#348).

Tests: `pkg/control/dispatch_internal_test.go`.
