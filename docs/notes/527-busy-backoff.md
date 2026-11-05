# #527 Backoff on serialization retries

2026-11-05.

- `sqlstore.Store.tx` already reran a transaction that lost a write race
  (Postgres 40001/40P01) up to 16 times, but at once. In the
  conformance test eight claimers then collided in lockstep, used up the
  tries and `CloseTicket` returned `ErrConflict`.
- Each rerun now waits a random time up to `1ms << try`, capped at 64ms
  (worst case about 0.7s over all tries), and stops if the context ends.
  SQLite has no `IsBusy` driver hook, so it never waits.
