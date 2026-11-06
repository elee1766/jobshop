# Agent shell tools after a controller restart (#780)

2026-11-06. Not a regression of #253. Sessions respawned after the
02:43 restart had the full spec env in their tool shells: `PATH` equal
to the host spec's (jobshop's dir first, then the controller unit's
PATH), `JOBSHOP_SESSION` and `__NIXOS_SET_ENVIRONMENT_DONE=1` set.

Cause: the controller unit's PATH no longer has `which`, `awk`, `ps`,
`pgrep` or `setsid`. JOB-23 (#720) cut the module's default PATH to what
jobshop itself execs (git, bash, coreutils, diffutils, openssh, plus
the unit's coreutils/findutils/grep/sed). NixOS keeps those other tools
in /run/current-system/sw/bin, which a unit's PATH does not include.
Fresh and respawned sessions alike miss them. The ones seen right after
a restart failed because their agents ran
`PATH=/usr/bin:/bin:$(dirname $(which go)) go test`: the missing `which`
emptied the PATH and then `go` was not found either. The agents read this
as a broken shell, ended two turns unclosed and were reaped `stuck`.

Fix: the module's base tools add gawk, which, procps, util-linux, gnutar
and gzip, so every agent shell has them on PATH. Hosts list anything
beyond that in `services.jobshop.path`.

Tests: a module-test check that the NixOS and home-manager controller
PATH carries them. `TestEnvRefresh` now also checks that an agent
respawned after a controller restart gets the session PATH and
`JOBSHOP_SESSION` in a `bash -c` child.
