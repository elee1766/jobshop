# Host outage pauses pools (JOB-194)

2026-11-10. On 2026-11-10 the host's GitHub token provider answered 422 for
about two hours. Every step that called `gh` failed and burned retries, a
run hit its limit and filed two escalations, and steps deferred by retry
backoff needed a manual `ticket undefer`. The fault was the host's, not the
steps'.

## What it does

An identity that cannot get credentials pauses the pools that use it, like
a rate limit backoff:

- No new claims (session loop and `claim` RPC), no scale up. Open sessions
  stay and wait. Chat pools are never paused, so the operator can still
  talk to the controller.
- A step closed `failed` or released plainly while its pool is paused is
  handed back at once and counts no attempt. A comment says why. Nothing
  reaches the retry limit, so no escalation is filed.
- A probe runs on the supervisor pass every `OutageProbe` (default 1m, off
  the pass): it forgets the cached token, mints a new one and, for a
  GitHub identity, calls `GET /rate_limit` with it (Linear: reads the
  viewer). Success lifts the pause. The held steps are claimed on the same
  pass, with no undefer.
- `pool list` has `outages: {<pool>: {identity, reason, since, next_probe}}`
  and Desk's system page shows `paused: host outage (<identity>)`.
- Attention has one `outage` row per identity, however many pools and
  steps wait on it. Choices: `Probe now` and `Keep paused`.
- `pool unquarantine <pool>` also lifts an outage pause for that pool.

## What counts as an outage

`identity.Outage(err)` is true for: a failed or timed-out credential
command (`config.SecretError`), a missing or unreadable token file, 401, a
403 that is not a rate limit, any 5xx, and a network error. The signal is
the controller's own token mint (`identity.token`, `identity.check`, and
the controller's GitHub and Linear clients). A step failing for its own
reasons never shows up there, so it counts attempts and escalates as
before.

The incident's `jobshop-github-token: provider HTTP status 422` came from a
credential command, so it is a `SecretError` and counts. A 422 from an API
endpoint itself is the request's fault and does not. The pool is paused
only when the failing identity is named by the pool or its project
(`identities:`).

## Limits

- State is in memory. A restart forgets the pause; the first failing mint
  sets it again.
- An agent that gets a token and then fails against a broken API (not a
  mint failure) is not detected here. The probe covers `/rate_limit` only.
