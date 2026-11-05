# Write keys and poll cursors on collections (#579)

2026-11-05. First domain PR of #579 step 2 (design:
`docs/notes/579-collections.md`).

- `pkg/writekey` keeps the idempotency keys (#280, #415) in the
  `write_keys` collection, keyed by the scoped key. State is `pending`
  or `done`; the value holds `op`, `result`, `actor`. `Claim` is `Put`,
  `Finish` is `Get` + `Update` (retried on a revision conflict, first
  result wins), `Prune` is `Delete` by `created_at`.
- Poller cursors (#341) live in the `poll_cursors` collection in
  `pkg/ingest/cursor.go`, keyed `poller:repo:stream`, value `since`
  (unix ms) and `etag`.
- `Store` loses `ClaimWriteKey`, `FinishWriteKey`, `PruneWriteKeys`,
  `PollCursor`, `SetPollCursor`, their sqlc queries and both tables.
  Migration sqlite 00020 / postgres 00005 copies existing rows into
  `records` (and creates the Postgres partitions) and drops the tables;
  Down copies them back.
