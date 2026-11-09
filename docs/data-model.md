# Jobshop: ticket data model

Companion to `plan.md` (sections 3 and 6).

## Model

Everything is a ticket with a `kind`, connected by edges. Same idea as Gas
City beads, with these differences (decided 2026-11-01):

- **Kinds and edge kinds are defined by the engine, in Go.** `task`,
  `session`, `message`, `run`, `step`, `batch`, and later control kinds. Each has a
  Go struct for its kind-specific fields, validated on write. Gas City has
  two undeclared kind layers (`issue_type` column plus `gc.kind` metadata)
  and 188 untyped `gc.*` keys; Jobshop has one `kind` and typed fields.
- **Packs do not declare schemas.** Packs (agents, prompts, workflows) get
  one free-form `meta` JSON document per ticket for their own data. The
  engine never reads or validates it. As in Gas City, packs can add
  behavior without changing the core; unlike Gas City, pack data cannot
  mix with engine data. A pack that needs engine behavior becomes a
  Jobshop change and a new typed field.
- **Anything the engine schedules on is a column** (status, pool, holder,
  lease, fence, parent, revision). Kind-specific engine data lives in
  `fields` (typed per kind in Go); pack data lives in `meta`.

SQL below targets SQLite (`modernc.org/sqlite`) and is written to port to
Postgres (notes inline). Timestamps are integer unix milliseconds.

## Tables

```sql
-- Core. One row per ticket of any kind.
CREATE TABLE tickets (
  id           TEXT PRIMARY KEY,            -- '<project prefix>-<sonyflake base32>', e.g. 'js-0031w02j0r012'; new sessions '<prefix>-s-<sonyflake>' (#615)
  project      TEXT NOT NULL,
  kind         TEXT NOT NULL,               -- engine-defined: task, session, message, run, step, ...
  title        TEXT NOT NULL DEFAULT '',
  body         TEXT NOT NULL DEFAULT '',
  status       TEXT NOT NULL DEFAULT 'open'
               CHECK (status IN ('open','in_progress','closed')),
  resolution   TEXT NOT NULL DEFAULT '',     -- set on close: done, failed, cancelled, ...
  priority     INTEGER NOT NULL DEFAULT 2,  -- lower runs first
  pool         TEXT NOT NULL DEFAULT '',     -- queue it is routed to (assign)
  holder       TEXT NOT NULL DEFAULT '',     -- claimant or pre-assignee (session id)
  lease_until  INTEGER NOT NULL DEFAULT 0,  -- claim expires at (backstop only); 0 = none
  fence        INTEGER NOT NULL DEFAULT 0,  -- bumped on every new claim
  parent_id    TEXT REFERENCES tickets(id),
  defer_until  INTEGER NOT NULL DEFAULT 0,
  ephemeral    INTEGER NOT NULL DEFAULT 0,  -- no history rows, reaped after close
  fields       TEXT NOT NULL DEFAULT '{}',  -- engine data, typed per kind in Go
  meta         TEXT NOT NULL DEFAULT '{}',  -- pack/user data, never read by the engine
  revision     INTEGER NOT NULL DEFAULT 0,  -- bumped on every write (CAS)
  created_by   TEXT NOT NULL DEFAULT '',
  created_at   INTEGER NOT NULL,
  updated_at   INTEGER NOT NULL,
  closed_at    INTEGER NOT NULL DEFAULT 0
);
-- Empty string / 0 instead of NULL for optional scalars (decided 2026-11-01):
-- simpler generated Go, no three-valued comparisons. parent_id stays NULL
-- because it is a foreign key. The SQL below is illustrative; the source of
-- truth is pkg/storage/store/sqlite/migrations and queries.
-- Postgres: fields/meta JSONB, booleans as BOOLEAN.

CREATE INDEX tickets_queue   ON tickets(project, pool, status, priority, created_at);
CREATE INDEX tickets_holder  ON tickets(holder) WHERE holder <> '';
CREATE INDEX tickets_lease   ON tickets(lease_until) WHERE status = 'in_progress';
CREATE INDEX tickets_parent  ON tickets(parent_id);
CREATE INDEX tickets_kind    ON tickets(kind, status);

-- Engine-owned field indexes. Fixed in migrations, since kinds are in Go.
-- Hot fields keys are copied to plain f_* columns by the store on every
-- write (#405, docs/notes/405-portable-sql.md), so no index or query
-- needs a JSON function.
CREATE INDEX tickets_message_to ON tickets(f_to)
  WHERE kind = 'message';
CREATE UNIQUE INDEX tickets_session_key ON tickets(f_key)
  WHERE kind = 'session' AND status <> 'closed' AND f_key IS NOT NULL;

-- Typed relations between tickets. Edge kinds are engine-defined:
--   blocks, waits_for  -> block readiness
--   tracks             -> informational (batches, related work)
CREATE TABLE edges (
  from_id    TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  to_id      TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  kind       TEXT NOT NULL CHECK (kind IN ('blocks','waits_for','tracks')),
  created_at INTEGER NOT NULL,
  PRIMARY KEY (from_id, to_id, kind)
);
CREATE INDEX edges_to ON edges(to_id, kind);
-- Convention: from_id is the dependent. "A blocks on B" = (A, B, 'blocks').

CREATE TABLE labels (
  ticket_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  label     TEXT NOT NULL,
  PRIMARY KEY (ticket_id, label)
);
CREATE INDEX labels_label ON labels(label);

-- Append-only audit log of ticket changes (who, when, before/after).
-- Not an event stream: nothing tails it, and it may be compacted. Live
-- change notifications come from the controller (admin socket), since all
-- writes go through it.
CREATE TABLE history (
  seq        INTEGER PRIMARY KEY AUTOINCREMENT,  -- Postgres: BIGSERIAL
  ticket_id  TEXT NOT NULL,
  at         INTEGER NOT NULL,
  actor      TEXT NOT NULL,                 -- session id, 'controller', or human
  op         TEXT NOT NULL,                 -- create, update, claim, release, close, lease_expired, edge_add, ...
  revision   INTEGER NOT NULL,              -- ticket revision after this op
  diff       TEXT NOT NULL DEFAULT '{}'     -- JSON: {"status":["open","in_progress"],"fields.state":[...]}
);
CREATE INDEX history_ticket ON history(ticket_id, seq);

-- Comments: human/agent notes. Separate from history so history can be
-- compacted without losing them (decided 2026-11-01). Gas City has no
-- comment primitive; bd uses a comments table, our sqlite-work fork packs
-- them into one metadata key (gc.work.comments).
CREATE TABLE comments (
  id         INTEGER PRIMARY KEY,           -- Postgres: BIGSERIAL
  ticket_id  TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  author     TEXT NOT NULL,                 -- session id or human
  body       TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  edited_at  INTEGER,
  deleted    INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX comments_ticket ON comments(ticket_id, id);
```

Live process state (PID, socket path, host log) is not stored here. It is
derived from `<root>/sessions/<id>/` at runtime. A `session` ticket holds
only the durable part.

## Engine kinds (Go)

Illustrative; the field structs live in the store package.

```go
type MessageFields struct {
    To        string `json:"to"`        // session id, "pool:<name>", or session key
    From      string `json:"from"`
    Interrupt bool   `json:"interrupt,omitempty"`
}

type SessionFields struct {
    Key        string `json:"key,omitempty"` // e.g. matrix:<room>; unique while open
    Agent      string `json:"agent"`
    ACPSession string `json:"acp_session,omitempty"`
    Current    string `json:"current,omitempty"` // ticket id being worked
}

// Runs (#89). Steps are the run's children; step i+1 blocks on step i.
type RunFields struct {
    Workflow string            `json:"workflow"`
    Pack     string            `json:"pack,omitempty"`
    Vars     map[string]string `json:"vars,omitempty"`
    Key      string            `json:"key,omitempty"`   // one open run per workflow and key
    State    string            `json:"state"`           // running, needs_human, done, cancelled
}

type StepFields struct {
    Run      string `json:"run"`
    Step     string `json:"step"`
    Index    int    `json:"index"`
    Retries  int    `json:"retries"`
    Failures int    `json:"failures"`
}

// Batches (#169, 2026-11-03). No pool, never claimed; member runs are its
// children. Order pair [a, b]: b's first step is blocked by a's last.
type BatchFields struct {
    Key    string      `json:"key"`    // one open batch per project and key
    Issues []string    `json:"issues"`
    Order  [][2]string `json:"order,omitempty"`
    State  string      `json:"state"`  // open, running, needs_human, done, failed
}
```

Adding a field is a Go change. Adding an index on it is a migration.
Packs that only need to stash data use `meta`:

```sql
UPDATE tickets SET meta = json_set(meta, '$.gx.linear_id', 'ENG-123'),
  revision = revision + 1, updated_at = :now
WHERE id = :id AND revision = :expected_revision;
```

## Queries

### Ready frontier

Open, not deferred, nothing blocking it still open.

```sql
SELECT t.*
FROM tickets t
WHERE t.project = :project
  AND t.status = 'open'
  AND (t.defer_until IS NULL OR t.defer_until <= :now)
  AND NOT EXISTS (
    SELECT 1 FROM edges e
    JOIN tickets b ON b.id = e.to_id
    WHERE e.from_id = t.id AND e.kind IN ('blocks','waits_for')
      AND b.status <> 'closed'
  )
ORDER BY t.priority, t.created_at;
```

### Claim next (the hook)

Three tiers in one statement: my in-progress work, then tickets pre-assigned
to me, then the pool queue. Atomic because it is a single UPDATE.

```sql
UPDATE tickets
SET status = 'in_progress',
    holder = :session,
    lease_until = :now + :lease_ms,
    fence = fence + 1,
    revision = revision + 1,
    updated_at = :now
WHERE id = (
  SELECT t.id FROM tickets t
  WHERE t.project = :project
    AND (
      (t.status = 'in_progress' AND t.holder = :session)             -- tier 1: resume
      OR (t.status = 'open' AND t.holder = :session)                 -- tier 2: pre-assigned
      OR (t.status = 'open' AND t.holder IS NULL AND t.pool = :pool) -- tier 3: queue
    )
    AND (t.defer_until IS NULL OR t.defer_until <= :now)
    AND NOT EXISTS (
      SELECT 1 FROM edges e JOIN tickets b ON b.id = e.to_id
      WHERE e.from_id = t.id AND e.kind IN ('blocks','waits_for')
        AND b.status <> 'closed')
  ORDER BY
    CASE WHEN t.status = 'in_progress' THEN 0
         WHEN t.holder = :session THEN 1 ELSE 2 END,
    t.priority, t.created_at
  LIMIT 1
)
RETURNING id, fence, revision;
-- Plus one history row (op='claim') in the same transaction.
-- Postgres: add FOR UPDATE SKIP LOCKED to the subquery for concurrent claimers.
```

### Heartbeat and close (fenced)

Succeed only if the caller still holds the claim. Zero rows affected means
the claim was lost; the CLI reports that and the agent stops.

```sql
UPDATE tickets
SET lease_until = :now + :lease_ms, updated_at = :now
WHERE id = :id AND holder = :session AND fence = :fence AND status = 'in_progress';
-- No revision bump and no history row: heartbeats are high-churn.

UPDATE tickets
SET status = 'closed', resolution = :resolution, closed_at = :now,
    lease_until = NULL, revision = revision + 1, updated_at = :now
WHERE id = :id AND holder = :session AND fence = :fence AND status = 'in_progress'
RETURNING revision;
```

### Compare-and-swap update

```sql
UPDATE tickets
SET fields = json_set(fields, '$.current', :ticket),   -- Postgres: jsonb_set
    revision = revision + 1, updated_at = :now
WHERE id = :id AND revision = :expected_revision
RETURNING revision;
```

### Releasing claims (decided 2026-11-01)

Primary path: the controller knows when a session dies (host gone, agent
exited, session stopped) and releases that session's claims immediately.

```sql
UPDATE tickets
SET status = 'open', holder = NULL, lease_until = NULL,
    revision = revision + 1, updated_at = :now
WHERE holder = :dead_session AND status = 'in_progress'
RETURNING id;   -- history op='release', actor='controller', reason in diff
```

Backstop: the lease. The controller sweeps expired leases on every tick (not
lazily by the next claimer, as in Gas City). This catches a live session
that stopped heartbeating, e.g. a wedged agent.

```sql
UPDATE tickets
SET status = 'open', holder = NULL, lease_until = NULL,
    revision = revision + 1, updated_at = :now
WHERE status = 'in_progress' AND lease_until < :now
RETURNING id, holder;   -- history op='lease_expired'
```

Both go straight back to `open`. Reopen counts come from history; retry
limits are a workflow concern (phase 2). Gas City for comparison: a 30-minute
lease, reclaimed only when a claimer with `AutoReclaimStaleClaims` (off by
default) trips over it, plus `releaseOrphanedPoolAssignments` for dead
sessions.

### Inbox: messages for a session or its pool

```sql
SELECT * FROM tickets
WHERE kind = 'message' AND status = 'open'
  AND json_extract(fields, '$.to') IN (:session, 'pool:' || :pool)
ORDER BY json_extract(fields, '$.interrupt') DESC, created_at;
```

### Find the session for a chat key

```sql
SELECT * FROM tickets
WHERE kind = 'session' AND status <> 'closed'
  AND json_extract(fields, '$.key') = 'matrix:!abc:example.org';
```

### Run progress (workflow children)

```sql
SELECT status, count(*) FROM tickets
WHERE parent_id = :run_id GROUP BY status;
```

### Graph apply (create a run)

One transaction: insert the run ticket, its step tickets with `parent_id`,
the `blocks` edges between steps, and history rows. IDs are generated
client-side so edges can reference steps before insert.

```sql
BEGIN IMMEDIATE;               -- Postgres: BEGIN
INSERT INTO tickets (...) VALUES (...);   -- run + steps
INSERT INTO edges (from_id, to_id, kind, created_at) VALUES (...);
INSERT INTO history (...) VALUES (...);
COMMIT;
```

### Timeline for `jobshop show`

```sql
SELECT at, actor, op, diff AS detail FROM history WHERE ticket_id = :id
UNION ALL
SELECT created_at, author, 'comment', body FROM comments
  WHERE ticket_id = :id AND deleted = 0
ORDER BY 1;
```

### Reap ephemeral tickets

```sql
DELETE FROM tickets
WHERE ephemeral = 1 AND status = 'closed' AND closed_at < :now - :retention_ms;
```

## Pack data and labels (decided 2026-11-01)

- `meta` is grouped by pack name: `{"gx": {"linear_id": "ENG-123"}}`.
  The controller rejects top-level keys that are not a known pack name
  (built-ins `github`, `trigger`, `run`, `gateway`, plus `meta_packs:` in jobshop.yaml);
  the store only checks the shape (#29, 2026-11-01). Gas City
  has only a flat `map[string]string` with an optional dot-prefix
  convention, and in the live city engine keys (`close_reason`, `state`,
  `session_id`, ...) and pack keys (`merge_strategy`, `kn.*`) share one
  namespace; about half of engine keys lack the `gc.` prefix.
- Labels are user-only. The engine never reads or writes them. Gas City
  uses labels as engine state (`gc:session`, `gc:wisp`, `hold:mayor`,
  `order-run:*`, `read`); in Jobshop those become `kind`, `ephemeral`,
  edges, or typed fields.

## Defaults (decided 2026-11-01)

- Ticket lease: 30 minutes (same as Gas City). Heartbeats renew it; the
  CLI heartbeats on a fraction of the lease (e.g. every 5 minutes) while a
  session holds a claim. Configurable per pool.

## Open questions

- None currently.
