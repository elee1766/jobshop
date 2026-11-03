# Reload on a yaml-only NixOS switch (#210)

2026-11-03. `systemd.services.jobshop.reloadTriggers = [ configFile ]`:
the generated jobshop.yaml's store path goes into the unit as
`X-Reload-Triggers`, so a switch that changes only the yaml (e.g.
`livePackVars`) makes switch-to-configuration run `ExecReload` (SIGHUP,
the same as `jobshop reload`). Before, the unit file did not change, the
tmpfiles link moved, and the controller kept the old config until a
manual reload. A changed binary still changes `ExecStart` and restarts.

Check: `nix/module-test.nix` "yaml-only change reloads the controller".
Live: `docs/verify/nix-yaml-reload.md`.
