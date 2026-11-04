# Ingest receiver unit (#344)

2026-11-04. NixOS module option `services.jobshop.ingest.enable`.

- `jobshop-ingest.service` runs `jobshop ingest serve` (#334, #339) as the
  jobshop user, on `settings.ingest.listen` (default 127.0.0.1:8740).
  Restarts when jobshop.yaml changes, since serve reads its config once.
- Public ingress (reverse proxy, path routing, tunnel) is host config,
  not part of this module. On gasserver it is Caddy behind Tailscale
  Funnel in gasserver.nix, routing `/jobshop/github/webhook` and
  `/jobshop/linear/webhook` to this unit.
