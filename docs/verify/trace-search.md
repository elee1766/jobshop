# Trace search

2026-11-07

The trace search form searches ticket IDs, titles, bodies, fields and metadata.
Space-separated terms must all match. Matching ignores case, and `%`, `_` and
`!` are literal characters. `#` separates a repository from its PR number so
qualified references also match structured GitHub metadata. Queries are limited
to 512 bytes and 16 terms.

Search includes closed tickets unless a status is selected. Results are newest
first, 50 per page. Selecting a result opens its graph. The separate `open`
action resolves an exact ticket, session or external reference.

A qualified PR reference can resolve through either GitHub metadata or an
explicit reference in a ticket's title/body. Matching workflow steps resolve to
their parent run. Multiple matching roots remain ambiguous, so use search to
choose a particular run.

## Verify

Use a disposable controller, never a production controller, for mutation tests.

```sh
cd ui
DESK_MUTATE=1 DESK_URL=http://127.0.0.1:7480 npm run test:e2e -- search.spec.js
```

The browser test checks completed tickets, pagination, title and PR searches,
status filters, result navigation, and mobile toolbar sizing. Repeat with
`DESK_PREFIX=1` against Desk configured with a trusted local proxy to check
subpath navigation. Store conformance tests exercise the same query contract
for SQLite and PostgreSQL. Set `JOBSHOP_TEST_POSTGRES` to an isolated PostgreSQL
server on which the test runner may create and drop databases.
