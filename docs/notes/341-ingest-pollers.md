# Ingest pollers for GitHub repos without a webhook (#341)

2026-11-04. Behavior as built.

- Config `ingest.pollers: [{name, source: github, repos, events,
  interval, identity}]`. Events are `issues`, `issue_comment`,
  `pull_request`, `pull_request_review` (empty: all). Interval default
  1m, at least 10s. `identity` names a github-app or github-token
  identity that covers every repo; empty uses the identity covering each
  repo, else the host's gh token. Poller names share the endpoint name
  space.
- Streams per repo, each with a cursor in the store table
  `poll_cursors` (migration 00016: poller, repo, stream, since, etag):
  `issues` (GET /repos/R/issues, issues and PRs), `comments` (GET
  /repos/R/issues/comments) and `reviews` (GET /pulls/N/reviews for each
  PR the issues stream saw updated). Lists ask for `since` the cursor,
  oldest first, with If-None-Match, so an idle repo is two 304s.
- A repo with no cursor starts at now: the first pass stores cursors and
  reads nothing. Cursors move only after the rows commit, so a failed
  pass re-reads the same window and the inbox dedupes it.
- Rows look like drained webhook rows: source github, the webhook kind,
  a payload subset (`action`, `repository.full_name`, `sender`, and the
  issue, pull_request, comment or review), and endpoint = poller name,
  so admission `match` sees one format. Polling sees the latest state,
  so issue and PR actions are `opened` or `updated`, comments `created`
  or `edited`, reviews `submitted`.
- Dedupe: comments and reviews get the drain's natural key and the
  delivery id `poll:<natural key>`, so a webhook and a poll of the same
  comment are one row (the first to arrive is kept). Issues and PRs use
  `poll:github:<kind>:<repo>:<n>:<updated_at>`; they have no webhook
  natural key, so a repo with both a webhook and a poller gets an issue
  change twice.
- Review comments (`pull_request_review_comment`) are not polled.
