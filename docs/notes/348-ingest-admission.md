# #348 Ingest admission

2026-11-04. Item 4 of #318. A controller stage (`pumpAdmission`,
`pkg/control/admission.go`) lease-claims `pending` inbox items and moves
each to `admitted`, `denied` or `held`. Nothing fills the inbox until
the drain (#335) lands; tests insert rows directly.

- Config: `ingest.admission`, a list of `{name, type: mutating|validating,
  command | url, secret_file, timeout, failure_policy: fail|ignore,
  match: {sources, kinds, repos, teams, actors, bots, projects}}`.
  `repos` and `teams` are `path.Match` globs; repos constrain GitHub
  items only and teams Linear items only. `bots: true` matches actors
  ending in `[bot]`, `false` only humans. Bad values fail the load.
- Event: `{on: "admission", delivery_id, source, kind, repo_or_team,
  number, comment_id, actor, project, received_at, payload}`.
  `delivery_id` is stable across re-claims so hooks can dedupe.
  `project` is the default project for now.
- Order: every matching mutating hook in config order, then every
  matching validating hook on the final event. A mutating answer is
  `{"event": {...}}` or `{"patch": [...]}` (RFC 6902, applied to the
  whole event), plus `reason`; empty output is no change. A hook may
  not change `on`, `delivery_id` or `source`. The final `payload` is
  stored as `admitted`; the raw payload is kept.
- Validating answers use the #188 format. Any deny: `denied`. Otherwise
  any ask_human: `held`, plus one needs-human ticket keyed
  `admission/<delivery_id>`. Once that ticket is closed, the next
  admission of the item takes the human's answer instead of the hooks.
  Held items are not re-claimed yet (#353).
- `hook_results` on the row: per hook `name, type, decision, reason,
  diff, duration_ms, error`. Mutating decisions are mutated, unchanged,
  failed or ignored.
- Failure: `fail` (default) makes a validating error, timeout, non-2xx
  or bad JSON a deny, and a mutating one leaves the event unchanged
  with result `failed`. `ignore` records `ignored` and moves on.
- Transports: `command` is `sh -c` in the root with the event on stdin.
  `url` is a POST with `X-Jobshop-Request-Id`, `X-Jobshop-Delivery` and
  `X-Jobshop-Signature-256: sha256=<hex HMAC-SHA256 of the body>` keyed
  by `secret_file` (trimmed), bounded by `timeout` (default 30s).
- `on: intake` hooks also run as validating hooks matched to
  `sources: [linear]`; their event has `on: "intake"`.
- Gap: dispatch does not read admitted rows yet (#318 item 5), so
  GitHub comments still reach subscribers unchecked and the Linear-only
  early return in `intakeComments` stays until dispatch moves over.
- JSON patch is a small local implementation to avoid a dependency.
