# #372: TestRateLimitBackoff counts backoffs from the log

2026-11-04.

TestRateLimitBackoff used to poll `session.list` every 20ms and require
seeing `BackoffUntil` set at least once. The backoffs last 100, 200 and
400ms, so under `-race -count=5` load the poll sometimes missed all three.
Because it then never recorded the host pid, the same-host check failed too.

The test now passes a `control.Options.Log` whose handler counts the
"provider rate limited, backing off" records, and asserts exactly three,
one per 429. The host pid is taken from any `session.list` sample, not
only from one taken during backoff. The claim-kept check still runs
whenever a sample does land inside a backoff. No controller code changed.
