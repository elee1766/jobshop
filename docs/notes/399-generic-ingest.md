# Generic ingest endpoints and `ingest put` (#399)

2026-11-04. Behavior as built.

- `source: generic` endpoints take any HMAC-signed JSON sender.
  `signature_header` (required) holds the hex HMAC-SHA256 of the raw
  body, after `signature_prefix` is cut; `delivery_id_header` names the
  sender's id. Same secret_file, path (`/generic/<name>`), 1 MiB cap and
  status codes as GitHub and Linear (202, 401, 400, 500); there is no
  204 filter. Without an id header or `map.delivery_id`, the body hash
  dedupes.
- `map` gives dotted JSON paths (`repo`, `prs.0`, `sender.login`) for
  `kind`, `repo_or_team`, `number`, `actor`, `comment_id` and
  `delivery_id`; `natural_key` is text with `{path}` parts, stored as
  `generic:<endpoint>:<text>`. Missing paths leave fields empty; a
  `number` that is not an integer sends the row to `spool_bad`.
- The map is stored with each spooled delivery (header `Jobshop-Map`), so
  the drain needs no endpoint config and a config change does not remap
  deliveries already spooled.
- `jobshop ingest put --endpoint NAME [--delivery-id ID] [--kind K] <
  payload.json` writes one delivery to the spool with no HTTP or
  signature. It prints `spooled` or `duplicate` and exits 0 for both;
  any write failure exits nonzero, so a queue puller acks only after 0.
- Schema: spool v3 and store migration 00011 rebuild their tables to
  allow the new source (SQLite cannot drop a CHECK). The inbox now takes
  any non-empty source. Admission `match.sources` accepts `generic`;
  `repos`/`teams` globs do not constrain generic items.
