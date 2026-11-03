# Outbox for side effects (#191)

2026-11-03. Design note only; nothing is built.

The question: should gateway posts and session comments be stored first
and sent later with retries (an "outbox", dsh "stored work")?

## What happens today

- Session comments (`jobshop comment`) and other ticket writes are
  already stored first. They are writes to the store through the
  controller, in one SQLite transaction, so they need no outbox.
- Gateway posts (`Gateway.post`, `pkg/gateway/gateway.go`) call the
  Matrix adapter once. If that fails, the gateway logs `gateway: post`
  at WARN and the text is lost. The session's journal still has the
  turn, so the text can be read with `session subscribe`, but nothing
  posts it again.
- The mautrix client keeps its default `DefaultHTTPRetries` of 0, so a
  429 or 5xx is not retried either.
- Escalation state (thread, session, final message) is written to the
  ticket's meta with revision checks and retried on the next pass, so
  that path already works like stored work.

## Live evidence (gasserver journal, 2026-10-27 to 2026-11-03)

- 2 `gateway: post` failures. Both were HTTP 400 `M_INVALID_PARAM:
  Invalid event_id in thread message` at 22:57 and 22:58 on 2026-11-02,
  just after an escalation thread opened. A retry gets the same 400.
  The cause is filed as #197.
- 0 transient failures (429, 5xx, timeouts) and 0 `gateway: subscribe`
  errors.

## What an outbox would cost

- A new store kind or table for pending posts: the thread, the text and
  attempt count, with a sender loop, backoff, a dead-letter state and
  metrics.
- Exactly-once is not possible with Matrix sends unless every retry uses
  the same transaction id. Without that a retry after a timeout can post
  the same text twice.
- Ordering: a later turn could post before an earlier one that is still
  being retried, unless the outbox sends per thread in order.

## Decision

No outbox. The only failures seen are permanent errors, which an outbox
would retry and then dead-letter, losing the same text. The cheap gap,
transient errors not being retried at all, is filed as #196: set
`DefaultHTTPRetries` on the mautrix client. Look at this again if the
journal shows lost posts from transient errors after #196.
