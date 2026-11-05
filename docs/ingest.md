# Webhooks and ingest

Moved from the README (2026-11-05, #551).

## `jobshop ingest serve`

`jobshop ingest serve` receives GitHub and Linear webhooks, checks the
signature, and writes each delivery to a spool database
(`ingest.spool`, default `<root>/ingest/spool.db`). It needs no
controller and does not touch jobshop.db. List one endpoint per hook,
each with its own secret; GitHub endpoints can limit which repos they
accept:

```yaml
ingest:
  endpoints:
    - {name: app, source: github, secret_file: secrets/gh-app, repos: [acme/app]}
    - {name: eng, source: linear, secret_file: secrets/linear-eng}
```

Each is served at `/<source>/<name>` unless `path` is set. A linear-app
identity with `webhook_secret_file` is also served at `/linear/webhook`.
It listens on `ingest.listen` (default `127.0.0.1:8740`), so put a
reverse proxy in front that terminates TLS and forwards only the
endpoint paths:

```
location /github/ { proxy_pass http://127.0.0.1:8740; }
location /linear/ { proxy_pass http://127.0.0.1:8740; }
```

It answers 202 when spooled, 204 for an event jobshop ignores, 401 for a
bad signature, 403 for a repo not on the endpoint's list, and 500 when
the spool write fails so the sender retries. Adding an endpoint needs a
restart; rotating a secret file does not. See
[notes/334-ingest-spool.md](notes/334-ingest-spool.md).

## Generic endpoints

Any other sender (a queue puller, CI, monitoring) uses a `generic`
endpoint: HMAC-SHA256 of the raw body in a header you name, and `map`
paths that fill the inbox fields. Trusted local senders skip HTTP with
`jobshop ingest put --endpoint hints [--delivery-id ID] < payload.json`,
which exits 0 for new or duplicate deliveries. See
[notes/399-generic-ingest.md](notes/399-generic-ingest.md).

```yaml
    - name: hints
      source: generic
      secret_file: secrets/hints
      signature_header: X-Hint-Signature
      signature_prefix: "sha256="
      delivery_id_header: X-Hint-Delivery
      map: {kind: event, repo_or_team: repo, number: prs.0, natural_key: "{repo}:{head}"}
```
