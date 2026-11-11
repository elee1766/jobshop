# Ticket meta and label filters as IN lists (JOB-191, part 2)

2026-11-11. After the first JOB-191 change was deployed, selfdev's quiet tick
was still about 1s. A trace of `tick run_triggers` showed one `store list`
span under `trigger curate` at 420ms on every tick (17 of 17 traced ticks),
and others under `acphost-release-pr` (270ms) and `master-red` (100ms).

The list is the trigger's newest ticket: `List` by
`meta.trigger.name = <name>`, newest first, limit 1. SQLite planned the
correlated `EXISTS (SELECT 1 FROM ticket_meta m WHERE m.ticket_id = t.id
...)` as a scan of every ticket in id order, probing `ticket_meta` once per
row. That held with or without statistics. How slow it got depended on the
value: for `curate`, modernc's SQLite 3.54.1 probed through
`ticket_meta_lookup` per row (330ms). The sqlite3 3.53 CLI used the primary
key instead (5ms), so the slow plan did not show there.

## What changed

`List` writes label and meta filters as uncorrelated lists:
`t.id IN (SELECT m.ticket_id FROM ticket_meta m WHERE ...)`. SQLite reads
the matching ids once through the lookup index, then fetches those tickets
by primary key. Postgres plans both forms as a semi-join and is unchanged.
`TestTicketMetaPlan` fails if a meta or label List plan scans `tickets`.

## Numbers

A copy of the production database (2026-11-11, 1.26 GB, 8,526 tickets,
11,822 `ticket_meta` rows), modernc v1.61.1, one connection. Every
distinct meta value (2,768), full ticket columns:

| Query shape | Before p50 / max | After p50 / max |
|---|---|---|
| newest, limit 1 | 19.7ms / 624ms | 0.18ms / 17ms |
| priority order, limit 500 | 19.8ms / 621ms | 0.18ms / 16ms |
| `kind = run`, newest | 5.6ms / 92ms | 0.16ms / 11ms |

Through `Store.List`: `curate` newest 327ms to 0.5ms, `acphost-release-pr`
172ms to 0.3ms, `master-red` 97ms to 0.3ms.

Selfdev tick before this change (2026-11-11 22:00 to 22:40 UTC): about 1.0s,
`run_triggers` 0.8 to 1.1s. Post-deploy numbers go on JOB-191.
