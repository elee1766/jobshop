# Session journals on a store collection (#579)

2026-11-05. Step 3, seventh group.

- Session journal summaries live on the `journals` collection
  (`pkg/journal/recall.go`): one record per session, keyed by session
  id, Search on `text`, `ticket` and `project` indexed. The controller
  rewrites it after every turn with `journal.Index`; an empty text
  deletes it.
- `Store.IndexRecall` and `store.RecallDoc` are gone. `Store.Recall` now
  searches ticket comments only (`RecallQuery` lost `Kind`). store.Store
  is at 20 methods: the 19 of the target plus comment search, which the
  last step moves to `fts_comments` when sqlc goes.
- The recall RPC queries comments, journals and learnings for the
  requested kind, each with the same limit, and interleaves them round
  robin (scores from separate indexes do not compare).
- Migration: sqlite 00024 and postgres 00009 copy `recall` rows of kind
  `journal` into the collection and delete them; Down copies them back.
  After this `recall` holds only comments.
