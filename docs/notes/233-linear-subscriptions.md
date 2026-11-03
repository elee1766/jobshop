# Subscriptions follow Linear comments (#233)

2026-11-03. Follow-up of #187 and #201.

- A task or step with `meta.linear.issue` (set by `tracker: linear`
  runs) follows that issue like its GitHub thread: entry
  `{linear: "ENG-5", seen}` in `meta.subscribe.threads`, starting at the
  current time.
- The controller reads it with `pkg/linear` GetIssue as the identity
  covering the key's team (any configured identity, not just the
  ticket's project). No identity: the fetch fails, logged, retried next
  pass, read position kept.
- Comments by the identity's own user (`isMe`, i.e. sessions posting
  with `jobshop linear comment`) and comments with no user (integrations)
  advance `seen` but are not delivered, like GitHub bots. `ignore`
  matches Linear display names too.
- Delivery rules are unchanged; the ticket comment author is
  `linear:<name>`.
- GetIssue reads the latest 100 comments; a thread with more than 100
  new comments in one interval loses the oldest.

Tests: `TestSubscriptionsLinear`, `pkg/linear`. Live check:
`docs/verify/linear-subscriptions.md`.
