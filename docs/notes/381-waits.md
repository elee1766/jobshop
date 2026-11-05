# Waits on outside state (#381)

2026-11-04. `jobshop wait add <cond>... --reason ... [--owner] [--any]
[--resume] [--timeout]` stores `fields.wait` on a task or step. A held
ticket is released as a deferral, not a failed attempt; the agent ends
its turn. While the wait is set, `ready` and claim skip the ticket.

Conditions (`TYPE:key=val,...`):

- `github.review:pr=o/r#5,login=bob[,head=sha][,state=approved]`: the
  login's latest review on that exact head. No head: the head seen on
  the first evaluation.
- `github.checks:pr=o/r#5,state=success|failure[,required=true]` (or
  `repo=o/r,sha=...`): success is at least one check, none pending or
  failed; failure is any failed or cancelled check.
- `github.pr:pr=o/r#5,state=merged|closed|head_changed`.
- `thread.reply:thread=o/r#7|ENG-12[,from=login]`: a human comment after
  the wait started, read through the subscription source and filters.
- `time:at=2h|RFC3339`.

The controller evaluates waits every subscriptions interval (default 1m)
and right after the ingest drain moves a new webhook delivery. One pass
fetches each PR once. All conditions must hold, or one with `--any`.
A review or checks wait whose PR merged, closed or moved head is
cleared as invalidated. A checks wait on a PR that conflicts with its
base is also invalidated, since no CI runs on it (#609). Clearing comments the reason on the ticket and,
with `--resume`, pre-assigns it to the parking session. Past
`--timeout` the ticket is labeled needs-human once and `attention` lists
it as kind `wait` with reason and owner. `wait clear <id>` removes a
wait by hand. Ticket edges (`waits_for`) stay the way to wait on a
ticket.
