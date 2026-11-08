# pkg/control split, step three: journal reads, search and recall (JOB-140, 2026-11-08)

Third subsystem out of `pkg/control`, after journal shipping (JOB-119)
and compression (JOB-89). No behavior change: RPC shapes, defaults
(`session.journal` 500/5000, `journal.search` 1000/10000), error
messages, seq cursors and compressed-file support are as before.

## Boundary

- `pkg/control/journalread` does the reading:
  - `Page(Access, params)` is `session.journal`: paging by seq, `from`
    and `kinds` filters, redaction of each frame. Reads go through
    acphost's `ReadJournal`, so plain and `.zst` journals page alike.
  - `Run(ctx, Search)` is `journal.search`: limits, `since`/`start`,
    truncation (asks VictoriaLogs for limit+1), 400 as `ErrInvalid`.
  - `Summary(dir)` is what recall indexes for a session, and
    `Env(dir)` is the env a session's secrets are redacted with.
- It takes already-authorized input only: an `Access` (session dir plus a
  `Redactor`) for one session, or a `Search` whose `Project` is the scope
  the controller picked. It imports no `pkg/control`, `scope` or config.

## Stays in control, and why

- `journal.go`: finding the ticket, `canReadJournal` (admin, own session,
  `session-journal` grant plus same project) and the `forbidden` errors.
  The decision needs `scope` and the live config.
- `journalsearch.go`: `requireGrant` (`journal-search`), the project
  scope (`Project` only for non-admin), the VictoriaLogs client from
  config, and `journal.backfill` (admin check, shipper call).
- `recall.go`, all orchestration: `recall` resolves the project scope
  (`c.project`), queries comments, journals (`journal.Recall`) and
  learnings (`learning.Recall`), and `mergeRecall` interleaves the three
  sources. It needs `scope`, the store and the learning package, none of
  which belong in a journal reader. `indexJournal` stays on `runner`
  (called after each turn with the runner's dir and project) and only
  asks `journalread.Summary` for the text.
- `session.subscribe` (`chat.go`) streams live frames with fsnotify and
  stays: it is not a paged read. It uses `journalread.Env` for redaction.

## Tests

`journalread` unit tests cover paging, filters, cursors, redaction,
compressed history, summaries and search against an httptest server.
Controller tests (`journal_internal_test.go`,
`journalcompress_internal_test.go`, `journalread_internal_test.go`) cover
forbidden access, redaction, compressed history through `session.journal`
and `session.subscribe`, search grant and project scope, and recall
scope over indexed journals.
