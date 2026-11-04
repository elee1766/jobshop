# Linear read commands (#279)

2026-11-04. Read half of "CLI parity" on target #258.

- `jobshop linear context <KEY> [--all-comments]`: the issue with
  labels, assignee, parent, children, relations (`related`, `duplicate`,
  `blocks`, and `blocked_by` from inverse relations), `blockers` and the
  10 newest comments (250 with `--all-comments`), oldest first.
- `jobshop linear list --team T` and `search <q> --team T`, with
  `--state --label --assignee` (`me` is the key's own user) and paging:
  `--limit` (default 50, max 250) and `--after <end_cursor>`. Output is a
  page `{issues, has_next, end_cursor}`.
- `jobshop linear catalog projects|labels|states|users --team T`.
- `jobshop linear eligible <KEY> [--skip-labels a,b]`: a read-only
  intake gate. Not eligible when archived, done (`completed`), canceled,
  blocked by an open issue, an open run (running or needs_human) has
  `meta.linear.issue` = KEY, or a label matches a skip label
  (case-insensitive). Prints `{issue, eligible, reason}`; exits 0 when
  eligible, 1 when not. packs/github has `linear_skip_labels` (default
  empty) and the Linear planner prompt runs the gate before `run start`.
- All of them get the team token through `identity.token` as the other
  linear commands do, so team scoping is unchanged. The list commands
  need `--team` for that reason.
- Tests: `pkg/linear/read_test.go` on recorded responses in
  `pkg/linear/testdata` (context, two list pages, search, catalog) and
  every eligible rejection reason.
