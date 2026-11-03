# Subscriptions live check (#187)

1. Set `subscriptions: {interval: 1m}` in the host config, deploy or
   `jobshop reload`.
2. Pick a pr-work step in progress. After a minute,
   `jobshop ticket show <step>` has `meta.subscribe.threads` with the
   issue and (after the PR comment) the PR.
3. As a human, comment on that PR. Within two intervals the step has a
   comment by `github:<you>` and the session's journal shows the message
   turn. A bot comment (CI) in the same window produces nothing.
4. On a closed step whose run is still open (e.g. waiting in
   review-merge), comment on the PR again: the step reopens assigned to
   its last session if open, else to the pool.
5. Restart the controller; no comment is delivered twice.

Link the step and comment URLs on #187 and the phase 6 target (#192).
