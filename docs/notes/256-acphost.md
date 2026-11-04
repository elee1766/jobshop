# #256: pkg/host replaced by acphost (2026-11-04)

- The session host is `github.com/elee1766/acphost` v0.1.0, a copy of the
  old `pkg/host`. `pkg/host` is deleted; callers import acphost as `host`.
- `pkg/hostspec` wraps `acphost.Spec` with the `conv_export` field. The
  JSON layout is unchanged, so spec.json written by older builds reads
  back and hosts started by them are adopted (same protocol 1, default
  `_jobshop/` method prefix and `jobshop-host-` GC root prefix).
- `jobshop session-host` passes `hostspec.HostOptions(dir)` to
  `acphost.Main`: an exporter on `telemetry.ConvExporter` when the spec
  sets conv_export. Dropped/truncated counts reach the controller as
  before through `_jobshop/export_stats`.
