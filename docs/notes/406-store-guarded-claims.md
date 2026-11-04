# Store: claims and sweeps without a single-writer database (#406)

2026-11-04. Behavior as built.

- A claim (`Claim`, `ClaimNext`) writes through `ClaimWriteTicket`: one
  `UPDATE ... WHERE revision = ? AND status IN ('open','in_progress')
  AND NOT EXISTS (<open blocker>) AND <free resource slot>)`. Both checks
  run in the same statement as the write, so a blocker that reopens (or
  a slot taken) after the claimer's read fails the claim (`ErrConflict`)
  instead of leaving a blocked ticket claimed or a resource over its
  capacity. The rules match `ReadyTickets` and `NextClaimCandidate`; the
  slot check is skipped for an in-progress resume, which holds its slots.
- `ClaimNext` retries on `ErrConflict` (another claimer took the
  candidate, or it became blocked), up to 64 tries, each in its own
  transaction, then returns the conflict. A candidate held by someone
  else when re-read is a conflict too.
- `SweepLeases` writes through `ExpireWriteTicket`, which re-checks
  `lease_until < now` at write time: `Heartbeat` renews without a
  revision bump, so the revision guard alone would let a sweep reopen a
  just-renewed ticket. Tickets that changed since the list was read are
  skipped and not returned. `ReleaseHolder` uses the same skip.
- SQLite `SQLITE_BUSY` from a transaction (a deferred transaction lost a
  write race) reruns the transaction, up to 16 times, then maps to
  `ErrConflict`. busy_timeout does not wait on a read-to-write upgrade,
  so without the rerun `-race` runs failed now and then.
- `sqlite.Options.Concurrent` (tests only) opens with
  `_txlock=deferred` and 8 connections. `TestConformanceConcurrent`
  runs the whole storetest suite that way, including the new
  `ConcurrentClaims` case: 8 claimers, every ticket claimed exactly once,
  a blocked ticket never, and 8 direct `Claim`s of one ticket with one
  winner.
- Inbox claim and renew were already single guarded statements
  (`InboxTake`, `InboxRenew`). #400 slot grants are in the claim guard.
