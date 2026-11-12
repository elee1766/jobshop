# Host stop waits through proc.Detached.Wait (#1229)

2026-11-12. Follow-up to acphost docs/notes/210-host-on-proc.md.

- The spawn half was done in #904: session hosts start with
  `acphost.StartDetached`, and `cli.Main` (and the pkg/orchestration/control
  `TestMain`) call `acphost.MaybeChild(session.HostOptions("")...)`. No
  `Setsid` spawn of a host is left; the only `Setsid` in the tree is the
  fake agent's daemon helper in `session_test.go`.
- `session.StopDetached` now waits with `proc.Open(dir).Wait(ctx)` under a
  `HostStopWait` timeout instead of jobshop's own flock on
  `supervisor.lock` (`waitHostExit`, removed). The own waiter existed only
  because `Detached.Wait` closed the lock file while its flock goroutine
  still used the fd (acphost#227); acphost v0.24.1 has the fix, its
  goroutine owns the file.
- The Wait result is ignored: `Running()` after it decides the SIGKILL, as
  before. A Wait that times out leaves its goroutine until the host exits,
  which the SIGKILL brings.
- StartDetached had every option jobshop needs (`InCgroup`, `InScope`), so
  no acphost issue was filed.

Tests: `TestDetachedHostSpawnStop` (start, connect, stop, exit recorded),
`TestStopHostKills` (SIGTERM ignored, SIGKILL after the wait),
`TestNoSupervisorPid`.
