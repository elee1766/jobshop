# Checks waits on a commit (#831)

2026-11-06. `github.checks:repo=<repo>,sha=<sha>` (no `pr=`) reads the
commit's check runs and its Actions workflow runs
(`actions/runs?head_sha=`). A workflow run queued behind a concurrency
group has no check runs yet, so before this a master commit looked
green as soon as the release workflow finished.

- Cancelled is not a failure: on master it means a newer push
  superseded the run. `state=failure` fires only on a failed check.
- Once nothing is pending or failed and something was cancelled, the
  commit can never go green, so the wait is invalidated with
  `checks cancelled on <sha>`. The post-merge prompts then wait on the
  newest master commit instead.
- Skipped checks count for neither side. `state=success` needs at least
  one passed check.

PR waits (`pr=`) are unchanged: a cancelled check there still counts as
failed.
