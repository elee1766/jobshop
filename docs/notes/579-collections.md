# Store collections (#579): design

2026-11-05. Step 1 of #579: the design, for operator review before any
code. Nothing here is built yet. Revised after the operator's answers on
#579 (string keys, no sqlc, one search index per collection) and the
approved target store interface.

## Why

`store.Store` has 75 methods. About 30 are tickets. The rest are five
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
store implements it once; domain logic (state machines, admission,
gating) moves into the domain packages. Domains stay apart through the
`collection` column, per-collection indexes and, on Postgres, one
partition each.

## Target interface

From the approved interface on #579 (2026-11-05). `Store` has 19 methods:

```go
type Store interface {
	Get(ctx context.Context, id string) (Ticket, error)
	List(ctx context.Context, f Filter) ([]Ticket, error)
	Thread(ctx context.Context, id string, p Parts) (Thread, error)
	Changes(ctx context.Context, after int64, n int) (ChangePage, error)
	NewID(project string) (string, error)

	Apply(ctx context.Context, actor string, g Graph) ([]Ticket, error)
	Update(ctx context.Context, actor, id string, u Update) (Ticket, error)
	Record(ctx context.Context, actor, id, op string, diff any) error

	ClaimNext(ctx context.Context, req ClaimNext) (Ticket, Claim, error)
	Claim(ctx context.Context, id, holder string, lease time.Duration) (Ticket, Claim, error)
	Heartbeat(ctx context.Context, c Claim, lease time.Duration) (Claim, error)
	Release(ctx context.Context, actor, id string, c *Claim, o Outcome) (Ticket, error)
	Expire(ctx context.Context, actor, holder, reason string) ([]string, error)

	SetResources(ctx context.Context, r ResourceConfig) error
	Slot(ctx context.Context, actor string, e SlotEvent) (Slot, string, error)
	ReleaseSlots(ctx context.Context) ([]Slot, error)
	Resources(ctx context.Context, pool string) (ResourceReport, error)

	Collection(ctx context.Context, spec CollectionSpec) (Collection, error)
	Close() error
}
```

`Filter`, `Outcome`, `Graph` and the rules (edges change only in
`Apply`, status only in `Claim`/`Release`/`Expire`, `List(Ready)` and
`ClaimNext` share one predicate) are as approved on #579 and are step 3.
This note covers `Collection` and the move off sqlc.

## Record

```go
type Record struct {
	Collection string
	Key        string          // the domain's natural key, unique in the collection
	State      string          // free text; the domain owns the values
	Value      json.RawMessage // the domain's struct, JSON
	Revision   int64           // 1 on insert, +1 on every write
	Holder     string          // "" when unleased
	LeaseUntil time.Time
	Attempts   int64           // +1 on every Claim
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
```

Keys are strings and are the domain's natural key (operator answer 1):

| Collection | Key |
| --- | --- |
| `inbox` | delivery id |
| `poll_cursors` | `poller:repo:stream` |
| `write_keys` | the write key |
| `learnings` | the old id as a decimal string; new ones from `NewKey` |
| `memory_edits` | the PR (already unique today) |
| `journals` | the session id |

Keys are never padded and key order means nothing. Ordering is always
by `created_at` (then key, for ties) or by a declared index. Domains
with no natural key (new learnings) get a key from `store.NewKey()`, a
time-sortable random string like ticket ids, so no sequence table is
needed. The domain packages stop assuming integer ids. The CLI and RPC
take the string (`jobshop learn show 312` still works, because the old
ids migrate as `"312"`).

## Registration

```go
type CollectionSpec struct {
	Name   string  // [a-z][a-z0-9_]*, at most 40 chars: it names a partition
	Index  []Field // indexed JSON fields
	Unique []Field // unique when non-empty (inbox natural_key)
	Search []Field // fields fed to this collection's full-text index
}
type Field struct {
	Name string // the name filters use
	Path string // JSON path in Value, [a-z_][a-z0-9_.]*, e.g. "pr"
}
```

A domain package declares its spec as a package var and calls
`st.Collection(ctx, spec)` once at startup. The call validates the spec,
creates the partition (Postgres) and the indexes if missing
(`IF NOT EXISTS`), and returns a handle. Adding a field needs no
migration. A dropped field's index stays until a migration removes it.
Specs are not stored: the code is the source of truth.

## Collection API

The 9 methods from the approved interface:

```go
type Collection interface {
	// Put inserts r, or does nothing if the key or a Unique field already
	// exists; it returns the stored record and whether it was created.
	Put(ctx context.Context, r Record) (rec Record, created bool, err error)
	Get(ctx context.Context, key string) (Record, error) // ErrNotFound
	// Update writes state, value and holder if r.Revision matches the
	// stored revision, else ErrConflict. A non-empty r.Holder must hold a
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
	Keys          []string
	States        []string
	Where         map[string]string // declared Index or Unique field = value
	Since, Until  time.Time         // created_at window
	UpdatedBefore time.Time         // prune by age
	Order         string            // "created" (default), "-created", "updated"
	Limit         int               // 0 means DefaultLimit
}
```

- `Where` takes only declared fields; anything else is `ErrInvalid`, so a
  query never scans a bucket by accident.
- `Delete` and `Claim` with an empty filter are `ErrInvalid`.
- `Search` on a collection with no `Search` fields is `ErrInvalid`.

Today's methods map onto it like this:

| Today | Collection |
| --- | --- |
| `InboxPut`, `InboxPutBatch` | `Put` (the batch is a loop in one ingest drain) |
| `InboxClaim`, `InboxRenew` | `Claim(RecordFilter{States: ...})`, `Renew` |
| `InboxSetState`, `InboxRequeue` | `Get` + `Update` |
| `InboxGet`, `InboxList` | `Get`, `List` |
| `InboxPrune` | `Delete(RecordFilter{States: terminal, UpdatedBefore: t})` |
| `InboxCounts`, `InboxCountOlder` | `Count`, `Count(RecordFilter{States: ..., Until: t})` |
| `AddLearning`, `GetLearning`, `UpdateLearning` | `Put`, `Get`, `Update` |
| `Learnings`, `LearningsBetween`, `LearningsByPR` | `List` with `Where{"scope"/"kind"/"pr"}` and a window |
| `CountLearnings` | `List` with a window, grouped in Go (globs; 339 rows) |
| `AddMemoryEdit`, `MemoryEdits`, `SetMemoryEditResult` | `Put`, `List`, `Update` |
| `ClaimWriteKey`, `FinishWriteKey`, `PruneWriteKeys` | `Put`, `Update`, `Delete` |
| `PollCursor`, `SetPollCursor` | `Get`, `Put` or `Update` |
| `IndexRecall`, `Recall` | `Search` per collection (see Search) |

## Claim and lease

The same rules as the inbox and ticket claims today:

- `Claim` picks the oldest record (`created_at, key`) matching the filter
  that is unleased (`holder = ''` or `lease_until <= now`), and takes it
  with a conditional `UPDATE ... WHERE collection = ? AND key = ? AND
  (holder = '' OR lease_until <= now)`. A lost race tries the next row,
  at most 10 times, then `ErrConflict`. This is the `InboxClaim` loop.
- `Claim` sets holder and lease and bumps `attempts` and `revision`.
- `Renew` extends a live lease held by holder, else `ErrNotHolder`. It
  does not bump the revision (as ticket heartbeats).
- `Update` with `r.Holder` set keeps or clears the lease as the record
  says. The store does not know which states are claimable: the domain
  passes them to `Claim`.
- Time comes from the store clock, so tests use a fake clock (#359).

## Physical layout

**SQLite: one `records` table, partial indexes per collection.**

```sql
CREATE TABLE records (
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
CREATE INDEX records_state   ON records (collection, state, lease_until, created_at);
CREATE INDEX records_created ON records (collection, created_at);
CREATE INDEX records_updated ON records (collection, state, updated_at);
-- registration, per Index field:
CREATE INDEX IF NOT EXISTS rec_inbox_source ON records (json_extract(value, '$.source'), created_at)
  WHERE collection = 'inbox';
-- per Unique field:
CREATE UNIQUE INDEX IF NOT EXISTS rec_inbox_u_natural_key ON records (json_extract(value, '$.natural_key'))
  WHERE collection = 'inbox' AND json_extract(value, '$.natural_key') != '';
```

A partial index is used only when the query's WHERE implies the
index's, so field queries write the collection as a literal
(`collection = 'inbox'`), not a parameter. The name is validated, so
this is safe. The conformance suite checks the plan with `EXPLAIN QUERY
PLAN`.

**Postgres: `records` partitioned by collection.**

```sql
CREATE TABLE records (
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
CREATE INDEX records_state   ON records (state, lease_until, created_at);
CREATE INDEX records_created ON records (created_at);
CREATE INDEX records_updated ON records (state, updated_at);
-- registration:
CREATE TABLE IF NOT EXISTS records_inbox PARTITION OF records FOR VALUES IN ('inbox');
CREATE INDEX IF NOT EXISTS records_f_source ON records ((value->>'source'), created_at);
CREATE UNIQUE INDEX IF NOT EXISTS records_inbox_u_natural_key ON records_inbox ((value->>'natural_key'))
  WHERE value->>'natural_key' != '';
```

- There is no DEFAULT partition, so a write to an unregistered collection
  fails loudly.
- Indexes are declared on the parent, so every partition gets them. A
  field index has no `WHERE collection`: partition pruning already sends
  a query to one partition.
- Unique fields are the exception. A unique index on a partitioned
  parent must include the partition key, so each one is declared on the
  collection's own partition.
- A Postgres test checks that registration creates the partition and that
  writing to an unregistered collection errors.

## SQL without sqlc

Operator answer 2: no sqlc. The store uses plain SQL in Go, with
`database/sql` for SQLite and pgx for Postgres. pgx (`jackc/pgx/v5`) is
already a direct dependency of `pkg/store/postgres`, so nothing is added.

- **Collections.** A small builder in `pkg/store/sqlstore/sqlq` renders
  each statement from the spec and the `RecordFilter`. It has a
  `Dialect` with the placeholder style (`?` or `$n`), the JSON field
  expression (`json_extract(value, '$.x')` or `value->>'x'`), the upsert
  form and the FTS query. Identifiers come only from validated specs.
  Values are always bound. The SQL text is cached per filter shape.
- **Postgres** runs on a `pgx` connection pool, not the
  `database/sql` shim, and the sqlite-to-postgres rewriter
  (`postgres/rewrite.go`) goes away with the last sqlc query.
- **Tickets** move off sqlc group by group in step 3, as the methods
  merge. `List(Filter)` builds its WHERE from the filter fields, with
  `Ready` as one shared fragment that `ClaimNext` also uses.
- **Removal.** The PR that removes the last sqlc query also deletes
  `sqlc.yaml`, `queries/`, `sqlcdb` and the sqlc CI check, and updates
  the sqlc rule in AGENTS.md. The goose migrations stay.

## Search

Operator answer 3: one full-text index per collection. `Spec.Search`
names the fields. A collection with Search fields gets its own index,
kept in the same transaction as the record write.

- **SQLite:** an FTS5 table `fts_<name>(key UNINDEXED, body)`. `Put`,
  and any `Update` that changes a Search field, rewrite the row. `body`
  joins the Search fields with newlines.
- **Postgres:** a partial GIN index on the parent,
  `CREATE INDEX records_fts_<name> ON records USING gin
  (to_tsvector('english', <fields joined>)) WHERE collection = '<name>'`,
  so it lands on that partition. Ranking uses `ts_rank`.
- **`Search`** queries only its own collection and returns records best
  match first.

Where today's `recall` rows go:

- **Learnings:** the `learnings` collection, Search on `summary` and
  `fix`. The FTS triggers on `learnings` go away.
- **Session journals:** a `journals` collection, keyed by session id,
  Search on `text`, with `ticket` and `project` indexed.
  `control/recall.go` puts the summary where it calls `IndexRecall` today.
- **Ticket comments:** a ticket-side index, not a collection. Comments
  are core ticket data written in `Update`, and their index must change
  in that transaction. SQLite keeps an FTS5 table `fts_comments`, fed
  by the existing comment triggers. Postgres keeps a GIN `tsvector`
  index on `comments.body`. `Thread` does not use it. The recall RPC
  reads it through one ticket-side search method, which the step 3
  ticket PR adds to `Store` (20 methods) or folds into
  `List(Filter{Text: ...})` with matching comment ids. The step 3 PR
  picks one.
- **`jobshop recall`** queries comments, journals and learnings for the
  requested kinds, each with the same limit. It merges by rank within
  each source, interleaved round robin, because bm25 and `ts_rank`
  scores are not comparable across indexes. The project and ticket
  filters apply per source: an index for journals, a join for
  comments, a Go filter for learnings (a few hundred rows).

## Queries that need more than an index

A domain that needs more than equality on declared fields, a state set,
a time window and an order:

1. reads with `List` and filters or groups in Go when the bucket is small
   (`CountLearnings` globs; memory edits are 22 rows);
2. otherwise declares a field that makes it an equality filter;
3. only if neither works, adds a named query in its domain package,
   written with the same builder and run through the collection's
   dialect, and says why in its note.

No step of the plan needs 3.

## Migration of existing data

One goose migration per domain PR, in plan order: poll_cursors and
write_keys (empty), memory_edits (22), learnings (339), inbox (3,774).
The first one also creates `records` and its base indexes.

- **Up:** creates the collection's indexes (SQLite) or partition and
  indexes (Postgres) with the same DDL the registration emits. It copies
  rows with
  `INSERT INTO records SELECT '<name>', CAST(<key column> AS TEXT), ...`,
  mapping the columns into `value` with `json_object(...)` (SQLite) or
  `jsonb_build_object(...)` (Postgres). Integer ids become their decimal
  string. `created_at` is the old created or received time. It fills the
  collection's `fts_<name>` from the copied rows. Then it drops the old
  table, its indexes and its triggers.
- **Learnings** also delete their old `recall` rows. **Journals** move
  out of `recall` in the learnings PR. The comment index becomes
  `fts_comments` in the step 3 PR that owns comments. After that `recall`
  is empty and is dropped.
- **Down:** recreates the old table and copies back with `json_extract`
  or `->>`, casting keys back to integers where they were. It is
  lossless for every field that exists today. Learnings created after Up
  have non-numeric keys. Down gives them new integer ids and keeps the
  key in a `key` column, so nothing is lost.
- **Check:** before merge, run Up, Down, Up on an `sqlite3 .backup` copy
  of gasserver's database. Compare row counts and a per-row hash of the
  old columns, and paste the counts in the PR.

Registration on a migrated database is a no-op. A test asserts that the
migration's DDL matches what registration emits.

## Plan

1. This note (this PR). Stop for operator review.
2. `Collection` in `sqlstore` on both backends, with the `sqlq` builder,
   pgx on Postgres and `storetest` conformance cases. Then one PR per
   domain: poll_cursors and write_keys, memory_edits, learnings (with
   journals), inbox. Each deletes its old methods, queries and table.
3. Ticket side, one PR per group, as on #579: `List(Filter)`, `Thread`,
   `Apply`, `Update`, `Release(Outcome)`, `Expire`, `Slot`,
   `Resources`. Each moves its queries off sqlc. The last one removes
   sqlc and moves comment search to `fts_comments`.
4. Record the final method count here. The target is 19 on `Store`
   (20 if comment search stays a method) plus 9 on `Collection`.

## Decisions taken in this note

- Keys are natural string keys. Only learnings mint one, with
  `store.NewKey()`, so there is no sequence table.
- `memory_edits` is keyed by PR, which is already unique.
- Comment search is a ticket-side index (`fts_comments`), not a
  collection.
- Cross-collection recall interleaves per-source rankings, because the
  scores are not comparable.
