# #579 step 3: Thread and Changes

2026-11-05. Second ticket-side group of #579.

- `Store.Thread(id, Parts)` replaces `Edges`, `Comments` and `History`.
  `Parts` picks which of the three to read; a part not asked for is nil.
  `store.AllParts` reads all three. Show, trace and artifact show read
  their thread in one call.
- `Store.Changes` returns a `ChangePage`: the rows after a seq and
  `Last`, the newest seq. `LastSeq` is gone; `n < 0` reads only `Last`.
  Rows are capped at `Last`, so a page never holds a row past it.
- Helpers in package store, not interface methods: `Edges`,
  `Comments`, `History`, `LastSeq`.
- The SQL is built in Go with bqb in `pkg/store/sqlstore/thread.go`.
  `ListHistory`, `ListComments`, `ListChanges` and `LastSeq` are gone
  from sqlc; `ListEdges` stays for the blocked check inside a claim.
- `store.Store` is at 40 methods (43 before).
