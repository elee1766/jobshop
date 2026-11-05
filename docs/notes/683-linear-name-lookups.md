# #683: Linear name lookups under the complexity budget

2026-11-05.

- `linear create --state/--label`, `linear update`, `linear state` and
  `linear catalog` failed with "Query too complex" even after #680
  split the team read into one connection per query.
- Cause, measured on the real API: Linear scores a connection nested
  under `teams(filter: ...)` once per team the list may return.
  `teams(filter) { states(first: 100) }` scored 12550 of 10000.
  The same page under `team(id: "JOB")` (a team key works as the id)
  is one object and passes.
- Fix: every team read (`team`, `teamConn`, `Catalog`, `SetState`) uses
  `team(id: $team)`, paged at 100 with the cursor. `SetState` now uses
  the same paged reader instead of an unpaged `states` read.
- Test: `complexity_test.go` scores queries like Linear (page sizes,
  times 125 under `teams(`), checks the old shape is rejected, and runs
  create with state, two labels and a parent, and `SetState`, against
  that budget.
