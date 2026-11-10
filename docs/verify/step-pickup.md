# Step pickup time

2026-11-10 (JOB-193). How to measure the time from a step's creation to its
first claim on a live instance. Run it before and after a deploy that
changes pass scheduling, and compare the same kind of step.

The store is SQLite at `<root>/jobshop.db`. Read it in read-only mode;
never write. `sqlite3` is not on the agent PATH:
`nix shell nixpkgs#sqlite -c sqlite3 -readonly <root>/jobshop.db`.
Timestamps are unix nanoseconds. History op `claim` is a claim: an agent
session claims an agent step, `controller` claims a command or http step.

## Query

Steps created since a cutoff that have no `blocks` edge into them (a step
behind another is ready only when that one closes, so its wait is not
pickup time), with the time to their first claim in seconds and their
kind:

```sql
SELECT
  CASE WHEN json_extract(t.fields, '$.type') IN ('command', 'http')
       THEN 'auto' ELSE 'agent' END AS kind,
  (MIN(h.at) - t.created_at) / 1e9 AS wait_s
FROM tickets t
JOIN history h ON h.ticket_id = t.id AND h.op = 'claim'
WHERE t.kind = 'step'
  AND t.created_at > strftime('%s', 'now', '-3 days') * 1000000000
  AND NOT EXISTS (SELECT 1 FROM edges e WHERE e.from_id = t.id AND e.kind = 'blocks')
GROUP BY t.id;
```

A `needs` is an edge from the step to what it needs, so `from_id = t.id`
keeps steps with no `needs`. Take the count, median and p90 of `wait_s`
per `kind`, for example by sorting the output and reading the middle and
the 90th percentile row.

## Reading the result

* Before JOB-193, a command step created by a callback waited for the next
  tick: a median near half the tick (16.7s on jobcity with a 30s tick) and
  a p90 near the tick (26.8s).
* After, a command step should be claimed within a second or two of its
  creation: the callback worker applies the invocation, wakes a pass, and
  the pass starts the step.
* Agent steps wait at most for an idle session's poll (`Poll`, 5s), or for
  a session to start when the pool has none. A long wait with a full pool
  is capacity, not scheduling.
* Extra passes from wakes should stay near the number of steps created.
  `jobshop.tick.step.duration` has one sample per step per pass, so its
  count shows the pass rate.

Only compare steps created after the deploy with steps from before it,
and keep the window the same length.
