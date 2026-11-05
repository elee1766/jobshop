# #579 step 3: Release(Outcome) and Expire

2026-11-05. Third ticket-side group of #579.

- `Store.Release(actor, id, claim, Outcome)` replaces `Release`,
  `Handoff` and `CloseTicket`. `Outcome.Close` closes with that
  resolution, `Outcome.Pool` hands off with `Note` as a comment, and
  neither returns the ticket to open (deferred when `Until` is set,
  a counted failure otherwise). Close and Pool together is
  `ErrInvalid`; return to open needs a claim.
- An empty id or actor comes from the claim; a claim for another
  ticket is `ErrInvalid`.
- `Store.Expire(actor, holder, reason)` replaces `ReleaseHolder` and,
  with holder "", `SweepLeases`.
- Helper in package store, not an interface method: `store.Close`.
- `store.Store` is at 37 methods (40 before).
