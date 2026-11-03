# Session tickets default to the creator's pool (#244)

2026-11-03. `ticket create` of a task from a session socket with no
`--pool` now gets the session's own pool, so a follow-up filed by a
worker is claimed instead of sitting unclaimed with reason `no_pool`.
That pool's prompt decides where it goes next (reroute with `ticket
update --pool`). Admin creates are unchanged: they may still leave the
pool empty and the orphan report (#84) still flags them.

Respawn env: a host is spawned with the controller's env (minus
per-start systemd variables) and PATH prefixed with the jobshop binary
dir, so gh, jobshop and GH_TOKEN from the unit survive a respawn. A unit
test covers the inherited env, the runbook check is
`docs/verify/respawn-env.md`. The loss seen after 655946a did not
reproduce on 2c04d34 (#216's gh wrapper).
