# Portable resource slot queries (#426)

2026-11-04. `fields.resources` is copied to a `ticket_resources(ticket_id,
rkey, rname)` table the store rewrites on every insert and write, like
`ticket_meta` (#405). `rname` is the part before the first ':', so the cap
lookup is a plain join (`c.name = rkey OR c.name = rname`). The ready,
claim, claim-by-id, occupancy and waiter queries join it instead of
`json_each`. Migration 00015 backfills existing tickets. `fields` stays
the source of record.
