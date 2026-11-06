# Event triggers with poll fallback (JOB-62, 2026-11-06)

- A trigger may list `on:` entries, each `github:` or `linear:` with
  fields `event`, `action`, `repo`, `branch`, `workflow`, `conclusion`,
  `label` (empty matches any, case-insensitive). `workflow` matches the
  workflow name or its file name without extension.
- Matching runs at ingest dispatch on webhook rows (not poller rows),
  in the background under the tick lock. The trigger's `when` conditions
  and dedup run as in a poll, so an event and a poll of the same thing
  fire once: with a `check`, its output is the key; without one, the
  event's `head_sha` (else the delivery id).
- `{{.event.<field>}}` is available in `run.key`, `run.vars` and
  `command.run`. Fields: source, event, action, repo, branch, workflow,
  conclusion, head_sha, pr, label, delivery. Empty on a poll.
- `fallback:` is the poll interval. With `on:` the default is 10m, and a
  trigger with `on:` and no `when` conditions is never polled. Without
  `on:`, no fallback keeps the old every-tick poll.
- `command:` is a third action next to `ticket:` and `run:`: `run` (sh -c
  in the root dir) and `timeout` (default 1m). Each fire is a message
  ticket labeled `trigger-command` carrying the trigger meta. Success
  closes it with the output as a comment. Failure leaves it open with
  `needs-human` and the output, which holds the trigger until a human
  closes it.
- The github pack's `master-red` (workflow_run failure) and `release-pr`
  (check_suite completed) fire on events, with a 5m fallback check.
- Self-deploy on CI success is not wired in the NixOS module yet: it
  needs a `command:` trigger that starts `jobshop-self-deploy.service`,
  which is host config (#826).
