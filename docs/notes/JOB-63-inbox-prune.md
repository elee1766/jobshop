# Inbox prune (JOB-63, 2026-11-06)

- `ingest.retention` defaults to `7d` (was 14d). The hourly controller
  pass deletes dispatched, denied and failed items whose `updated_at` is
  older than now minus the retention. Pending, admitted and held items
  stay.
- Deletes run in batches of 200 keys, each its own short write
  transaction. The log line `inbox pruned items=<n>` gives the count.
- Dedup by delivery id and natural key only covers stored items, so a
  redelivery older than the retention is taken as new. GitHub redelivers
  within hours.
- The controller does not VACUUM. To reclaim space once after the first
  large prune, stop the controller and run `sqlite3 <db> VACUUM`.
