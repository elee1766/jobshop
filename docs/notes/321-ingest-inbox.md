# #321 Ingest inbox: table and store API

2026-11-04. Item 1 of #318. Nothing writes to the inbox yet; webhooks
and pollers move onto it in later #318 items.

- Table `inbox` (migration 00008): one row per webhook delivery or
  poller find, with source (github or linear), subject columns, raw
  payload, admitted payload, hook results, state and a lease.
- States: `pending` (stored) -> `admitted` | `denied` | `held` ->
  `dispatched` | `failed`. dispatched, denied and failed are terminal.
- Dedupe: `delivery_id` is unique. `InboxPut` of a known id changes
  nothing and returns the stored row with `dup` true, so a redelivered
  webhook or a poller seeing the same thing again is a no-op.
- Lease: `InboxClaim(states, holder, lease)` takes the oldest row in
  those states whose lease is empty or expired and adds one to
  `attempts`. The take is a conditional update, so two claimers never
  get the same row. `InboxRenew` and `InboxSetState` are fenced by
  holder and a live lease (`ErrConflict` otherwise); `InboxSetState`
  drops the lease so the next stage can claim the row in its new state.
- At least once: a worker that dies leaves its lease to expire and the
  row is claimed again. Consumers must be idempotent (run keys, write
  keys).
- Retention: `ingest.retention` (default `14d`, Go duration or `Nd`).
  The controller prunes terminal rows received before now minus the
  retention, at most once an hour on its tick.
