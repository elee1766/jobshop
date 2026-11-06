# HTTP hooks and retry

2026-11-06.

- Guardrail hooks (`hooks:`, #188) take `url` + `secret_file` instead of
  `command`, like admission url hooks (#348). The event JSON is POSTed,
  signed `X-Jobshop-Signature-256: sha256=<hmac>` with a random
  `X-Jobshop-Request-Id`; the response body is the answer, read like a
  command's stdout (empty is allow, bad JSON is deny). A url hook runs on
  the controller, not in the pool sandbox.
- Both kinds of url hook take `retry: {attempts, backoff, max_backoff}`.
  `attempts` counts the first try (0 or 1: none, at most 20). A transport
  error, a timed-out try, HTTP 429 or 5xx is retried; another status is
  final. The wait is `backoff` (default 1s) doubling per try, or the
  server's `Retry-After`, either capped at `max_backoff` (default 30s).
  `timeout` bounds each try. Out of tries, the last error decides: deny
  for a guardrail hook, the `failure_policy` for admission.
- The admission lease counts every try and the longest waits, so an
  item is not reclaimed while its hooks retry.
- `retry` and `secret_file` on a command hook are config errors.
