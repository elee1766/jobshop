# Workspace sync during a switch (#475, 2026-11-05)

A self-deploy switch stops jobshop.service. If the workspace sync unit ran
in that window and the checkout moved, `jobshop reload` failed on the missing
admin socket, the switch reported the sync unit failed, and self-deploy
rolled back a green commit (17f4a6c).

Now the sync script skips the reload, logs `reload skipped` and exits 0 when
jobshop.service is not active, before or during the reload. The controller
loads packs on start, so nothing is lost. The sync unit is also ordered after
jobshop.service.

The `failed` state file blocks only the rev it names, so the next green
commit (this fix) supersedes 17f4a6c; no manual clearing is needed.
