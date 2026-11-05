# 530: journal search on VictoriaLogs

2026-11-05. Operator decision: journal search is an external VictoriaLogs,
the same for single node and HA. Conversation export over OTLP (#45) is
removed.

## Measured (2026-11-05, 40 gasserver journals, 57.3 MB, 132k frames)

| | Size | `status:failed` by session | "permission denied" | `tool_call_update` + "go test" |
|---|---|---|---|---|
| VictoriaLogs | 3.59 MB (16x) | 2.6 ms | 4 ms | 131 ms |
| Parquet zstd (DuckDB) | 2.28 MB (25x) | 6 ms | 431 ms (full scan) | 238 ms |

Parquet has no per-block token index, so text search scans every row.
Bleve's index was about the size of the raw data. Embedding
`lib/logstorage` brings global flags, `flag.Parse` and `Fatalf`.

## As built

- Config `journal: {victorialogs: {url, headers, retention_hint}, lag_max}`.
  Headers expand `${VAR}`. Absent: no shipping, and `journal.search` and
  `journal.backfill` fail with `journal search needs journal.victorialogs`.
  `session.journal` (#529) still reads the files.
- Shipping: a controller goroutine wakes every second, reads each open
  session's `journal.ndjson` (and, at start, sessions closed in the last
  24h) after `fields.shipped_seq`, and posts JSON lines to
  `/insert/jsonline` in batches up to 1 MB. It reads the file, not the ACP
  stream, so a restart or outage only delays rows. A failed batch stays
  queued; retries back off from 2s to 1m. The cursor is saved at most every
  5s, on close and on shutdown; after a crash up to 5s of frames are sent
  again.
- Row: `_time` frame time, `_msg` the frame redacted by `pkg/redact` (the
  #529 redaction), stream fields `project`, `pool`, `session`, fields `seq`,
  `from`, `method`, `update`, `status`, `tool` (the tool call's title),
  `ticket` (from the prompt's `_meta`), `run`, `workflow`, `step`.
- Duplicates: VictoriaLogs keeps both copies of a resent row. Collapse them
  with `| uniq by (session, seq)` (checked against v1.53.0).
- `journal.search {query, since, start, until, limit}`: LogsQL, forwarded to
  `/select/logsql/query`. Admin sees all projects. A pool needs the
  `journal-search` grant and gets `extra_stream_filters={project="<own>"}`,
  applied by VictoriaLogs outside the query, so `project:other` or
  `{project="other"}` in the query finds nothing. Limit 1000 by default,
  10000 max; `truncated` says it cut. A LogsQL syntax error is invalid input.
- `journal.backfill {since}` (admin): ships every closed session (closed
  within `since`, if set) from its cursor, 0 for old sessions, then flushes.
- CLI `jobshop journal search '<logsql>' [--since 2h] [--limit]` prints one
  JSON row per line; `jobshop journal backfill [--since 168h]`. Desk
  `GET /api/journal/search?q=&since=&start=&until=&limit=`.
- Metrics by pool: `jobshop_journal_ship_frames_total`, `_bytes_total`,
  `_errors_total`, `_lag_seconds` (oldest queued frame, 0 caught up).
  Attention `journal-lag` when lag passes `journal.lag_max` (default 10m).
- NixOS: `services.jobshop.journal.victorialogs.enable` runs nixpkgs
  `services.victorialogs` on `listenAddress` (default 127.0.0.1:9428) with
  `-retentionPeriod` (default 90d) and sets the url; or set `url` to an
  external instance. On gasserver use a port other than Gas City's.
- Removed: `pkg/telemetry/convexport.go`, `hostspec.ConvExport`, the
  `jobshop_conversation_export_*` metrics. `conversation_export` on a pool
  loads with a warning for one release, then becomes an error.
- Tests run a real `victoria-logs` (`VICTORIALOGS_BIN` or PATH; CI
  downloads the release and sets `JOBSHOP_REQUIRE_VICTORIALOGS`).

## Example queries

Failed tools by session:

    jobshop journal search 'update:tool_call_update status:failed | uniq by (session, seq) | stats by (session) count() failed'

A phrase in agent text:

    jobshop journal search --since 24h 'update:agent_message_chunk "permission denied"'

Top failing commands over 24h:

    jobshop journal search --since 24h 'status:failed tool:* | uniq by (session, seq, tool) | stats by (tool) count() n | sort by (n desc) | limit 10'

`uniq by` keeps only the fields it names, so name every field a later
pipe uses.
