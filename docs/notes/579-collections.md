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

The target store interface the operator approved on #579
(2026-11-05) is the starting point. `Store` keeps 19 methods (tickets,
claims, resources, `Collection`, `Close`); the 26 Inbox, Learning,
MemoryEdit, WriteKey, PollCursor and Recall methods become the 9 on
`Collection`:

```go
// Store.Collection returns the handle for a domain bucket, creating its
// SQLite partial indexes or Postgres partition on first use.
Collection(ctx context.Context, spec CollectionSpec) (Collection, error)

type Collection interface {
    // Put inserts r, or does nothing if the key (or a Unique field)
    // exists; it returns the stored record and whether it was new.
    Put(ctx context.Context, r Record) (rec Record, created bool, err error)
    Get(ctx context.Context, key string) (Record, error) // ErrNotFound
    // Update writes state, value and holder if r.Revision matches the
    // stored one, else ErrConflict. A non-empty r.Holder must hold a
    // live lease, else ErrNotHolder (the inbox fence today).
    Update(ctx context.Context, r Record) (Record, error)
    List(ctx context.Context, f RecordFilter) ([]Record, error)
    Count(ctx context.Context, f RecordFilter) (map[string]int, error) // by state
    Claim(ctx context.Context, f RecordFilter, holder string, lease time.Duration) (Record, error)
    Renew(ctx context.Context, key, holder string, lease time.Duration) (Record, error)
    Delete(ctx context.Context, f RecordFilter) (int, error)
    Search(ctx context.Context, query string, limit int) ([]Record, error)
}

type RecordFilter struct {
    Keys          []string          // exact keys; one key is a point delete
    States        []string
    Where         map[string]string // indexed or unique field = value
    Since, Until  time.Time         // created_at window
    UpdatedBefore time.Time         // prune by age
    Order         string            // "created" (default), "-created", "key", "-key", "updated"
    Limit         int               // 0 means DefaultLimit
}
```

The earlier sketch's extra methods fold into these:

| Earlier sketch | Approved interface |
| --- | --- |
| `PutBatch` | `Put` in a loop inside the caller's batch; inbox drain is the only user and is not hot |
| `GetBy(field, v)` | `List(RecordFilter{Where: {field: v}, Limit: 1})` |
| `Delete(key)`, `Prune(states, before)` | `Delete(RecordFilter{Keys: ...})`, `Delete(RecordFilter{States: ..., UpdatedBefore: t})` |
| `Claim(states, ...)` | `Claim(RecordFilter{States: ..., Where: ...}, ...)` |
| store-wide `IndexRecall`, `Recall` | the `recall` collection's `Put` and `Search` (see Search) |

`Where` accepts only indexed and unique fields, anything else is
`ErrInvalid`, so a query never scans a large bucket by accident.
`Delete` and `Claim` with an empty filter are `ErrInvalid` too.

## Claim and lease

The same rules as the inbox and ticket claims today:

- `Claim` picks the oldest record (`created_at, key`) matching the
  filter (its states, plus any indexed fields) that is unleased
  (`holder = ''` or `lease_until <= now`).
  It takes it with a conditional `UPDATE ... WHERE key = ? AND (holder =
  '' OR lease_until <= now)`. A lost race tries the next row, at most 10
  times, then `ErrConflict`. This is the current `InboxClaim` loop.
- `Claim` sets holder and lease, bumps `attempts` and `revision`.
- `Renew` extends a live lease held by holder; otherwise `ErrNotHolder`.
  It does not bump the revision (as ticket heartbeats).
- `Update` with r.Holder set clears the lease when the new state is not
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

Learnings recall uses FTS5 today, fed by triggers on `learnings`, and
`IndexRecall`/`Recall` are store methods. The approved interface drops
both from `Store` and gives each collection `Search`:

- `Spec.Search` lists the JSON fields that feed search. On `Put` and on
  an `Update` that changes them, the store writes one FTS row in the
  same transaction, keyed by (collection, key). SQLite: one FTS5 table
  `records_fts(collection UNINDEXED, key UNINDEXED, body)`. Postgres: a
  generated `tsvector` column on the parent with a GIN index declared
  there, so every partition gets it. `Search` filters on the collection
  and ranks within it.
- Recall across kinds (comments, journals, learnings) becomes one
  `recall` collection. A `RecallDoc` is its record: key `<kind>:<ref>`,
  `kind`, `project` and `ticket` as indexed fields, `text` searched.
  `control/recall.go` calls `recall.Put` where it calls `IndexRecall`
  today, and `recall.Search` where it calls `Recall`, dropping hits
  of another kind or project in Go (`Search` takes no filter, so it
  asks for a few times the limit). The learn
  package puts each learning's summary and
  fix into `recall` too, so learnings need no triggers and recall keeps
  one ranking across kinds.

So the existing recall table becomes the `recall` collection, and the
only search code is the FTS table or column behind `Search`.

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
`List(RecordFilter{Where: {"pr": ...}})`.

## Migration of existing data

One goose migration per domain PR, in plan order: poll_cursors and
write_keys (empty), memory_edits (22), learnings (339), inbox (3,774),
then recall (the existing FTS rows, into the `recall` collection and
`records_fts`). Each migration:

- **Up:** creates the `records` table and the collection's indexes
  (SQLite) or its partition (Postgres)
  with the same DDL the registration would, then
  `INSERT INTO records SELECT '<name>', ...` mapping columns into `value` with
  `json_object(...)` (SQLite) or `jsonb_build_object(...)` (Postgres).
  Keys are the old primary key (ids zero padded), `created_at` the old
  created or received time, and the sequence starts at `max(id) + 1`.
  Then it drops the old table, its indexes and its recall triggers. For
  learnings it re-keys the old recall rows (`kind = 'learning'`, ref =
  new key) in the same migration, before the recall move.
- **Down:** recreates the old table and copies back with `json_extract`
  / `->>`. Down is lossless for every field that exists today.
- **Check:** before merge, run Up, then Down, then Up on an
  `sqlite3 .backup` copy of gasserver's database. Compare row counts and
  a per-row hash of the old columns. Paste the counts in the PR.

Registration at startup is then a no-op on a migrated database, and the
migration's DDL must match the registration's. A test asserts that.

## After

The ticket side follows the approved interface in step 3: `List(Filter)`,
`Thread`, `Apply`, `Update`, `Release(Outcome)`, `Expire`, `Slot` and
`Resources` replace today's ticket methods, one PR per group, each with
conformance cases for every filter field and outcome. Target: 19
methods on `store.Store` plus 9 on `Collection`. The final count goes
here when step 3 is done.

## Review questions, answered

The three questions this note first left open are settled by the
interface the operator approved on #579 (2026-11-05):

1. **Keys: strings.** `Collection.Get(ctx, key string)` and
   `Renew(ctx, key, ...)` take string keys. Integer ids become zero
   padded keys, and `Spec.Sequence` mints them.
2. **Queries built in Go.** `RecordFilter.Where` names indexed fields
   known only at registration, and the Postgres partition name comes from
   the spec, so sqlc cannot express them. The Go builder above stays;
   ticket queries stay sqlc.
3. **Search per collection.** `Collection.Search` replaces the
   store-wide `Recall`; cross-kind recall is the `recall` collection
   (see Search).
