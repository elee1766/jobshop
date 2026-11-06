# Agent shell probe before spawn (JOB-67)

2026-11-06. Before a session host starts, the controller checks the env
the agent's tool shells inherit (controller env, overridden by the spec
env): `JOBSHOP_SESSION` is set and `jobshop`, `gh`, `git`, `sh`, `env`,
`dirname`, `which` and `awk` resolve on its PATH. A miss fails the spawn
with a permanent error, `agent shell probe: missing <names> on the agent PATH`.

A permanent spawn error ends the session before it claims anything, so
the ticket stays ready and no step attempt is counted. The pool is
quarantined (pool attention item, one needs-human escalation) until
`jobshop pool unquarantine` or a reload. This is what #780 needed: there,
sessions started after a restart without `which`/`awk` and every step
they took ended stuck.

Sandboxed pools are skipped: their commands resolve inside the sandbox.
The probe runs on every spawn, including respawns. `Options.ShellProbe`
holds the list; `jobshop serve` sets `control.DefaultShellProbe`, tests
leave it empty.
