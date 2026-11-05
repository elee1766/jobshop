# #579 step 3: Apply

2026-11-05. Fourth ticket-side group of #579.

- `Store.Apply(actor, Graph)` replaces `Create`, `CreateGraph`,
  `ApplyGraph`, `AddEdge` and `RemoveEdge`. `Graph.RemoveEdges` is new.
- One transaction, in this order: check `Unclaimed`, insert `Tickets`,
  re-parent, remove `RemoveEdges` (a missing edge is `ErrNotFound`), add
  `Edges`. Removing before adding lets one call reverse a blocking edge
  without a cycle error.
- Helpers in package store, not interface methods: `store.Create`,
  `store.AddEdge`, `store.RemoveEdge`.
- `store.Store` is at 33 methods (37 before).
