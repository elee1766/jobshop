# In-process hosts for controller-logic tests (#361)

2026-11-04. Step 4 of #359.

`control.Options.Hosts` (a `HostRunner`: `Start(dir)`, `Stop(dir)`) replaces
the detached `session-host` process per session. Production leaves it nil.

In `pkg/control` tests, `poolRoot` sets it to `inProcHosts`, which runs
`acphost.Main(dir)` in a goroutine of the test binary and stops it with the
host's stop method instead of SIGTERM. The agent is still the fake-agent
child acphost starts; only the host spawn (about 1s each under `-race`) is
saved. The host writes the test's pid to `host.pid`, so cleanup skips it.

Tests about the process tree use `poolRootReal` (or `Hosts: realHosts{}`):
controller restart and adopt, crash and quarantine, sandbox, env refresh,
stuck and busy, worktree respawn, and tests that detect a respawn by a new
host pid. The list is in the comment on `realHosts` in `session_test.go`.
