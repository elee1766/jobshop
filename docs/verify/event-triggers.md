# Event triggers fire on webhooks (JOB-62)

1. Confirm the repo's GitHub webhook sends `workflow_run` and
   `check_suite` events to ingest (repo settings, Webhooks, recent
   deliveries show 200 for both kinds).
2. Push a commit that fails CI on master (or use the ci-repair drill,
   `docs/verify/ci-repair-drill.md`). Within a minute of the CI run
   finishing, the controller log has `trigger fired trigger=master-red
   event=workflow_run`, and one ci-repair run exists for the SHA.
3. Wait past the 5m fallback: no second ci-repair run for the same SHA.
4. Stop the ingest listener, fail CI again: the run still starts, within
   the 5m fallback.
