# Subscription bot allowlist (#402)

2026-11-04. Subscriptions (#187, #233) dropped every bot comment, so
review bots (a reviewer App, Kody line findings) never reached a session.

```yaml
subscriptions:
  bots: ["kody[bot]", "coderabbitai[bot]"]   # default empty: all bots dropped
```

- A listed login's comments, reviews and review comments are delivered
  like a human's: an interrupt message to the working session, a ticket
  comment, and a reopen within `resume_window`. Exact match, including
  the `[bot]` suffix on GitHub; on Linear the app actor's name.
- `ignore` still wins. `authors` (#340) applies to humans only, so a
  listed bot is delivered even when `authors` is set.
- The host's own identity is always dropped, even if listed: a
  github-token's user (GET /user), a github-app's `<slug>[bot]` (GET /app
  with the App JWT, looked up only when `bots` is set), a Linear key's
  `isMe`, and a linear-app's own app actor (viewer name).
- Empty `COMMENTED` reviews stay dropped; unlisted bots are dropped and
  still advance the read position.
- Test: `TestSubscriptionsBots`, `TestGitHubAppLogin`.
