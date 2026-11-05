# #614 Desk session view attach time

2026-11-05. Opening a session in Desk took about 4s on gasserver because
`sessionStream` called `session.list --all` to find one session's last
seq, and `session.list` decoded every journal line of every session.

- `session.subscribe` takes `tail`: with `since_seq` 0 the controller
  starts at the last `tail` frames. Desk passes its `?tail=` (default 200)
  and no longer lists sessions.
- `pkg/control/journaltail.go` reads the last frame from the end of the
  journal (64 KiB block, doubled until a complete line fits) instead of
  `acphost.JournalTail`, which scans the whole file. It should move into
  acphost later.
- Not done: batching `sessionInfo` store queries. Measure `session list
  --all` after deploy first.
