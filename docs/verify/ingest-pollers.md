# Ingest pollers (#341)

On gasserver after deploying #341, with a repo the token can read
(a scratch repo, or gfx-labs/jobshop):

1. Add to the config and reload:
   `ingest: {pollers: [{name: gh-poll, source: github, repos: [<repo>], interval: 30s}]}`.
   `jobshop config check` passes. Within a minute the controller log
   has no `ingest poller` WARN, and
   `sqlite3 -readonly <root>/jobshop.db "select * from poll_cursors"`
   shows three rows for `gh-poll` (issues, reviews, comments).
2. Comment on an issue in the repo. Within one interval the inbox has a
   row: `sqlite3 -readonly <root>/jobshop.db "select endpoint, kind,
   number, natural_key from inbox order by id desc limit 3"` shows
   `gh-poll  issue_comment  <n>  github:comment:...`.
3. Wait two more intervals with no activity: no new inbox rows, and
   `gh api rate_limit` `core.used` does not grow by more than the
   per-PR review reads (issues and comments lists answer 304).
4. Restart the controller. No row is written twice and no old activity
   is read again (count of `gh-poll` rows unchanged).
5. On a repo that also has a webhook endpoint: one new comment gives
   one inbox row, whichever path arrived first.
