# Session hosts start through acphost.StartDetached (#904)

2026-11-06.

- The controller starts a session host with `acphost.StartDetached`: this
  binary re-exec'd (argv[0] `acphost-host`) in its own session, with
  `acphost.InCgroup(<sessions>/hosts)` when the sessions cgroup is set up
  (JOB-91). If that cgroup cannot be opened the start is retried without
  it and the host stays in the controller's cgroup, as before. The call
  returns once the agent started; an agent setup failure comes back as
  `*acphost.SpawnError` and restarts the session (counting toward the
  crash limit), as when `connect` found it.
- `cli.Main` and the pkg/control `TestMain` call
  `acphost.MaybeChild(hostspec.HostOptions("")...)` instead of
  `proc.MaybeChild`. The hidden `jobshop session-host` command is gone:
  nothing starts it.
- `Options.HostExe` no longer picks the host binary. It stays for what
  sessions and command steps see (PATH, `JOBSHOP_BIN`, the sandbox bind,
  background job filtering).
- Liveness and the host pid come from `proc.Open(dir)`: `supervisor.pid`,
  and `supervisor.lock` held while the host runs. `stopHost` sends
  SIGTERM through `Detached.Stop`, waits on the lock up to 10s, then
  SIGKILLs the pid if it still runs. The wait is jobshop's own flock on
  `supervisor.lock` (`waitHostExit`), not `Detached.Wait`, which races
  when its timeout wins (acphost#227).
- Legacy hosts: a host started before acphost v0.17 has no
  `supervisor.pid`. One such host was still running on gasserver when
  this changed, so when `proc.Open` finds no `supervisor.pid` the
  controller falls back to `host.pid`/`host.lock` for liveness, pid and
  stop. Remove the fallback (pkg/control/hostproc.go) once no such host
  is left, alongside acphost#222.
