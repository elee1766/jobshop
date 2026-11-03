# Subscriptions to GitHub threads (#187)

2026-11-03.

- Config: `subscriptions: {interval, resume_window, ignore}`. Absent:
  off. Defaults 1m and 168h. The controller polls with `gh api`, so it
  uses the host's gh token and reads any repo that token can.
- A task or step follows threads in `meta.subscribe.threads`
  (`[{repo, number, pr, seen}]`). `seen` is the read position (newest
  comment time handled), so restarts do not redeliver. Followed
  without asking: its `meta.github` issue or pr (#183) and PR URLs in
  its comments (the pr-work `PR <url> head <sha>` comment). A session
  adds others with `jobshop subscribe owner/name#n [--pr]` or a URL.
  New threads start at the current time: older comments are not news.
- Read: issue comments; for PRs also reviews and review comments. Bots
  (`[bot]` logins, user type Bot) and `ignore` logins are dropped but
  still advance `seen`. Empty `COMMENTED` reviews are dropped (their
  line comments come on their own).
- Delivery, one follower per thread: in-progress ticket first, then
  open, then the most recently closed. The text is always added as a
  ticket comment by `github:<login>`. In progress: also an interrupt
  message to the holding session. Open: the comment is in the prompt on
  claim. Closed within `resume_window`: the ticket is reopened and
  assigned to the session that last claimed it if that is still open,
  so the same conversation continues; a step of a closed run is not
  reopened (operator uses `run resume`).
- Not an issue sync: no tickets are created from GitHub.
- Polling, not webhooks: no public endpoint on gasserver. Revisit if
  the API budget gets tight.

Tests: `pkg/control/subscribe_test.go`. Live check:
`docs/verify/subscriptions.md`.
