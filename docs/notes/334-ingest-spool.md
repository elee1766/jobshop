# Ingest spool and `jobshop ingest serve` (#334, #339)

2026-11-04. Item 2 of #318. #334 built a file spool with one GitHub and
one Linear secret; #339 replaced both, the same day, before anything used
them. Nothing drains the spool yet (#335).

- Endpoints: `ingest.endpoints` is a list of `{name, source, secret_file
  | identity, path, repos}`. `name` is `[a-z0-9-]+` and unique. `path`
  defaults to `/<source>/<name>` and must be unique. `secret_file` is
  relative to the root; `identity` names a linear-app identity and uses
  its `webhook_secret_file`. A request is checked only against its
  path's endpoint(s).
- Implicit Linear endpoints: a linear-app identity with
  `webhook_secret_file` that no endpoint names gets `linear-<identity>`
  at `/linear/webhook`. These share the path; the secret that signed
  picks the endpoint and identity, as in #300.
- `repos` (GitHub only): `repository.full_name` must be listed
  (case-insensitive), else 403 and nothing is spooled. With `repos` set,
  deliveries without a repository (org, App) get 403 too; with it empty
  everything passes.
- Removed: `ingest.github_secret_file`, `ingest.linear_secret_file` and
  `ingest.spool_dir` fail the load with an error naming the replacement.
- Spool: a SQLite database `ingest.spool` (default
  `<root>/ingest/spool.db`), separate from jobshop.db so only the
  controller writes the main database and the receiver works while the
  controller is down. Table `spool`: id, endpoint, source, delivery_id,
  kind, received_at, headers (signature, event and delivery headers
  only, JSON), payload; `UNIQUE(source, delivery_id)`, written with
  `INSERT ... ON CONFLICT DO NOTHING`. WAL, `busy_timeout`,
  `synchronous=FULL`: a 202 means the row is durable. Schema is an
  embedded `CREATE TABLE IF NOT EXISTS` and `PRAGMA user_version`, no
  goose for one table; a newer version than the binary refuses to open.
  `pkg/ingest.Spool`: `Write`, `List(limit)`, `Delete(ids)`.
- Delivery id: `X-GitHub-Delivery` or `Linear-Delivery`; with none, a
  hash of the body, so a resend still dedupes.
- Checks: GitHub `X-Hub-Signature-256`; Linear `Linear-Signature` and
  `webhookTimestamp` within 60s. Payload-only filters (#300): GitHub
  `ping`, Linear events other than Issue and Comment or not by a human.
  The org and self checks need a Linear token and stay with the
  consumer.
- Status: 202 spooled or already spooled, 204 filtered, 401 bad
  signature or stale, 403 repo not allowed, 400 bad body, 404 no such
  endpoint, 500 spool write error (the sender retries).
- `jobshop ingest serve`: reads jobshop.yaml once; adding an endpoint
  needs a restart. Secret files are re-read per request, so rotating
  one does not. Listens on `ingest.listen` (default `127.0.0.1:8740`).
- Controller: `linear.webhook_listen` serves the Linear endpoints with
  the same handler and writes the same spool database, then runs the
  #300 in-memory step for a new delivery until the drain lands. The
  inbox gained an `endpoint` column (migration 00009) for admission
  hooks to match on.
