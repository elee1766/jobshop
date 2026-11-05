# Observability (#14, #15)

- `telemetry.metrics_listen` serves `/metrics`: `jobshop_sessions` (by
  pool and state idle/working/backoff), `jobshop_session_restarts_total`,
  `jobshop_tickets_ready`, `jobshop_claims_total`, `jobshop_closes_total`,
  `jobshop_lease_expiries_total`, `jobshop_rpc_duration_seconds`,
  `jobshop_turns_total` (by pool and `jobshop_turn_stop_reason`: the ACP
  stop reason, or `error` when the prompt failed; counted once per turn
  after rate-limit retries, not on shutdown),
  `jobshop_sessions_ended_total` (by pool and `jobshop_resolution`: idle,
  crashed, stuck, ...; counted in `runner.finish`, #71, 2026-11-01). No
  ticket or session ids on metrics.
- `telemetry.otlp_endpoint` sends traces and slog logs over OTLP/HTTP.
  Spans: `supervisor tick`, `rpc <method>`, session lifecycle. Each session
  has one trace; its traceparent is stored on the session and passed to
  the agent as `JOBSHOP_TRACEPARENT`, and the CLI propagates it.
- `otlp_traces_endpoint` and `otlp_logs_endpoint` (full URLs) override
  the base per signal, and either alone turns that signal on.
  `otlp_headers` go on both; `${VAR}` in values expands from the
  controller env at startup, never at parse, so `config show` prints the
  placeholder. The nix module exposes these as `services.jobshop.telemetry.*`
  plus `environmentFile` for the tokens (#44).
- Grafana dashboard: `deploy/grafana/jobshop.json` (sessions by pool and
  state, backoff, restarts, ready queue, claims and closes, rpc p95,
  turns by stop reason). A test scrapes `/metrics` and fails if a panel
  names a series or label that is not exported.
- `deploy/verify-telemetry.sh` (POSIX sh, curl and jq, `--help`) checks a
  deployment: `/metrics` serves `jobshop_*`, the VictoriaMetrics target is
  up, `jobshop_sessions` and `jobshop_tickets_ready` exist per pool, and
  VictoriaTraces has a `session` trace in the last hour. Counters that
  have not incremented yet are listed, not failed. CI shellchecks it
  (`shellcheck -s sh`, #65).
  Runbook: `docs/verify/telemetry.md` (#55).
- `deploy/verify-sessions.sh` (POSIX sh, jq) fails when any session closed
  `crashed` or `stuck` within `--since` (default 24h), printing id, pool
  and end reason. Used in `docs/verify/baseline.md` (#77).
- Planned: store transaction spans.
- Conversation export (decided 2026-11-01, #45; removed 2026-11-05 by
  #530, replaced by journal shipping to VictoriaLogs, see
  `docs/notes/530-journal-search.md`). It made hosts emit OTLP log
  records `jobshop.conversation.<event>` per pool with
  `conversation_export: true`; that key now loads with a warning. Its
  redaction moved to `pkg/redact`.
- gasserver backends: VictoriaMetrics, VictoriaLogs, VictoriaTraces, all
  OTLP/HTTP, no collector.
