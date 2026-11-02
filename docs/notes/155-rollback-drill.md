# Self-deploy rollback drill

2026-11-02, #155. A marker file `/var/lib/jobshop-deploy/drill` makes the
next `jobshop-self-deploy` run treat the new controller as unhealthy
after a successful switch, so the rollback path runs without a bad
commit: switch back to the previous generation, rev written to `failed`,
same rev skipped later. The unit deletes the marker before it switches,
so a drill covers one deploy. A file, not a module option, so setting it
needs no rebuild.

`nix/module-test.nix` runs the generated deploy script against stubs
(gh, nixos-rebuild, nix-env, systemctl, the controller binary) for the
drill, the no-retry, a real unhealthy switch and a healthy deploy.
Runbook: `docs/verify/rollback-drill.md`.
