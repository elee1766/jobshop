# Recall retries on SQLITE_SCHEMA (#1069, 2026-11-08)

`TestRecall` failed once in CI with `database schema has changed: vtable
constructor failed: recall (17)`. Cause: recall reads the FTS5 `recall`
table on the read pool, while `Collection()` registers other collections'
FTS5 tables on the writer (`CREATE VIRTUAL TABLE IF NOT EXISTS`). A read
that reloads the `recall` virtual table while the schema changes fails
with SQLITE_SCHEMA instead of being re-prepared. Reproduced with
`TestRecallDuringCollectionDDL` (fails every run before the fix).

Fix: `Store.recall` repeats the query on SQLITE_SCHEMA, up to 16 tries
with the jittered `busyWait` between them. Five immediate retries were
not enough: the error lasts while the writer applies the DDL. Other
errors, and a cancelled context, return at once.

Not changed: collection `Search` reads FTS5 tables on the writer
connection. Only the read pool was seen failing, so it is left alone.
