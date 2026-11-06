# Event-driven stall triggers (JOB-72, 2026-11-06)

- A stall trigger (`when.stall`) takes `on:` and `fallback:` with the
  JOB-62 rules ([notes](JOB-62-event-triggers.md)). A matching admitted
  webhook delivery runs the trigger's stall evaluation at once, under
  the tick lock. Dedup per `stall:<signal>:<subject>`, `cooldown`,
  `max_runs` and the open-run cap are the same as on a poll, so an
  event and a poll of the same subject start one run.
- An event re-evaluates every signal of the trigger. Store signals
  ignore the event; check signals see it.
- `fallback` is the poll interval: 10m by default with `on:`. Without
  `on:` a stall trigger is polled every tick as before (or every
  `fallback` if set).
- With `on:`, a signal's `check` is a template over
  `{{.event.<field>}}` (the JOB-62 fields), rendered when it runs. On a
  poll every field is empty, so the same command must work for both,
  for example `grep "^{{.event.pr}}"`. Any other var in the check is a
  load error. Without `on:` the check runs as written.
- Event values go into `sh -c` unquoted, so a value with characters
  outside `A-Za-z0-9._/@:+=-` (a branch or label with spaces or shell
  syntax) is rendered empty, as on a poll.
