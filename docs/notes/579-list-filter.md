# #579 step 3: List(Filter)

2026-11-05. First ticket-side group of #579.

- `Store.List(Filter)` replaces `Ready`, `Orphaned`, `ListChildren`,
  `ListRuns`, `FindOpenRunByKey`, `FindOpenBatchByKey` and
  `StepsCreated`. The store interface loses seven methods.
- New `Filter` fields: `Kinds`, `Statuses`, `Unheld`, `PoolNotIn`,
  `Ready`, `Key`, `Created` (`TimeRange`, `[Since, Until)` on
  created_at), `Order` (`priority`, `newest`, `oldest`, `index`) and
  `Limit: NoLimit`. `Newest bool` is gone; `rpc.ListParams.Newest`
  stays and maps to `OrderNewest`.
- Helpers in package store, not interface methods: `Children`,
  `FindOne`, `OpenRun`, `OpenBatch`.
- The SQL is built in Go with bqb in `pkg/store/sqlstore/list.go`.
  `Driver.Render` turns it into engine text (Postgres: `ToPgsql`) and
  `Driver.Conn` routes it through SQLite's statement cache (#433).
  `ListTickets`, `ListTicketsNewest`, `ReadyTickets`,
  `NextClaimCandidate`, `OrphanedTickets`, `ListChildren`, `ListRuns`,
  `FindOpenRunByKey`, `FindOpenBatchByKey` and `ListStepsCreated` are
  gone from sqlc.
- `List(Filter{Ready: true})` and `ClaimNext` share the readiness
  fragments (`notBlocked`, `slotsFree`, `providerReady`, `notWaiting`),
  so the rule is written once. `ClaimWriteTicket`, `ResourceFull` and
  `ProviderEmpty` still repeat it in sqlc; they move with `Claim`.
- `Statuses` of exactly open and in_progress is written
  `status <> 'closed'`, so the partial index `tickets_run_key` applies.
- Conformance: `storetest` case `ListShapes`.
