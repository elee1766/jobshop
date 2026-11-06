# Reload deadline (JOB-102)

2026-11-06.

`jobshop reload` on jobcity (58 pools) timed out at the CLI's 30s while
the controller logged `config reloaded` about 18s later, so the
workspace sync unit failed on reloads that applied.

A reload holds the tick lock, so it first waits for the running
supervisor pass, then loads the config and runs the isolation preflight
for every isolated pool. Preflight with `sandbox.dns_check` spawns pasta
and resolves a name (up to 15s each), and it ran once per pool, in
series.

As built:

- Pools whose preflight spec and dns_check are the same share one check;
  distinct checks run up to 8 at once. Each pool still gets its own
  quarantine from the shared report.
- `jobshop reload` waits up to 5m (`--timeout`), not the 30s every other
  call uses.
- The `config reloaded` log line has `took`, `lock_wait`, `load` and
  `preflight`, so a slow reload says where the time went.
