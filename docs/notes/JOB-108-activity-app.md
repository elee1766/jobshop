# Activity feed for GitHub App identities (JOB-108)

2026-11-06. `GET /repos/{r}/events` answers 200 with `[]` for an
installation token, so a repo covered by a `github-app` identity showed
only `ci` items (#255).

- A repo covered by a `github-app` identity, or polled with a `ghs_`
  token, reads the per-resource endpoints instead of events: `pulls`
  (opened, merged, closed), `pulls/{n}/reviews` for the 5 newest PRs,
  `issues` (PRs skipped), `issues/comments`, `pulls/comments` and
  `commits` on the default branch (one `push` item per commit). Each has
  its own ETag in `activity.json` (`repos.<r>.etags`), so an unchanged
  repo is 304s. No `X-Poll-Interval` applies: the poll is `activity.interval`.
- A refused endpoint (the App lacks that permission) is logged once per
  source by the existing error dedupe, and the others still feed.
- Limits against events: a PR's merger is not in the list, so its author
  is the actor; a push is seen only on the default branch; reviews older
  than the 5 newest PRs' are missed.
- Self login: `identity: self login` warned every poll because a
  `github-token` holding an installation token gets 403 from `GET /user`.
  `Login` now refuses that token with a message naming the fix, and the
  new identity field `slug` (`github-token`, or `github-app` to skip
  `GET /app`) gives the login `<slug>[bot]` with no call. The warning is
  logged once per identity until the error changes.
- Tests: `TestLiveActivityApp`, `TestGitHubTokenSlugLogin`.
