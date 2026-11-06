# Telemetry resource and instance (JOB-106, 2026-11-06)

Jobshop sets its own OTel resource, so the collector no longer has to
tag instances by receiver port. `pkg/telemetry.NewResource` builds one
resource per process; `startTelemetry` (serve, desk, gateway, ingest)
passes it to the trace and log providers, and serve hands it to the
controller's meter provider, so `/metrics` `target_info` carries it too.

Attributes:

- `service.name=jobshop`
- `service.version`: the build version; the commit is added as
  `+<commit>` unless the version has it already (the Nix build does).
- `jobshop.commit`: the commit, when known.
- `service.instance.id` and `jobshop.instance`: `telemetry.instance`,
  or by default the root dir's base name, or `hostname:<root>` when
  the base name is `jobshop` (both module default dirs are).
- `telemetry.resource` (map): extra attributes such as
  `deployment.environment`. Values expand `${VAR}` from the process
  environment at startup, like `otlp_headers`.

Precedence, lowest first: jobshop's own attributes, then
`telemetry.resource`, then `OTEL_RESOURCE_ATTRIBUTES` and
`OTEL_SERVICE_NAME` (`resource.WithFromEnv`). The environment wins so
one process can be relabeled without a config change. A malformed
`OTEL_RESOURCE_ATTRIBUTES` keeps the pairs that parsed.

The NixOS and home-manager modules have
`services.jobshop.telemetry.instance` (default null: jobshop's default).
`telemetry.resource` goes through `settings.telemetry.resource`.

The resource is read at startup: changing it needs a restart, like the
OTLP endpoints. Once the hosts set `instance`, the gasserver collector's
`transform/selfdev` and `resource/jobcity` processors can go (host
change).
