# Ingest workers (#443)

1. Set `ingest.workers: external`, reload, and stop any worker:
   after `ingest.worker_missing` `jobshop attention` lists `ingest-worker`.
2. Start `jobshop ingest work` (or the NixOS unit): the item clears and
   `jobshop ingest workers` shows the worker with a recent `seen`.
3. Comment on a followed issue: the inbox row goes `admitted` by
   `worker:<id>` (controller log `inbox admission ... by=worker:...`),
   then `dispatched`.
4. `curl $metrics | grep jobshop_inbox_items` shows counts by state.
