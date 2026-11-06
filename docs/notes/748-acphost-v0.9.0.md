# #748: acphost v0.9.0 (2026-11-06)

go.mod moves acphost from v0.8.5 to v0.9.0.

What jobshop adopts:

- Isolation preflight at start and reload (#728) passes the pool's
  sandbox `deny`, `dns` and `resolv_conf` to `acphost.Preflight`. With
  `deny` set, Preflight now checks that nft is on PATH, so a missing
  nftables quarantines the pool at load instead of failing every spawn.
- A pool with `sandbox.dns_check` adds `acphost.WithDNSCheck(name)` to
  that preflight. It spawns pasta, so it only runs there, not per
  session. The per-pool check at spawn (`checkSandboxDNS`, JOB-20) stays
  for a host that changes between reloads.
- `Options.Preflight` takes `...acphost.PreflightOption` like
  `acphost.Preflight`.
- Fix hints were already shown: the quarantine reason is the report's
  `String()`, which v0.9.0 gives a `fix:` line per failure.

Not adopted: `acphosttest.Start`. The in-process host runner in
`pkg/control/inproc_test.go` stays, for the reasons in its comment
(#723): the controller calls Start off the test goroutine, has already
written the spec, and polls the socket itself.

JournalTail and inotify journal follow need no change.
