# Pack sync follows the unit's binary (#223)

2026-11-03. On gasserver the controller was down, so
`/proc/$MainPID/exe` did not exist, the sync script's deployed commit
was empty, and it fast-forwarded the packs checkout to origin. The new
packs needed `pack_prefix` (#184), which the running binary did not know,
so every restart failed on config load.

Built: `jobshop-workspace-<name>-sync` runs `version` on the unit's
ExecStart binary (the module's package path) instead of the live
process. If that commit is empty or not in the checkout, it logs a
warning and leaves the checkout where it is.

Not covered here: why the unit ran 53b6b27 while
`/var/lib/jobshop-deploy/deployed` said af9c010.
