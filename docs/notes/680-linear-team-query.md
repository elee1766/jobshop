# Linear write planners: one connection per query (2026-11-05)

`linear create` (and update, ask, relate) resolved names through one team
query that fetched states, labels and members at 250 each. Linear rejected
it as "Query too complex" (#680). The planners now read the team id first,
then only the lists they need (states for `--state`, labels for `--label`,
members for `--assignee` or mentions), each in its own paged query like
`Catalog` (#653).
