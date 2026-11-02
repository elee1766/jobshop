# Ticket trace (#53)

`jobshop ticket trace [id] [-d depth]` prints a tree as JSON. Each node
shows the trigger that created the ticket (`meta.trigger.name`), its
creator, the sessions that claimed it (from history), its edges, and the
GitHub issue and PR URLs in its body and comments. Children are tickets
with `parent_id` set to the node, plus tickets that one of its holders
created while holding it (claim to close, release or lease expiry). The
second rule links a plan ticket to the dev tickets its planner filed. The
depth defaults to 5, and a ticket appears once. Store `Filter.CreatedBy`
backs the lookup.

`deploy/verify-closed-loop.sh <plan-ticket>` (#76) checks one cycle on that
tree: the plan ticket came from a trigger and is closed, its planner
session created a dev ticket, one of those reached `refinery`, and it links
a PR that `gh` reports as merged. It exits 1 naming the first missing link.
`--trace FILE` reads saved JSON. Tests in `deploy/` run it against
`deploy/testdata` with a fake `gh`.
