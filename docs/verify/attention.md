# Attention view (#185)

After a deploy that includes #185, on gasserver:

1. `JOBSHOP_ROOT=/var/lib/jobshop jobshop attention` returns, oldest
   first, every open `needs-human` ticket in
   `jobshop ticket list --label needs-human -s open`, and nothing else
   when no run is `needs_human` and no pool is quarantined. Healthy:
   `{"items": []}`.
2. `grep rollback_file /var/lib/jobshop/jobshop.yaml` shows
   `/var/lib/jobshop-deploy/failed`. If that file is newer than
   `deployed`, the list has a `rollback` item; otherwise none.
3. With `services.jobshop.gateway.<name>.attentionDigest` set, the
   escalation room gets one digest thread at that time on a day the
   list is non-empty, and none on a healthy day (journal of
   `jobshop-gateway-<name>` has no `gateway: digest` WARN).
