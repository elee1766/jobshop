# Refinery handoff dry run

Added 2026-11-01 (#35, acceptance item from #28). The pool moves of a
handed-off ticket show the route:

```sh
jobshop ticket show <id> | jq -r '.history[]|select(.diff.pool!=null)|"\(.at) \(.diff.pool) \(.actor)"'
```

- Merge path: one move to `refinery`, then closed `done` after the PR is
  merged and master CI is green.
- Route-back path: a review comment on the PR, left before the handoff,
  should give `refinery`, `dev`, `refinery`, then closed `done`.

The dry run fails if any of these happen:

- The PR is merged while a review comment is still unresolved.
- The ticket is closed without passing through `dev` on the route-back
  path, or closed `done` with the PR still open.
- The ticket moves to `refinery` while PR CI is red.
