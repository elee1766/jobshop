# Provider network errors back off (JOB-66, 2026-11-06)

- A session/prompt error whose message says "Failed to send request" or
  "error sending request" (the agent could not reach its provider) is
  retried on the same host with the rate-limit backoff: doubling from
  `Options.Backoff`, capped at 15x. The session keeps its claims and does
  not restart, so the crash and quarantine counters are untouched.
- At most 6 such retries per turn, then the turn fails as before.
- Log line: `provider unreachable, backing off` with the attempt number.
