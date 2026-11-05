# Refinery waits on open questions (2026-11-05)

The github pack's `review-merge` step merged PR #580 (a design note gated
on operator review) and PR #392 while the operator's review question on
the issue was still open: the merge prompt only checked CI, approvals and
review comments.

The step now counts open `needs-human` message tickets whose
`meta.github.issue` is the run's issue, and checks for a `needs-decision`
label on the issue or PR. Either one is case 5 (waiting for approval),
which retries and then escalates.
