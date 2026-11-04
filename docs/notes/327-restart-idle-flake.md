# TestSessionRestartIdle flake (#327)

2026-11-04. The test reloaded a config that changes agent env and then
asserted the host spec still had the old env. Since #106 an idle runner
re-checks the env hash on every poll and refreshes the host by itself, so
under -race the refresh sometimes won and the assert at line 1127 failed.
The test now waits for that refresh (new spec env, new live host pid),
then runs `session restart` against the refreshed host and checks it
respawns again with claims kept and no crash counted.
