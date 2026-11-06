# JOB-38: storetest race and read-after-write cases

2026-11-06. `pkg/store/storetest/races.go` adds cases that run on both
engines through `storetest.Run`:

- RaceClaimNextExtend: holders re-claim their own work; one holder per
  ticket, stored fence not below any fence handed out.
- RaceReleaseClaimFences: claim/release loop on one ticket; fences are
  unique and increase, released claims cannot close.
- RaceReleaseExpire, RaceHeartbeatSweep: holder writes vs Expire; a
  swept claim cannot heartbeat, an unswept one is still held.
- RaceUpdateClaim, RaceCASUpdates: no lost label update, one claim wins,
  one CAS update wins.
- RaceSlotRelease: ClaimNext vs ReleaseSlots; one slot per claim, none
  held after every claim ended.
- ReadAfterWrite: Get and List see create, update, claim and close at
  once from any goroutine (SQLite reader pool).

Store fixes the cases found:

- Postgres reads outside a transaction use a read committed pool
  (Driver.ReadDB). Under serializable, Get and List failed with 40001.
- Expire and ReleaseSlots reset their result when tx reruns after a
  lost write race. Before, a rolled-back attempt's ids stayed in the
  result, so the sweep reported tickets it did not reopen.
- Heartbeat runs through the retrying write transaction, so a Postgres
  serialization failure is rerun instead of returned raw.

Goroutines start on a barrier (16 racers). Expire and AddLabel retry on
ErrConflict, the error a serialization failure maps to. Run with
`go test -race ./pkg/store/...`.
