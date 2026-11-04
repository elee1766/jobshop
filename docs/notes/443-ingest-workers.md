# Ingest split mode (#443)

2026-11-04.

- `ingest.workers: external` stops in-process admission. Answered holds
  are still released and admitted rows still dispatched by the controller.
- `jobshop ingest work` (admin socket) loops: `ingest.claim` leases the
  oldest pending row to holder `worker:<id>` with the hook list and
  lease length; the worker runs the hooks in the root; `ingest.admit`
  sends the run back. The controller settles the state (a closed human
  answer wins, then deny, then ask_human holds) and writes it fenced by
  the lease, so a dead worker's late report fails and another worker
  picks the row up after the lease.
- Every claim is a heartbeat (id, host, last claim, admitted count),
  kept in memory; `jobshop ingest workers` lists them with inbox counts.
- Attention kind `ingest`: `ingest-worker` when external and no
  heartbeat for `ingest.worker_missing` (5m, timed from controller start
  when none yet), `ingest-backlog` when over `ingest.backlog_max` (0)
  pending rows are older than `ingest.backlog_after` (10m). The same
  conditions are stall signals `ingest_worker` and `ingest_backlog`.
- Metrics: `jobshop_inbox_items{state}` and `jobshop_ingest_workers`.
- NixOS: `services.jobshop.ingest.worker.enable` adds
  `jobshop-ingest-work.service`.
