# CI waits in the GitHub pack (#555)

2026-11-05. The implement, merge, post-merge, ci-fix and ci-merge prompts
(GitHub and Linear trackers) wait on CI with a `github.checks` wait
(#381) instead of a fixed `release --until 15m/20m`, or polling inside
the turn:

- PR head: `github.checks:pr=<repo>#<n>,state=success` or
  `...state=failure`, with `--any --resume --timeout 1h`. A push moves the
  head and invalidates the wait, so the agent sets it again after each push.
- Master: `github.checks:repo=<repo>,sha=<merge sha>`, the same way.
- Deploy pending in post-merge: still `release --until`, now 5m. There is
  no condition for "deploy live".

The controller checks waits every subscriptions pass (1m) and after each
webhook drain, so a step resumes about a minute after CI finishes, not
10 to 15 minutes later. Count of fixed CI deferrals, before and after:

```sql
select strftime('%Y-%m-%dT%H', created_at) h, count(*) from comments
where body like 'Released until%' and body like '%CI%' group by h order by h desc limit 24;
```
