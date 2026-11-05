# Store collections (#579): design

2026-11-05. Step 1 of #579: the design, for operator review before any
code. Nothing here is built yet.

## Why

`store.Store` has 75 methods. About 30 are tickets (tickets, edges,
labels, comments, history, meta, graph, runs, claims). The rest are five
domains, each with its own table, sqlc queries and methods that every
backend implements again:

| Domain | Table | Rows on gasserver | Methods |
| --- | --- | --- | --- |
| ingest | `inbox` | 3,774 | 11 |
| learnings | `learnings` (+ FTS `recall`) | 339 | 7 + `IndexRecall`, `Recall` |
| memory | `memory_edits` | 22 | 3 |
| idempotency | `write_keys` | 0 | 3 |
| polling | `poll_cursors` | 0 | 2 |

A collection is a named bucket of records with one generic schema. The
store implements it once; the domain logic (state machines, admission,
gating) moves into the domain packages. Each domain keeps its own bucket
(operator, 2026-11-05), so there is no shared blob table.

## Record

```go
// pkg/store/collection.go
type Record struct {
    Collection string
    Key        string          // unique within the collection
    State      string          // free text; the domain owns the values
    Value      json.RawMessage // the domain's struct, JSON
    Revision   int64           // starts at 1, +1 on every write
    Holder     string          // "" when unleased
    LeaseUntil time.Time
    Attempts   int64           // +1 on every Claim
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

Keys are strings. Domains that use integer ids today (learnings, memory
edits, inbox) keep them as the key, zero padded to 20 digits so key order
is id order. A collection can ask for a minted id (`Spec.Sequence`); the
store then fills `Key` from a per-collection counter on `Put` with an
empty key.

## Registration

```go
type CollectionSpec struct {
    Name     string   // [a-z][a-z0-9_]*, at most 40 chars
    Index    []Field  // indexed JSON fields
    Unique   []Field  // unique when non-empty (inbox natural_key)
    Sequence bool     // mint integer keys
    Search   []string // JSON fields fed to full-text search; opt in
}
type Field struct {
    Name string // used in filters
    Path string // JSON path in Value, e.g. "$.pr"
    Type string // "text" or "int"
}
```

Domain packages declare their spec in a package var and call
`st.Collection(ctx, spec)` once at startup. The call creates the table or
partition and indexes if missing (`IF NOT EXISTS`), and returns a
`Collection` handle. Adding an indexed field needs no migration. A
removed field's index is left in place. Specs are not stored: the code
is the source of truth, so two binaries with different specs on one
database only differ in which extra indexes exist.

## API

About 8 methods on a handle, replacing the 26 domain methods:

```go
type Collection interface {
    // Put inserts r, or does nothing if the key (or a Unique field)
    // exists. It returns the stored record and whether it was new.
    Put(ctx, r Record) (Record, bool, error)
    PutBatch(ctx, rs []Record) (fresh int, err error)
    Get(ctx, key string) (Record, error)            // ErrNotFound
    GetBy(ctx, field, value string) (Record, error) // a Unique field
    // Update writes state and value if r.Revision matches the stored
    // one; else ErrConflict. A non-empty holder must hold the lease, else
    // ErrNotHolder (the fence the inbox uses today).
    Update(ctx, r Record, holder string) (Record, error)
    List(ctx, q Query) ([]Record, error)
    Claim(ctx, states []string, holder string, lease time.Duration) (Record, error)
    Renew(ctx, key, holder string, lease time.Duration) (Record, error)
    Delete(ctx, key string) error
    Prune(ctx, states []string, before time.Time) (int64, error) // by updated_at
    Count(ctx, q Query) (map[string]int64, error)                // grouped by state
}

type Query struct {
    States []string
    Where  map[string]string // indexed field = value
    Since, Until time.Time   // on created_at
    Order  string            // "created" (default), "-created", "key", "-key", "updated"
    Limit  int               // 0 means DefaultLimit
}
```

`Where` accepts only indexed and unique fields; anything else is
`ErrInvalid`, so a query never scans a large bucket by accident.

## Claim and lease

The same rules as the inbox and ticket claims today:

- `Claim` picks the oldest record (`created_at, key`) whose state is in
  `states` and that is unleased (`holder = ''` or `lease_until <= now`).
  It takes it with a conditional `UPDATE ... WHERE key = ? AND (holder =
  '' OR lease_until <= now)`. A lost race tries the next row, at most 10
  times, then `ErrConflict`. This is the current `InboxClaim` loop.
- `Claim` sets holder and lease, bumps `attempts` and `revision`.
- `Renew` extends a live lease held by holder; otherwise `ErrNotHolder`.
  It does not bump the revision (as ticket heartbeats).
- `Update` with a holder clears the lease when the new state is not
  claimable. The domain says which states are claimable by passing them
  to `Claim`; the store does not know.
- Time comes from the store clock (`s.now()`), so tests use the fake
  clock as now (#359).

## Physical layout

**SQLite: one `records` table, partial indexes per collection**
(operator, 2026-11-05: one table with indexes is enough; a deployment
that outgrows SQLite moves to Postgres). No per-collection tables.

```sql
CREATE TABLE IF NOT EXISTS records (
  collection  TEXT NOT NULL,
  key         TEXT NOT NULL,
  state       TEXT NOT NULL DEFAULT '',
  value       TEXT NOT NULL DEFAULT '{}',
  revision    INTEGER NOT NULL DEFAULT 1,
  holder      TEXT NOT NULL DEFAULT '',
  lease_until INTEGER NOT NULL DEFAULT 0,
  attempts    INTEGER NOT NULL DEFAULT 0,
  created_at  INTEGER NOT NULL,
  updated_at  INTEGER NOT NULL,
  PRIMARY KEY (collection, key)
) WITHOUT ROWID;
-- claim, list and prune
CREATE INDEX IF NOT EXISTS records_state ON records (collection, state, updated_at);
CREATE INDEX IF NOT EXISTS records_created ON records (collection, created_at);
-- per indexed field, partial on the collection
CREATE INDEX IF NOT EXISTS rec_inbox_source ON records (json_extract(value, '$.source'), created_at)
  WHERE collection = 'inbox';
-- per unique field
CREATE UNIQUE INDEX IF NOT EXISTS rec_inbox_u_natural_key ON records (json_extract(value, '$.natural_key'))
  WHERE collection = 'inbox' AND json_extract(value, '$.natural_key') != '';
```

Queries repeat the index's `collection = '<name>'` literal and the same
`json_extract(value, '$.x')` expression, so the planner can use the
partial index (a bound parameter for the collection would not match the
index's WHERE). The conformance suite checks this with `EXPLAIN QUERY
PLAN`. Sequences live in one `collection_seq (name, next)` table.

**Postgres: one partitioned parent.**

```sql
CREATE TABLE IF NOT EXISTS records (
  collection  TEXT NOT NULL,
  key         TEXT NOT NULL,
  state       TEXT NOT NULL DEFAULT '',
  value       JSONB NOT NULL DEFAULT '{}',
  revision    BIGINT NOT NULL DEFAULT 1,
  holder      TEXT NOT NULL DEFAULT '',
  lease_until BIGINT NOT NULL DEFAULT 0,
  attempts    BIGINT NOT NULL DEFAULT 0,
  created_at  BIGINT NOT NULL,
  updated_at  BIGINT NOT NULL,
  PRIMARY KEY (collection, key)
) PARTITION BY LIST (collection);
CREATE TABLE IF NOT EXISTS records_inbox PARTITION OF records FOR VALUES IN ('inbox');
CREATE INDEX IF NOT EXISTS records_inbox_claim ON records_inbox (state, lease_until, created_at);
CREATE INDEX IF NOT EXISTS records_inbox_f_source ON records_inbox ((value->>'source'), created_at);
CREATE UNIQUE INDEX IF NOT EXISTS records_inbox_u_natural_key ON records_inbox ((value->>'natural_key'))
  WHERE value->>'natural_key' != '';
```

One query text serves every collection (`WHERE collection = $1`), and
partition pruning sends it to one partition. A partition can be
truncated, detached or moved on its own.

**Queries are built in Go, not sqlc.** sqlc cannot take a table name or
a JSON path as a parameter, and the field set is only known at
registration. A small builder in `sqlstore` renders the SQL per backend
(table name and JSON expression differ; placeholders go through the
existing Postgres rewriter). Identifiers come only from validated specs,
never from callers, and values are always bound parameters. Ticket
queries stay sqlc.

## Search

Learnings recall uses FTS5 today, fed by triggers on `learnings`. That
does not fit a generic table with triggers per collection, so:

- `Spec.Search` lists the JSON fields that feed search. On `Put` and on
  an `Update` that changes those fields, the store writes the existing
  `recall` index in the same transaction (kind = collection name, ref =
  key). That is what `IndexRecall` does for session journals now.
- `Recall` stays one store method, over all kinds, as it is today.
  Postgres keeps its `tsvector` recall table. No per-partition FTS.

So the only search-specific code is the existing recall table, and
learnings stop needing triggers.

## Queries that need more than an index

A domain that needs more than equality on indexed fields, a state set,
a created-at window and an order:

1. reads with `List` and filters or groups in Go when the bucket is small
   (learnings `CountLearnings` already groups in Go for globs; memory
   edits are 22 rows);
2. otherwise adds an indexed field that makes it an equality filter;
3. only if neither works, adds a named method to its domain store
   interface, implemented per backend next to the collection code, and
   says why in its note.

No step of the plan needs 3. Inbox `CountOlder` is `Count` with
`Until`; `Requeue` is `Get` then `Update`; `LearningsByPR` is
`Where{"pr": ...}`.

## Migration of existing data

One goose migration per domain PR, in plan order: poll_cursors and
write_keys (empty), memory_edits (22), learnings (339), inbox (3,774).
Each migration:

- **Up:** creates the `records` table and the collection's indexes
  (SQLite) or its partition (Postgres)
  with the same DDL the registration would, then
  `INSERT INTO records SELECT '<name>', ...` mapping columns into `value` with
  `json_object(...)` (SQLite) or `jsonb_build_object(...)` (Postgres).
  Keys are the old primary key (ids zero padded), `created_at` the old
  created or received time, and the sequence starts at `max(id) + 1`.
  Then it drops the old table, its indexes and its recall triggers. For
  learnings it re-keys the `recall` rows (`kind = 'learning'`, ref = new
  key) in the same migration.
- **Down:** recreates the old table and copies back with `json_extract`
  / `->>`. Down is lossless for every field that exists today.
- **Check:** before merge, run Up, then Down, then Up on an
  `sqlite3 .backup` copy of gasserver's database. Compare row counts and
  a per-row hash of the old columns. Paste the counts in the PR.

Registration at startup is then a no-op on a migrated database, and the
migration's DDL must match the registration's. A test asserts that.

## After

The ticket core keeps tickets, edges, labels, comments, history,
`ticket_meta`, graph, runs, resources and slots (resources and slots
join tickets in `Ready`/`ClaimNext`; revisit later). The target is about
40 methods on `store.Store` plus `Collection(spec)` and `Recall`. The
final count goes here when step 3 is done.

## Open questions for review

Status 2026-11-05: the operator has answered none of these yet. This
note merges only after they are answered here.

1. String keys with zero-padded ids, versus an integer key column.
2. Queries built in Go for collections, versus keeping sqlc with one
   query set per collection.
3. Recall shared across collections (above), versus a search index per
   collection.
