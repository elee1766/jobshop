# Memory edits on a collection (#579)

2026-11-05. Second domain PR of #579 step 2 (design:
`docs/notes/579-collections.md`).

- `pkg/memedit` keeps curator gate edits (#274) and local memory writes
  (#592) in the `memory_edits` collection. State is `pending`,
  `helped` or `not_helped`; the value holds the old columns, times as
  unix ms.
- Key is the edit id as text, not the PR as the design note said: local
  writes have no PR, and `learn evaluate --id`, `memory:<id>` refs on
  learnings and the RPC name edits by integer id. `pr` is a Unique field,
  so recording a PR twice returns the first edit. A new id is one past
  the highest of the 20 newest records; a taken id retries the next.
- `SetResult` no longer writes the cited learnings' gate in the same
  transaction: the controller writes it after, through
  `Store.SetLearningGate`, which goes when learnings move.
- `Store` loses `AddMemoryEdit`, `AddMemoryWrite`, `MemoryEdit`,
  `MemoryEdits`, `SetMemoryEditResult` and their sqlc queries, and gains
  `SetLearningGate`. Migration sqlite 00021 / postgres 00006 copies rows
  into `records` (ids kept) and drops the table; Down copies them back.
