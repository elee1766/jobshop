# Ingest dispatch in workers (#633)

2026-11-05.

- `jobshop ingest work --dispatch` also dispatches admitted items: each
  pass calls `ingest.dispatch` (admin), and the controller runs the same
  dispatch as its own pass, claiming rows as holder `worker:<id>` with
  the 5 minute dispatch lease. The worker sends no store writes itself.
- Every dispatcher (the controller's `dispatch` holder and any number of
  workers) claims rows through the fenced inbox lease, and comment
  delivery keeps the `dispatch/<row id>` write key, so a row is sent
  once even when a lease expires mid-hand-off.
- `ingest.dispatch: external` (default internal) stops the controller's
  own dispatch. `--no-admit` makes a dispatch-only worker for
  `ingest.workers: internal`.
- An `ingest.dispatch` call is a heartbeat; `ingest-worker` attention
  also fires when only dispatch is external.

Tests: `pkg/control/dispatchwork_internal_test.go` (three racing
dispatchers, a worker dying mid-lease on fake time).
