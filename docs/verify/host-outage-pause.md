# Host outage pause on gasserver (JOB-194)

2026-11-10. After deploying, on a host with a `github-token` identity
whose `token_file` is a command.

1. Break the provider: make the token command fail (for example point it
   at a missing file). From a dev session run
   `jobshop identity token <owner/repo>`. It fails.
2. `jobshop pool list` shows the pools using the identity under `outages`
   with the error, `since` and `next_probe`. `jobshop attention` has one
   `outage` item for the identity, not one per step.
3. Create a ticket in a paused pool: no session claims it, no session
   starts, and `jobshop ticket show` of a step shows failures 0 and no
   `needs-human` escalation appears.
4. Fix the provider. Within `OutageProbe` (1m) the `outages` entry and the
   attention item go, and the held step is claimed with no undefer.
5. `Probe now` on the attention item (Desk) runs the probe on the next
   pass instead of waiting.
