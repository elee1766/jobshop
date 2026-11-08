# Linear outage backoff (JOB-118)

2026-11-08. Linear calls made by the wait pass (`linear.state`,
`linear.label`, Linear `thread.reply`), the subscription poll, the
agent-session sweep and `linear_issue_no_run` intake go through
`linearCall` (`pkg/control/linearhealth.go`):

- Each request is bounded by `linearRequestTimeout` (15s) through the
  shared client's request timeout, and the token fetch too.
- A timeout, failed connection, 5xx or 429 (`linear.Unavailable`) marks
  the identity down. Calls as it then fail at once until a backoff passes
  (30s, doubling to 5m), so one hung API costs one request timeout per
  backoff, and the rest of the wait pass (GitHub, run, time waits) goes on.
- The outage logs one WARN `linear unavailable` when it starts and one
  INFO `linear recovered` (with `down_for`, `failed_calls`) when Linear
  answers again. Callers skip their own per-wait or per-thread WARN for
  it. A GraphQL error (not found, auth) is an answer, not an outage.
- Intake's per-issue blocker read stops the whole read on an outage
  instead of timing out once per issue.

Other Linear callers (writes, hook built-ins, a session's `jobshop
linear ...` request) are not backed off: they see the error directly.
