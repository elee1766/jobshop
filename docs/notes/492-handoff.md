# 492: jobshop handoff

2026-11-05.

`jobshop handoff [<ticket>] --to <pool> [--note "..."]` calls the
`handoff` RPC. In one store transaction it sets the pool, reopens the
ticket (holder and lease cleared), bumps the fence, writes one `handoff`
history entry (from_pool, pool, holder, note) and adds the note as a
comment by the caller. One revision, so no state shows the ticket held
in the new pool, and the old holder's claim is fenced out.

- A session may hand off only the ticket it holds (else forbidden).
- Admin, or a pool with the `handoff` admin grant, may hand off any
  open ticket, held or not. The github pack grants it to the plan pool,
  where stall-check runs.
- The pool must be configured. A closed ticket is rejected.
- Not a step failure, unlike a plain `release`.

The stall-check, refinery and dev prompts use it in place of
`ticket update --pool` plus `release`.
