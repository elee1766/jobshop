# Jobshop: ticket data model (draft)

Status: draft for discussion. Companion to `plan.md` section 5.

## Model

Everything is a ticket with a `kind` (task, run, step, session, message,
gate, batch, ...), connected by edges. Same idea as Gas City beads, with two
changes:

- Kinds are declared (YAML, built in or from packs). A declaration lists the
  kind's fields with types, required fields, indexes, and edge kinds. The
  store validates writes against it. Gas City has two undeclared kind layers
  (`issue_type` column + `gc.kind` metadata) and 188 untyped `gc.*` keys.
- Anything the engine schedules on is a real column (status, pool, holder,
  lease, fence, parent, revision). Kind-specific data lives in a typed JSON
  `fields` document, with expression indexes where declared.

SQL below targets SQLite and is written to port to Postgres (notes inline).
Timestamps are integer unix milliseconds.

## Tables

```sql
-- Core. One row per ticket of any kind.
CREATE TABLE tickets (
  id           TEXT PRIMARY KEY,            -- 'js-4f2a9c', random, prefixed
  project      TEXT NOT NULL,
  kind         TEXT NOT NULL,               -- must exist in kinds
  title        TEXT NOT NULL DEFAULT '',
  body         TEXT NOT NULL DEFAULT '',
  status       TEXT NOT NULL DEFAULT 'open'
               CHECK (status IN ('open','in_progress','closed')),
  resolution   TEXT,                        -- set on close: done, failed, cancelled, ...
  priority     INTEGER NOT NULL DEFAULT 2,  -- lower runs first
  pool         TEXT,                        -- queue it is routed to (assign)
  holder       TEXT,                        -- claimant (session id)
  lease_until  INTEGER,                     -- claim expires at
  fence        INTEGER NOT NULL DEFAULT 0,  -- bumped on every claim
  parent_id    TEXT REFERENCES tickets(id),
  defer_until  INTEGER,
  ephemeral    INTEGER NOT NULL DEFAULT 0,  -- no history rows, reaped after close
  fields       TEXT NOT NULL DEFAULT '{}',  -- JSON, validated against kind schema
  revision     INTEGER NOT NULL DEFAULT 0,  -- bumped on every write (CAS)
  created_by   TEXT NOT NULL DEFAULT '',
  created_at   INTEGER NOT NULL,
  updated_at   INTEGER NOT NULL,
  closed_at    INTEGER
);
-- Postgres: fields JSONB, booleans as BOOLEAN.

CREATE INDEX tickets_queue   ON tickets(project, pool, status, priority, created_at);
CREATE INDEX tickets_holder  ON tickets(holder) WHERE holder IS NOT NULL;
CREATE INDEX tickets_lease   ON tickets(lease_until) WHERE status = 'in_progress';
CREATE INDEX tickets_parent  ON tickets(parent_id);
CREATE INDEX tickets_kind    ON tickets(kind, status);

-- Typed relations between tickets.
CREATE TABLE edges (
  from_id    TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  to_id      TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  kind       TEXT NOT NULL,                 -- blocks, waits_for, tracks, delivered_to, ...
  blocking   INTEGER NOT NULL,              -- copied from the edge-kind declaration at insert
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

-- Append-only change log; doubles as the event stream.
CREATE TABLE history (
  seq        INTEGER PRIMARY KEY AUTOINCREMENT,  -- Postgres: BIGSERIAL
  ticket_id  TEXT NOT NULL,
  at         INTEGER NOT NULL,
  actor      TEXT NOT NULL,                 -- session id, 'controller', or human
  op         TEXT NOT NULL,                 -- create, update, claim, release, close, lease_expired, edge_add, ...
  revision   INTEGER NOT NULL,              -- ticket revision after this op
  diff       TEXT NOT NULL DEFAULT '{}'     -- JSON: {"status":["open","in_progress"],"fields.pr.state":[...]}
);
CREATE INDEX history_ticket ON history(ticket_id, seq);

-- Comments: append-mostly human/agent notes. Separate from history so
-- history can be compacted without losing them (decided 2026-11-01).
-- Gas City has no comment primitive; bd uses a comments table, our
-- sqlite-work fork packs them into one metadata key (gc.work.comments).
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
-- Adding a comment also writes a history row (op='comment_add', diff holds
-- the comment id only) so the change stream sees it.
-- Structured notes (handoffs, blocker reports) are a declared kind instead.

-- Kind and edge-kind declarations loaded from YAML. Stored so the
-- database is self-describing and old rows can be checked against the
-- schema version they were written under.
CREATE TABLE kinds (
  name       TEXT PRIMARY KEY,
  source     TEXT NOT NULL,                 -- 'builtin' or pack name
  version    INTEGER NOT NULL,
  schema     TEXT NOT NULL                  -- JSON Schema for fields, plus flags (ephemeral, statuses)
);
CREATE TABLE edge_kinds (
  name       TEXT PRIMARY KEY,
  source     TEXT NOT NULL,
  blocking   INTEGER NOT NULL
);
```

Live process state (PID, socket path, host log) is not stored here. It is
derived from `<data>/sessions/<id>/` at runtime. A `session` ticket holds
only the durable part.

## Kind declaration (YAML)

```yaml
kinds:
  - name: pr
    source: pr-delivery          # pack
    fields:
      repo:   { type: string, required: true }
      number: { type: integer, required: true }
      state:  { type: string, enum: [open, merged, closed], index: true }
      ci:     { type: string, enum: [pending, passing, failing] }

  - name: message
    ephemeral: false
    fields:
      to:      { type: string, required: true, index: true }  # session, pool or key
      from:    { type: string, required: true }
      interrupt: { type: boolean, default: false }

  - name: session
    fields:
      key:         { type: string, index: true, unique: true }   # e.g. matrix:<room>
      agent:       { type: string, required: true }
      acp_session: { type: string }
      current:     { type: string }                              # ticket id being worked

edge_kinds:
  - { name: blocks,    blocking: true }
  - { name: waits_for, blocking: true }
  - { name: tracks,    blocking: false }
```

`index: true` makes the store create an expression index:

```sql
CREATE INDEX tickets_f_pr_state ON tickets(json_extract(fields, '$.state'))
  WHERE kind = 'pr';
-- Postgres: ON tickets ((fields->>'state')) WHERE kind = 'pr'
CREATE UNIQUE INDEX tickets_f_session_key ON tickets(json_extract(fields, '$.key'))
  WHERE kind = 'session' AND status <> 'closed';
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
    WHERE e.from_id = t.id AND e.blocking = 1 AND b.status <> 'closed'
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
      WHERE e.from_id = t.id AND e.blocking = 1 AND b.status <> 'closed')
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
the lease was lost; the CLI reports that and the agent stops.

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
SET fields = json_set(fields, '$.state', :state),   -- Postgres: jsonb_set
    revision = revision + 1, updated_at = :now
WHERE id = :id AND revision = :expected_revision
RETURNING revision;
```

### Lease expiry sweep (controller tick)

```sql
UPDATE tickets
SET status = 'open', holder = NULL, lease_until = NULL,
    revision = revision + 1, updated_at = :now
WHERE status = 'in_progress' AND lease_until < :now
RETURNING id, holder;   -- write op='lease_expired' history rows for these
```

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
INSERT INTO edges (from_id, to_id, kind, blocking, created_at) VALUES (...);
INSERT INTO history (...) VALUES (...);
COMMIT;
```

### Reap ephemeral tickets

```sql
DELETE FROM tickets
WHERE ephemeral = 1 AND status = 'closed' AND closed_at < :now - :retention_ms;
```

### Timeline for `jobshop show`

```sql
SELECT at, actor, op, diff AS detail FROM history WHERE ticket_id = :id
  AND op <> 'comment_add'
UNION ALL
SELECT created_at, author, 'comment', body FROM comments
  WHERE ticket_id = :id AND deleted = 0
ORDER BY 1;
```

## Open questions

- Lease expiry: straight back to `open`, or a `stale` state the controller
  reviews first?
- Kind schema changes: how to migrate existing rows when a pack bumps a
  kind's version (validate-on-read, migrate-on-write, or explicit migration)?
- Should `blocking` live on the edge row (fast query, denormalized) or be
  joined from `edge_kinds` at query time?
