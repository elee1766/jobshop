# Learning kinds and trigger filters are config (#524)

2026-11-05. Learning kinds and what the curate trigger counts are set in
config, so an improvement pack per repo can record its own kinds and fire
on its own learnings.

- `learnings: {kinds: [...]}` at the top level of jobshop.yaml or a
  pack. Entries are kind names (`[a-z][a-z0-9-]*`) or `path.Match` globs
  (`review-*`, `*`). Packs append to the list like `prompt_rules`. Unset:
  `correction`, `failure`, `workaround`, as before. A pattern that can
  match no valid name is a load error.
- `jobshop learn add <kind> "<summary>"` records any allowed kind; the
  controller rejects others with the patterns in the error.
  `learn correction|failure|workaround` stay as shorthands. `correction`
  and `failure` are always allowed: the controller records them itself.
  `learn list --kind` takes a glob. `jobshop config show` prints the
  resolved `learnings.kinds`.
- `when.learnings` takes optional filters, ANDed: `kinds` (globs),
  `scope` (globs over `global`, `repo:<owner/name>`, `pool:<name>`; `*`
  also crosses `/`) and `source` (`agent` or `controller`). `min`,
  `every` and the since-last-fire window count only matching learnings.
- `CountLearnings` groups by kind, scope and source in SQL and matches
  the globs in Go, so it works on SQLite and Postgres alike (Postgres has
  no `GLOB`). A store conformance case covers it on both.
- Migrations sqlite 00017 and postgres 00002 replace the `kind IN (...)`
  CHECK with a name pattern check (sqlite rebuilds the table and its
  recall triggers, keeping ids).
- The selfdev `curate` trigger sets `scope` (var `curate_scope`) to
  `repo:elee1766/jobshop` and its pools, so learnings from another
  instance (acphost) no longer start jobshop's curator.
