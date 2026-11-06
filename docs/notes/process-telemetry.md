# Process telemetry

2026-11-06. Every long-running command (`serve`, `desk`, `ingest serve`,
`ingest work`, `gateway`) builds its logger and tracer with
`telemetry.Start`, through `startTelemetry` in `pkg/cli`. It reads the
`telemetry:` section of `jobshop.yaml` once at start, logs text to stderr,
and with an OTLP logs URL also exports each record. Records carry
`process=<command>`. Changing the export needs a restart.

The controller no longer builds its own providers: `control.Options`
takes `Log` and `Tracer` from the caller. Without a tracer it makes spans
with no exporter, so traceparent still reaches agents.
