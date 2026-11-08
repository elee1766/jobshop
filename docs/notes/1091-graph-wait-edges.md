# Wait edges in the trace graph (#1091)

2026-11-08

`ticket.graph` returns a relation of kind `wait` from a parked ticket to what
each of its wait conditions names:

- `run`: the run ticket (it becomes a node and is explored like any other).
- `linear.state`, `linear.label`, `github.pr`, `github.checks`,
  `github.review`, `github.label`, `github.tag`, `deploy`, `thread.reply`: an
  entry in the new `external` list (`id`, `type`, `label`, `url`, `at`). The
  URL is the canonical page. A Linear URL needs a workspace, which comes from
  the project's identities when one workspace covers the key's team, else it
  is empty and Desk shows the label only.
- `time` has nothing to point at and gets no edge.

The relation carries `type`, `reason`, `since` and `state`. A live wait is
`waiting`, or `met` for a condition already saved as met. A wait that is over
is rebuilt from the ticket's history (the `fields` diffs that set and removed
`wait`) and the `Wait (...) ended: ...` comment: `invalidated`, `cleared`,
`met`, or `ended` when the comment does not say (mode any with several
conditions and none saved as met). Up to 10 finished waits per ticket are
shown. A later wait on the same target replaces the earlier edge.

Desk draws the edge dotted, colored by state with the state as its label, and
the external node as a dashed card placed on the time axis by the earliest
wait on it.
