# Tracing check (JOB-105)

On a controller with `telemetry.traces_url` set to a VictoriaTraces (or any
OTLP) endpoint.

1. `jobshop reload`, then find the `rpc config.reload` trace. It has a
   `config reload` child with `config load`, `set resources`, one
   `preflight` per distinct isolation check (attribute `jobshop.pools`)
   and, when the controller was busy, `wait tickMu`.
2. Find a `supervisor tick` trace. Its children are `tick sweep`,
   `tick run_triggers`, ... `tick compress_journals`. Under
   `tick run_triggers`, a `trigger <name>` span per due trigger with
   `jobshop.trigger.result`; under that, `exec sh` for a trigger check
   (attributes `process.command`, `jobshop.exec.label`,
   `process.exit.code`) and `HTTP POST <host>` for a Linear read.
3. No span attribute holds a token, a webhook URL path, a query string,
   a script body or an env value. Search the traces for a known secret
   value: no hits.
4. On a slow tick, the slowest child names the step. The same step shows on
   `/metrics`:
   `histogram_quantile(0.95, sum by (le, step) (rate(jobshop_tick_step_duration_seconds_bucket[5m])))`,
   plus `jobshop_outbound_duration_seconds` by `kind` and `target`, and
   `jobshop_lock_wait_seconds` by `lock` and `for`.
5. On jobcity, the next tick over 20s names its slow child in the trace.
