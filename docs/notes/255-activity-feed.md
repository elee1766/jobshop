# Activity feed (#255)

2026-11-04. Recent GitHub and Linear activity for every repo and team the
host works, in Desk (`?view=activity`), the run drawer and
`jobshop activity`.

- Config: `activity: {interval, window, max_items, repos}`. Absent: off.
  Defaults 1m, 168h, 1000 per source. Repos are every workflow `repo`
  (one per packs/github instance) plus `activity.repos`; Linear teams
  are the `teams` of every Linear identity (`*` is skipped). Nothing
  else to maintain.
- GitHub: `GET /repos/{r}/events` (pushes with commit messages, PRs
  opened/closed/merged, issues, issue and PR comments, reviews, review
  comments, branch and tag create/delete, releases) and completed
  `actions/runs` on the default branch. Both with `If-None-Match`; the
  next events poll waits for `X-Poll-Interval`. Comment edits and
  deletes, stars and forks are dropped. Token: the identity covering the
  repo (#200), else `GH_TOKEN` or `gh auth token` (as #187).
- Linear: the 50 most recently updated issues of each team since the
  last poll, with new comments and state and assignee history, as the
  identity covering the team. Comments by that identity's own user and
  by integrations count as bots. No Linear identity: no team, no poll,
  nothing logged.
- Storage: a cache file `<root>/activity.json` (items, ETags, default
  branches, Linear read times), not a store table. The data is external
  and refetchable, so it needs no migration or conformance tests, and
  store writes would add a history row and a change event per item. A
  restart reloads the file, so ETags still apply and nothing refills.
- Links: when listing, items whose `repo#n` or Linear key matches a
  run's, step's or task's `meta.github` / `meta.linear` carry that id (a
  step links as its run). Computed per call, so they follow meta edits.
- API: `activity.list` (admin, or pool grant `activity`) with `since`,
  `source`, `repo`, `kind`, `actor`, `bots` (`humans`/`bots`/all),
  `ticket`, `limit`; newest first, plus the list of polled sources.
  New items publish an `activity` event (seq 0) on `events.subscribe`;
  Desk refetches on it.
- Errors are logged once per source until the message changes or the
  source recovers, so a broken token does not log every minute.
- Desk: day groups, chips for humans (default)/bots/all and for each
  source, a kind select, lucide icons, filters in the URL (`who`,
  `source`, `repo`, `kind`). Bodies are escaped plain text cut to 280
  characters: the UI has no marked + DOMPurify path yet, so markdown is
  shown as source rather than adding a sanitizer for this view.
- The events API no longer includes a PR's title in PullRequestEvent
  payloads (seen 2026-11-04), so PR and merge rows read "PR #11
  merged"; the run link and the URL carry the context.

Rate limit: two 304s per repo per minute when nothing changed, which
GitHub does not count. Measured on gasserver: see
`docs/verify/activity-feed.md` (record the numbers there on the first
run).

Tests: `pkg/control/activity_internal_test.go`, `TestChanges` in
`pkg/linear`, `TestActivity` in `pkg/desk`.
