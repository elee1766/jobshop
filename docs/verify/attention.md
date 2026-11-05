# Attention view (#185)

After a deploy that includes #185, on gasserver:

1. `JOBSHOP_ROOT=/var/lib/jobshop jobshop attention` returns, oldest
   first, every open `needs-human` ticket in
   `jobshop ticket list --label needs-human -s open`, and nothing else
   when no run is `needs_human` and no pool is quarantined. Healthy:
   `{"items": []}`.
2. No deploy item while healthy. A self-deploy rollback is one open
   needs-human ticket titled "self-deploy rolled back <rev>: ..."
   (`deploy:rollback`, #577), closed by the next good deploy; see
   `docs/verify/rollback-drill.md`.
3. With `services.jobshop.gateway.<name>.attentionDigest` set, the
   escalation room gets one digest thread at that time on a day the
   list is non-empty, and none on a healthy day (journal of
   `jobshop-gateway-<name>` has no `gateway: digest` WARN).
