# GitHub author allowlist (#340)

2026-11-04. The repo is public, so issue text and comments from anyone
would otherwise reach prompts.

- Pack var `authors` (github pack, comma-separated). `issue_no_run`
  starts stall-check only for issues filed by an App (the planner),
  `human`, or a login in `authors`, or carrying the `accepted` label.
  Logins compare case-insensitively.
- `accepted` lets a maintainer admit an outside report. gh does not say
  who added a label; on this repo only collaborators can add labels.
- stall-check validates a person-filed issue before starting pr-work:
  it reproduces or confirms it, comments, then starts the run or closes
  the issue with the reason.
- `subscriptions.authors`: when set, only those logins' comments are
  delivered. Others advance `seen` like `ignore`.
- Ingest admission (#318) should apply the same list.
