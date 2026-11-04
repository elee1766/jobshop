# Ingest spool and `jobshop ingest serve` (#334)

2026-11-04. Item 2 of #318. Nothing drains the spool yet; that is the
next item.

- Spool: one file per delivery in `ingest.spool_dir` (default
  `<root>/ingest/spool`), named `<source>-<delivery_id>.json`, holding
  `{source, delivery_id, kind, received_at, headers, payload}`. Headers
  are only the signature, event and delivery headers. The delivery id
  is `X-GitHub-Delivery` or `Linear-Delivery`; with none, a hash of the
  body.
- Atomic: write a dot-prefixed temp file in the same dir, fsync, link it
  to the final name, fsync the dir. `List` skips dot files, so a crash
  mid-write leaves nothing visible. Link fails on an existing name, so
  a delivery already spooled is a no-op, even for two concurrent writes.
- Package `pkg/ingest`: `Spool.Write`, `List`, `Read`, `Remove`, and
  `Handler`, which serves `POST /github/webhook` and `/linear/webhook`.
- Checks: GitHub `X-Hub-Signature-256` with `ingest.github_secret_file`.
  Linear `Linear-Signature` with `ingest.linear_secret_file` or any
  linear-app identity's `webhook_secret_file`, plus `webhookTimestamp`
  within 60s. Filters that need only the payload (#300): GitHub `ping`,
  Linear events other than Issue and Comment, or not by a human. The
  org and self checks need a Linear token and stay with the consumer.
- Status: 202 spooled or already spooled, 204 filtered, 401 bad
  signature or stale, 400 bad body, 404 source off, 500 spool write
  error (the sender retries).
- `jobshop ingest serve` reads jobshop.yaml only: no controller, no
  database. `ingest.listen` (default `127.0.0.1:8740`), `--listen` and
  `--spool-dir` override. A source with no secret is not served.
  Secret files are re-read per request.
- Controller: `linear.webhook_listen` uses the same handler and spool.
  After a new Linear delivery is spooled it still runs the #300
  in-memory step, so behavior is unchanged except the answer is 202 (or
  204 for a filtered event) instead of 200.
