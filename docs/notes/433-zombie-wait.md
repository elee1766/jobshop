# 433: env refresh no longer waits on zombies

2026-11-04. Part of #433 (pkg/control -race time).

`stopSessionProcs` SIGTERMs processes carrying a session's env (a
detached agent daemon) after its host stops, then waits up to 5s for
them to go. It tested exit with `kill(pid, 0)`, which succeeds on a
zombie until its parent reaps it. A detached daemon's parent is init or
a subreaper, so a daemon that exited at once still held every env
refresh (deploy, restart) for the full 5s.

`procGone` now reads `/proc/<pid>/stat` and treats state Z or X as gone.
TestEnvRefreshStopsDaemon went from 6.2s to 3.1s alone.

Measured on gasserver (6 cores, load ~15), the `pkg/control -race`
wall time stays near 45-50s: the suite is CPU bound (sum of test
times ~500s at -parallel 6, and -parallel 24 is no faster). The
remaining CPU is about 22% SQLite parsing (first prepares per
connection and statements first seen inside a transaction, which run
unprepared) and 35% race checkptr in modernc. TestQuarantine (12 real
host spawns) stays on real hosts: an in-process host leaves the
controller's connection open after the agent exits, so the controller
never sees the crash.
