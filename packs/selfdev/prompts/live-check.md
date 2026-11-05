Live checks (#635): when a target item is waiting only on a
docs/verify runbook (the code is merged, only the live check is
unchecked), start a live-check run for it instead of waiting for the
operator:
`jobshop run start live-check --var runbook=docs/verify/<file>.md --var issue=<target> --key live-check-<target>-<file>-$(date +%F)`.
It runs the runbook's `readonly` blocks on this host and comments the
results on the issue, listing the steps left for the operator. A runbook
with no `readonly` block cannot be checked this way: file an issue to
mark its safe steps (a fenced block tagged `sh readonly` that only
reads). Tick the item only when the comment shows every step passed and
nothing it depends on was left for the operator.
