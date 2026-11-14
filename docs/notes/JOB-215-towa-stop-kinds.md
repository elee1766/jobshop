# towa stop kinds in provider backoff (JOB-215, 2026-11-14)

towa reports a provider failure as a typed stop in the ACP error's data,
`{"towa": {"stop": {"kind": ...}}}`, not in the message text. The
backoff checks in `pkg/execution/session/errors.go` now read it.

Recognised, by `data.towa.stop.kind` (data decoded as an object, then
`towa`, `stop`, `kind` as a string; the code is not part of the check,
towa sends these as `-32603`):

- `network`: `ProviderUnreachable`. Backs off on the same host, at most
  `providerNetRetries` (6) per turn, as for jcode (JOB-66). towa's text is
  "send request", which the jcode text match ("failed to send request",
  "error sending request") missed.
- `rate_limit` and `overload`: `RateLimited`. Backs off without a bound.
  `overload` is towa's word for a provider that is overloaded or
  unavailable, what `is unavailable` covers for jcode. towa sends the stop
  after its own retries and fallbacks.

The jcode text matches are unchanged.

Not recognised: `auth`, `context_full`, `limit`, `request` and `denied`.
They are not transient and what a session should do with each is a
separate decision (the tests assert they match neither check). The stop's
`retryAfterSeconds` and `resetAt` are not used as the wait: the existing
backoff applies.

Not captured from a real provider: `rate_limit` and `overload`. Only
`network` was seen (dead proxy, 2026-11-13). The unit and controller
tests build those two from the documented shape.

## Why a pool keeps starting sessions

Before this change a towa session with an unreachable provider ended
`crashed` after 3 restarts in 10 minutes, and the pool started a new
session that claimed the same ticket. That is by design, not a gap: each
crashed session counts toward the pool's crash log, and the third within
10 minutes pauses the pool for 15 minutes (`poolCrashes`,
`poolCrashWindow`, `poolPause` in `quarantine.go`). The ticket is
quarantined after 3 crashed sessions too. The observation stopped after
two sessions, short of both limits. Not changed here.

Tests: `TestPromptErrors` (the 2026-11-13 payload, `network`,
`rate_limit`, `overload`, `request` and others as negatives, malformed
data), `TestRateLimitBackoff/{jcode,towa}`,
`TestProviderUnreachableBackoff/{jcode,towa}`. The test agent's
`ratelimit` and `netfail` files take `N towa` for the towa shape. Check on
a real controller: [towa-provider-unreachable](../verify/towa-provider-unreachable.md).
