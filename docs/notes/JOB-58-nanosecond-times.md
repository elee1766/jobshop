# JOB-58: timestamps in nanoseconds

2026-11-06

Store timestamp columns hold unix nanoseconds (int64), not milliseconds:
tickets, edges, history, comments, recall, resource_slots, records, and
the ingest spool (spool schema v4). Migrations sqlite 00026 and postgres
00011 multiply existing values by 1,000,000; Down divides.

- Ticket creation uses a per-store `storerules.Clock`: each created
  ticket gets its own `now`, strictly increasing (same or earlier clock
  reading becomes last+1ns). A batch insert lists in creation order; the
  id tiebreak only matters across controllers sharing a nanosecond.
- `ms`/`fromMS` in the stores are now `ns`/`fromNS`; `storerules.Stamp`
  keeps full precision.
- The step gate hold (`gateHold`) moved from year 9999 to 2200-01-01,
  since 9999 overflows int64 ns. The migration clamps any defer_until
  past 2200 to that value, so held steps still match.
- JSON values inside records keep their own units (journal `at`, poll
  cursor `since`, memory edit times are still ms). API output stays
  RFC 3339.
