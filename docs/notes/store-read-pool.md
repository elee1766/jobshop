# SQLite read pool

2026-11-05. Production SQLite used one connection for everything, so Desk
and CLI reads queued behind supervisor writes (Desk `/api/sessions` 6s,
`/api/pools` 5s, `/api/attention` 5s on gasserver).

`sqlite.Open` now opens a second pool of 4 `query_only` connections for
reads outside a transaction (`Get`, `List`, thread, recall, resource and
slot reads). Writes and reads inside a write transaction stay on the
single writer connection, so claims and read-then-write logic are
unchanged.

Consistency: WAL readers see the last committed transaction. A read
issued after a write commits sees it; a read during an open write sees
the state before it, as it did before (it waited, then saw it). In-memory
and `Concurrent` databases keep one pool.
