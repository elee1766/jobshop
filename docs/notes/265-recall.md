# Recall (#265)

2026-11-04.

`jobshop recall <words> [-n N] [--scope comment|journal|learning] [-t ticket]`
calls the `recall` RPC. Sessions search their own project; admin searches
all projects unless `--project` is set. Output is JSON hits, best first:
`kind`, `ref`, `ticket_id`, `project`, `at`, `snippet` (matches in
`[...]`) and `score` (bm25, lower is better).

- Index: FTS5 table `recall` (migration 00004, porter tokenizer).
  modernc sqlite has FTS5 built in, so `CGO_ENABLED=0` builds keep it.
- Comments: insert/update/delete triggers on `comments`; the migration
  backfills existing comments. A new comment is searchable at once.
- Journals: after every turn the controller replaces the session's
  document (ref = session id) with `acphost.Summarize(journal, 64 KiB)`:
  prompts, agent text, tool titles and the plan. Journals of sessions that
  ended before this change are not indexed.
- Learnings (#264): insert/update/delete triggers on `learnings` index
  summary and fix (ref = learning id); the migration backfills them.
  Project comes from the learning's ticket; learnings with no ticket are
  shown to every project.
- Query words are quoted, so FTS syntax in an error message is matched as
  text. All words must match; if nothing does, any word may.
- packs/github adds a system prompt rule `recall`: run `jobshop recall` on
  a failure and before asking a human.
