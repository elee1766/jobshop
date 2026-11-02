# Rate limits, drain, quarantine (#10, #12, #13)

- A prompt error mentioning `429`, `rate_limit` or `is unavailable` is not
  a crash: the session keeps host and claims and retries the same prompt
  after 1m, doubling to 15m. While any session of a pool backs off, no new
  sessions start for it. `session list` shows `backoff_until`.
- `jobshop drain [--pool P]` / `undrain`: no new sessions or claims for the
  pool; running turns finish and sessions end `drained`. Controller
  memory only, so a restart undrains.
- A `crashed` session releases its tickets with reason `session crashed`.
  After 3 such releases a ticket is deferred 1h and labeled `quarantined`
  with a comment listing the exits. Undo: `ticket update <id> --defer 0`
  and `label rm <id> quarantined`. A pool with 3 crashed sessions in 10
  minutes starts none for 15 minutes; the pause is set before the release.
