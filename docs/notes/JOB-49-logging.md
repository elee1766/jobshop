# JOB-49: logging

2026-11-06

- `serve`, `desk`, `gateway`, `ingest serve` and `ingest work` take
  `--log-level debug|info|warn|error` and `--log-format text|json`
  (env `JOBSHOP_LOG_LEVEL`, `JOBSHOP_LOG_FORMAT`). One constructor:
  `pkg/cli/log.go`. No `log:` block in jobshop.yaml yet: the flags and
  env cover the NixOS module (it can set the env).
- Libraries never fall back to `slog.Default()`. A nil logger in
  `control.Options` or `ingest.Handler` discards. Controller tests log
  to `t.Output()` (`testLog`), so output shows only for a failing test
  or with `-v`.
- `pkg/logctx` carries ids in a context: `logctx.With(ctx, "run", id)`.
  Its handler adds them to every record logged with that ctx. The
  controller wraps its handler with it outermost, so stderr, the OTel
  bridge and the incident log all see them. Session runners put
  `session` and `pool` in their ctx.
- Log calls use `XContext(ctx, ...)` wherever a ctx is in scope, so the
  OTel bridge sets trace_id/span_id. sloglint (`context: scope`,
  `no-global: all`) keeps it that way.
