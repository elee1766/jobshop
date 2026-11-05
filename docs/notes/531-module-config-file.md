# NixOS module: user-managed jobshop.yaml (#531)

2026-11-05.

Two modes, chosen by `services.jobshop.configFile`:

- **null (default):** as before. `settings` plus the module's keys are
  generated into `jobshop.yaml`, linked into `dataDir` on each switch;
  a switch that changes it reloads the controller.
- **a path:** the user owns the file. `dataDir/jobshop.yaml` links to it
  (no link when the path is `dataDir/jobshop.yaml` itself) and the module
  never writes it. `settings` must be empty (assertion).

What the module still owns goes to `dataDir/jobshop.module.yaml`, linked
on each switch: `identities`, `gateway`, live packs (`packs`),
`attention.rollback_file` (self-deploy), `defaults.sandbox` and
`shared_tools` (sandbox), and `telemetry`. The user's file pulls it in:

```yaml
include: [jobshop.module.yaml]
projects:
  js: {identity: elee-agent}
```

`include` is a new core config key: a list of files (relative to the
config's directory) merged before the file's own keys. Maps merge per
key with the including file winning; `packs`, `shared_tools`,
`prompt_rules` and `hooks` append, includes first. Included files cannot
include others. An explicit key rather than a `jobshop.d/*.yaml` lookup,
so the root file says where every key comes from and a stray file in a
directory cannot change config.

Reload: `jobshop-config.path` watches the user's file and
`jobshop.module.yaml`. On a change `jobshop-config.service` runs
`jobshop config check --file dataDir/jobshop.yaml` and, if it passes,
`systemctl try-reload-or-restart jobshop`. A failing check logs
"keeping the running config" and does not reload; a later controller
restart with a broken file uses the last good config (#227).

Self-deploy's config check already loads `dataDir/jobshop.yaml`, so it
follows the link and the include.
