# #579 step 3: Update

2026-11-05. Fifth ticket-side group of #579.

- `Store.Update(actor, id, Update)` absorbs `AddLabel`, `RemoveLabel`,
  `AddComment` and `Reopen`. `Update` gains `Reopen`, `AddLabels`,
  `RemoveLabels` and `Comment`.
- One transaction, in this order: reopen, fields, labels, comment. A
  failing part rolls back the rest.
- Labels and the comment do not bump the revision, as before; each
  writes its own history row (`label_add`, `label_remove`, `comment`).
  An update with only those leaves the revision alone. `IfRevision`
  applies to the reopen or field write.
- Helpers in package store: `store.AddLabel`, `store.RemoveLabel`,
  `store.Reopen`, `store.AddComment` (reads the new comment back
  through `Thread`).
- `store.Store` is at 29 methods (33 before).
