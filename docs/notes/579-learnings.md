# Learnings on a collection (#579)

2026-11-05. Third domain PR of #579 step 2 (design:
`docs/notes/579-collections.md`).

- `pkg/learning` keeps learnings (#264) in the `learnings` collection.
  Key is the id as text (ids stay integers: the CLI, the RPC, memory
  edits and `memory:<id>` refs use them); state is the status; the
  value holds the old columns plus the ticket's project. Indexed
  fields: kind, scope, ticket, pr. A new id is one past the highest of
  the 20 newest records.
- Search: summary and fix are the collection's search fields
  (`fts_learnings` in SQLite, a GIN index on the partition in
  Postgres). Learning rows and their triggers leave the shared `recall`
  index. The `recall` RPC searches comments and journals there and
  learnings in the collection, and interleaves the two lists, each in
  its own rank order, because the scores do not compare (as the design
  note planned for all sources). A learning with no project still
  matches every project.
- `CountLearnings`, `LearningsBetween` and `LearningsByPR` read with
  `List` and filter in Go (globs; hundreds of rows), as the design
  note's "queries that need more than an index" rule 1.
- `Store` loses `AddLearning`, `Learnings`, `UpdateLearning`,
  `GetLearning`, `CountLearnings`, `LearningsBetween`, `LearningsByPR`
  and `SetLearningGate`. Migration sqlite 00022 / postgres 00007 copies
  rows with their ids, builds the search index and drops the table;
  Down copies them back and restores the recall rows and triggers.
  `storemigrate` no longer copies a learnings table (records carry them)
  and skips SQLite `fts_*` tables, which Postgres indexes itself.
- Not in this PR: session journals stay in `recall` (the design note
  moved them with learnings); they move to a `journals` collection in a
  later PR, with the comment index.
